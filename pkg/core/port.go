package core

import (
	"fmt"
	"strconv"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/pkg/model"
)

func findPortsUniqueToSetA(setA []*Port, setB []*Port) (ports []*Port) {
	for _, pA := range setA {
		unique := true
		for _, pB := range setB {
			if pA.Port.Number == pB.Port.Number && pA.Port.ExternalNumber == pB.Port.ExternalNumber && pA.Port.Protocol == pB.Port.Protocol {
				unique = false
				break
			}
		}
		if unique {
			ports = append(ports, pA)
		}
	}
	return
}

type Port struct {
	core *Core
	*model.Port
	service *guber.Service
	// entrypoint is nil if it's an internal port
	entrypoint *model.Entrypoint
}

func (p *Port) name() string {
	return strconv.Itoa(p.Number)
}

func (p *Port) internalAddress() *model.PortAddress {
	svcMeta := p.service.Metadata
	host := fmt.Sprintf("%s.%s.svc.cluster.local", svcMeta.Name, svcMeta.Namespace)
	return &model.PortAddress{
		Port: p.name(),
		// Address: fmt.Sprintf("%s://%s:%d", protoWithDefault(p.Protocol), host, p.Number),
		Address: fmt.Sprintf("%s:%d", host, p.Number),
	}
}

func (p *Port) externalAddress() *model.PortAddress {
	if p.entrypoint == nil {
		host := ""
		node := new(model.Node)
		if err := p.core.DB.First(node); err != nil {
			p.core.Log.Errorf("Error when fetching nodes for external address IP: %s", err)
		} else {
			host = node.ExternalIP
		}
		return &model.PortAddress{
			Port: p.name(),
			// Address: fmt.Sprintf("%s://%s:%d", protoWithDefault(p.Protocol), host, p.nodePort()),
			Address: fmt.Sprintf("%s:%d", host, p.nodePort()),
		}
	}

	return &model.PortAddress{
		Port: p.name(),
		// Address: fmt.Sprintf("%s://%s:%d", protoWithDefault(p.Protocol), p.entrypoint.Address, p.elbPort()),
		Address: fmt.Sprintf("%s:%d", p.entrypoint.Address, p.elbPort()),
	}
}

// The following methods apply to external ports only

func (p *Port) nodePort() int64 {
	for _, port := range p.service.Spec.Ports {
		if port.Port == p.Number {
			return int64(port.NodePort)
		}
	}
	panic(fmt.Sprintf("Could not find NodePort for %#v", *p.Port))
}

func (p *Port) elbPort() int64 {
	if !p.PerInstance && p.ExternalNumber != 0 {
		return int64(p.ExternalNumber)
	}
	return int64(p.nodePort())
}

// TODO like the comment above, this only applies when there is an EntrypointDomain
func (p *Port) addToELB() error {
	return p.core.Entrypoints.SetPort(p.entrypoint.ID, p.entrypoint, p.elbPort(), p.nodePort())
}

func (p *Port) removeFromELB() error {
	return p.core.Entrypoints.RemovePort(p.entrypoint.ID, p.entrypoint, p.elbPort())
}
