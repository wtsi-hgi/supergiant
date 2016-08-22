package core

import (
	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/pkg/models"
)

type ServiceSet struct {
	core          *Core
	component     *models.Component
	release       *models.Release
	labelSelector map[string]string
	portFilter    func(*models.Port) bool

	internalServiceName string
	externalServiceName string
	namespace           string

	previous    *ServiceSet
	entrypoints map[int64]*models.Entrypoint

	internal *guber.Service
	external *guber.Service
}

func NewServiceSet(core *Core, component *models.Component, release *models.Release, baseName string, labelSelector map[string]string, portFilter func(*models.Port) bool) (*ServiceSet, error) {
	if release == nil {
		release = component.CurrentRelease
	}

	serviceSet := &ServiceSet{
		core:                core,
		component:           component,
		release:             release,
		labelSelector:       labelSelector,
		portFilter:          portFilter,
		internalServiceName: baseName,
		externalServiceName: baseName + "-public",
		namespace:           component.App.Name,
	}

	// Load old ServiceSet (config)
	if component.CurrentRelease != nil && component.TargetReleaseID != nil && *release.ID == *component.TargetReleaseID {
		previous, err := NewServiceSet(core, component, component.CurrentRelease, baseName, labelSelector, portFilter)
		if err != nil {
			return nil, err
		}
		serviceSet.previous = previous
	}

	// Load Entrypoints
	serviceSet.entrypoints = make(map[int64]*models.Entrypoint)
	for _, port := range serviceSet.externalPortDefs() {
		if port.EntrypointID == nil {
			continue
		}
		entrypoint := new(models.Entrypoint)
		if err := core.Entrypoints.GetWithIncludes(port.EntrypointID, entrypoint, []string{"Kube.CloudAccount"}); err != nil {
			// NOTE if return error here delete will fail if Entrypoint has been deleted
			core.Log.Warn(err)
			continue
		}
		serviceSet.entrypoints[*port.EntrypointID] = entrypoint
	}

	return serviceSet, nil
}

func (s *ServiceSet) internalService() (*guber.Service, error) {
	if s.internal == nil {
		svc, err := s.getService(s.internalServiceName)
		if err != nil {
			return nil, err
		}
		s.internal = svc
	}
	return s.internal, nil
}

func (s *ServiceSet) externalService() (*guber.Service, error) {
	if s.external == nil {
		svc, err := s.getService(s.externalServiceName)
		if err != nil {
			return nil, err
		}
		s.external = svc
	}
	return s.external, nil
}

func (s *ServiceSet) selectPortDefs(portFilterArg func(*models.Port) bool) (ports []*models.Port) {
	for _, container := range s.release.Config.Containers {
		for _, port := range container.Ports {
			if (s.portFilter == nil || s.portFilter(port)) && portFilterArg(port) {
				ports = append(ports, port)
			}
		}
	}
	return
}

func (s *ServiceSet) internalPortDefs() []*models.Port {
	return s.selectPortDefs(func(port *models.Port) bool {
		return !port.Public
	})
}

func (s *ServiceSet) externalPortDefs() []*models.Port {
	return s.selectPortDefs(func(port *models.Port) bool {
		return port.Public
	})
}

func (s *ServiceSet) internalPorts() (ports []*Port, err error) {
	svc, err := s.internalService()
	if err != nil {
		return nil, err
	} else if svc == nil {
		return // Service does not exist (which... can be an error depending on the context called)
	}
	for _, port := range s.internalPortDefs() {
		ports = append(ports, &Port{s.core, port, svc, nil})
	}
	return
}

func (s *ServiceSet) externalPorts() (ports []*Port, err error) {
	svc, err := s.externalService()
	if err != nil {
		return nil, err
	} else if svc == nil {
		return // Service does not exist (which... can be an error depending on the context called)
	}
	for _, port := range s.externalPortDefs() {
		var entrypoint *models.Entrypoint
		if port.EntrypointID != nil {
			ep, ok := s.entrypoints[*port.EntrypointID]
			if !ok {
				s.core.Log.Errorf("Entrypoint %d does not exist", port.EntrypointID)
				continue
			}
			entrypoint = ep
		}
		ports = append(ports, &Port{s.core, port, svc, entrypoint})
	}
	return
}

func (s *ServiceSet) provision() (err error) {
	s.internal, err = s.provisionService(s.internalServiceName, "ClusterIP", asKubeServicePorts(s.internalPortDefs()))
	if err != nil {
		return err
	}
	s.external, err = s.provisionService(s.externalServiceName, "NodePort", asKubeServicePorts(s.externalPortDefs()))
	if err != nil {
		return err
	}
	if err := s.addExternalPortsToEntrypoint(); err != nil {
		return err
	}
	return nil
}

func (s *ServiceSet) delete() error {
	if err := s.removeExternalPortsFromEntrypoint(); err != nil {
		return err
	}
	if err := s.deleteService(s.internalServiceName); err != nil {
		return err
	}
	if err := s.deleteService(s.externalServiceName); err != nil {
		return err
	}
	return nil
}

func (s *ServiceSet) allOldAndNewPorts() (ni []*Port, ne []*Port, oi []*Port, oe []*Port, err error) {
	if ni, err = s.internalPorts(); err != nil {
		return
	}
	if ne, err = s.externalPorts(); err != nil {
		return
	}
	if s.previous != nil {
		if oi, err = s.previous.internalPorts(); err != nil {
			return
		}
		if oe, err = s.previous.externalPorts(); err != nil {
			return
		}
	}
	return
}

// AddNewPorts adds any new ports defined in containers to the existing
// Services. This is used as a part of the deployment process, and is used in
// conjunction with RemoveOldPorts.
// We use the config returned from the services themselves, as opposed to just
// updating the config, because auto-assigned ports need to be preserved.
func (s *ServiceSet) addNewPorts() error {
	ni, ne, oi, oe, err := s.allOldAndNewPorts()
	if err != nil {
		return err
	} else if len(oi) == 0 && len(oe) == 0 { // No old serviceSet (no current release)
		return nil
	}
	newInternalPorts := findPortsUniqueToSetA(ni, oi)
	newExternalPorts := findPortsUniqueToSetA(ne, oe)
	if len(newInternalPorts) > 0 {
		addPortsToService(s.core, s.internal, newInternalPorts)
	}

	if len(newExternalPorts) > 0 {
		addPortsToService(s.core, s.external, newExternalPorts)

		for _, port := range newExternalPorts {
			if port.EntrypointID == nil {
				continue
			}
			if err := port.addToELB(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *ServiceSet) removeOldPorts() error {
	ni, ne, oi, oe, err := s.allOldAndNewPorts()
	if err != nil {
		return err
	} else if len(oi) == 0 && len(oe) == 0 { // No old serviceSet (no current release)
		return nil
	}
	oldInternalPorts := findPortsUniqueToSetA(oi, ni)
	oldExternalPorts := findPortsUniqueToSetA(oe, ne)

	if len(oldInternalPorts) > 0 {
		removePortsFromService(s.core, s.internal, oldInternalPorts)
	}

	if len(oldExternalPorts) > 0 {
		removePortsFromService(s.core, s.external, oldExternalPorts)

		for _, port := range oldExternalPorts {
			if port.EntrypointID == nil {
				continue
			}
			if err := port.removeFromELB(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *ServiceSet) addExternalPortsToEntrypoint() error {
	svc, err := s.externalService()
	if err != nil {
		return err
	} else if svc == nil {
		return nil
	}

	ports, err := s.externalPorts()
	if err != nil {
		return err
	}

	// NOTE we find from the service so that we don't try to add not-yet serviced
	// ports to the ELB

	for _, svcPort := range svc.Spec.Ports {
		for _, port := range ports {
			if port.Number == svcPort.Port {
				if port.EntrypointID == nil {
					continue
				}
				if err := port.addToELB(); err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

func (s *ServiceSet) removeExternalPortsFromEntrypoint() error {
	ports, err := s.externalPorts()
	if err != nil {
		return err
	}
	for _, port := range ports {
		if port.EntrypointID == nil {
			continue
		}
		if err := port.removeFromELB(); err != nil {
			return err
		}
	}
	return nil
}

func (s *ServiceSet) getService(name string) (svc *guber.Service, err error) {
	svc, err = s.core.K8S(s.component.App.Kube).Services(s.namespace).Get(name)
	if err != nil && isKubeNotFoundErr(err) {
		err = nil
	}
	return
}

func (s *ServiceSet) provisionService(name string, svcType string, ports []*guber.ServicePort) (*guber.Service, error) {
	if len(ports) == 0 {
		return nil, nil
	}

	svc, err := s.getService(name)
	if err != nil {
		return nil, err
	} else if svc != nil {
		return svc, nil
	}

	svc = &guber.Service{
		Metadata: &guber.Metadata{
			Name: name,
		},
		Spec: &guber.ServiceSpec{
			Type:     svcType,
			Selector: s.labelSelector,
			Ports:    ports,
		},
	}

	s.core.Log.Infof("Creating Service %s", name)
	return s.core.K8S(s.component.App.Kube).Services(s.namespace).Create(svc)
}

func (s *ServiceSet) deleteService(name string) error {
	s.core.Log.Infof("Deleting Service %s", name)
	if err := s.core.K8S(s.component.App.Kube).Services(s.namespace).Delete(name); err != nil && !isKubeNotFoundErr(err) {
		return err
	}
	return nil
}

func (s *ServiceSet) externalAddresses() (addrs []*models.PortAddress, err error) {
	ports, err := s.externalPorts()
	if err != nil {
		return nil, err
	}
	for _, port := range ports {
		addrs = append(addrs, port.externalAddress())
	}
	return addrs, nil
}

func (s *ServiceSet) internalAddresses() (addrs []*models.PortAddress, err error) {
	iPorts, err := s.internalPorts()
	if err != nil {
		return nil, err
	}
	ePorts, err := s.externalPorts() // external ports also have internal addresses
	if err != nil {
		return nil, err
	}
	ports := append(iPorts, ePorts...)
	for _, port := range ports {
		addrs = append(addrs, port.internalAddress())
	}

	return addrs, nil
}

//------------------------------------------ move below to kube helpers

func asKubeServicePorts(inPorts []*models.Port) (outPorts []*guber.ServicePort) {
	for _, port := range inPorts {
		outPorts = append(outPorts, asKubeServicePort(port))
	}
	return
}

func addPortsToService(c *Core, svc *guber.Service, ports []*Port) error {
	c.Log.Infof("Adding new ports to Service %s", svc.Metadata.Name)
	for _, port := range ports {
		svc.Spec.Ports = append(svc.Spec.Ports, asKubeServicePort(port.Port))
	}
	return svc.Save()
}

func removePortsFromService(c *Core, svc *guber.Service, ports []*Port) error {
	c.Log.Infof("Removing old ports from Service %s", svc.Metadata.Name)
	for _, port := range ports {
		for i, svcPort := range svc.Spec.Ports {
			if svcPort.Port == port.Number {
				svc.Spec.Ports = append(svc.Spec.Ports[:i], svc.Spec.Ports[i+1:]...)
			}
		}
	}
	return svc.Save()
}
