package core

import (
	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

type ServiceSet struct {
	core          *Core
	release       *ReleaseResource
	namespace     string
	baseName      string
	labelSelector map[string]string
	portFilter    func(*common.Port) bool

	previous *ServiceSet

	internal *guber.Service
	external *guber.Service
}

func (s *ServiceSet) internalServiceName() string {
	return s.baseName
}

func (s *ServiceSet) externalServiceName() string {
	return s.baseName + "-public"
}

func (s *ServiceSet) internalService() (*guber.Service, error) {
	if s.internal == nil {
		svc, err := s.getService(s.internalServiceName())
		if err != nil {
			return nil, err
		}
		s.internal = svc
	}
	return s.internal, nil
}

func (s *ServiceSet) externalService() (*guber.Service, error) {
	if s.external == nil {
		svc, err := s.getService(s.externalServiceName())
		if err != nil {
			return nil, err
		}
		s.external = svc
	}
	return s.external, nil
}

func (s *ServiceSet) selectPortDefs(portFilterArg func(*common.Port) bool) (ports []*common.Port) {
	for _, container := range s.release.Containers {
		for _, port := range container.Ports {
			if (s.portFilter == nil || s.portFilter(port)) && portFilterArg(port) {
				ports = append(ports, port)
			}
		}
	}
	return
}

func (s *ServiceSet) internalPortDefs() []*common.Port {
	return s.selectPortDefs(func(port *common.Port) bool {
		return !port.Public
	})
}

func (s *ServiceSet) externalPortDefs() []*common.Port {
	return s.selectPortDefs(func(port *common.Port) bool {
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
		ports = append(ports, newInternalPort(s.core, port, svc))
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
		entrypoint, ok := s.release.entrypoints[*port.EntrypointDomain]
		if !ok {
			Log.Errorf("Entrypoint %s does not exist", *port.EntrypointDomain)
			continue
		}
		ports = append(ports, newExternalPort(s.core, port, svc, entrypoint))
	}
	return
}

func (s *ServiceSet) provisionServices() error {
	internal, err := s.provisionService(s.internalServiceName(), "ClusterIP", asKubeServicePorts(s.internalPortDefs()))
	if err != nil {
		return err
	}
	s.internal = internal

	external, err := s.provisionService(s.externalServiceName(), "NodePort", asKubeServicePorts(s.externalPortDefs()))
	if err != nil {
		return err
	}
	s.external = external

	return nil
}

func (s *ServiceSet) deleteServices() error {
	if err := s.deleteService(s.internalServiceName()); err != nil {
		return err
	}
	if err := s.deleteService(s.externalServiceName()); err != nil {
		return err
	}
	return nil
}

func (s *ServiceSet) provision() error {
	if err := s.provisionServices(); err != nil {
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
	if err := s.deleteServices(); err != nil {
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
		addPortsToService(s.internal, newInternalPorts)
	}

	if len(newExternalPorts) > 0 {
		addPortsToService(s.external, newExternalPorts)

		for _, port := range newExternalPorts {
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
		removePortsFromService(s.internal, oldInternalPorts)
	}

	if len(oldExternalPorts) > 0 {
		removePortsFromService(s.external, oldExternalPorts)

		for _, port := range oldExternalPorts {
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
		if err := port.removeFromELB(); err != nil {
			return err
		}
	}
	return nil
}

func (s *ServiceSet) getService(name string) (svc *guber.Service, err error) {
	svc, err = s.core.k8s.Services(s.namespace).Get(name)
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

	Log.Infof("Creating Service %s", name)
	return s.core.k8s.Services(s.namespace).Create(svc)
}

func (s *ServiceSet) deleteService(name string) error {
	Log.Infof("Deleting Service %s", name)
	if err := s.core.k8s.Services(s.namespace).Delete(name); err != nil && !isKubeNotFoundErr(err) {
		return err
	}
	return nil
}

//------------------------------------------ move below to kube helpers

func asKubeServicePorts(inPorts []*common.Port) (outPorts []*guber.ServicePort) {
	for _, port := range inPorts {
		outPorts = append(outPorts, asKubeServicePort(port))
	}
	return
}

func addPortsToService(svc *guber.Service, ports []*Port) error {
	Log.Infof("Adding new ports to Service %s", svc.Metadata.Name)
	for _, port := range ports {
		svc.Spec.Ports = append(svc.Spec.Ports, asKubeServicePort(port.Port))
	}
	return svc.Save()
}

func removePortsFromService(svc *guber.Service, ports []*Port) error {
	Log.Infof("Removing old ports from Service %s", svc.Metadata.Name)
	for _, port := range ports {
		for i, svcPort := range svc.Spec.Ports {
			if svcPort.Port == port.Number {
				svc.Spec.Ports = append(svc.Spec.Ports[:i], svc.Spec.Ports[i+1:]...)
			}
		}
	}
	return svc.Save()
}
