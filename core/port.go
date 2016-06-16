package core

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

func protoWithDefault(protocol string) string {
	if protocol == "" {
		return "tcp"
	}
	return strings.ToLower(protocol)
}

func findPortsUniqueToSetA(setA []*Port, setB []*Port) (ports []*Port) {
	for _, pA := range setA {
		unique := true
		for _, pB := range setB {
			if reflect.DeepEqual(*pA.Port, *pB.Port) {
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
	*common.Port
	service *guber.Service
	// entrypoint is nil if it's an internal port
	entrypoint *EntrypointResource
}

func (p *Port) name() string {
	return strconv.Itoa(p.Number)
}

func newInternalPort(c *Core, p *common.Port, svc *guber.Service) *Port {
	return &Port{c, p, svc, nil}
}

func newExternalPort(c *Core, p *common.Port, svc *guber.Service, e *EntrypointResource) *Port {
	return &Port{c, p, svc, e}
}

func (p *Port) internalAddress() *common.PortAddress {
	svcMeta := p.service.Metadata
	host := fmt.Sprintf("%s.%s.svc.cluster.local", svcMeta.Name, svcMeta.Namespace)
	return &common.PortAddress{
		Port:    p.name(),
		Address: fmt.Sprintf("%s://%s:%d", protoWithDefault(p.Protocol), host, p.Number),
	}
}

func (p *Port) externalAddress() *common.PortAddress {
	if p.entrypoint == nil {
		host := ""
		nodes, err := p.core.Nodes().List()
		if err != nil {
			Log.Errorf("Error when fetching nodes for external address IP: %s", err)
		} else if len(nodes.Items) == 0 {
			Log.Error("Error no nodes present when building external address")
		} else {
			host = nodes.Items[0].ExternalIP
		}
		return &common.PortAddress{
			Port:    p.name(),
			Address: fmt.Sprintf("%s://%s:%d", protoWithDefault(p.Protocol), host, p.nodePort()),
		}
	}

	return &common.PortAddress{
		Port:    p.name(),
		Address: fmt.Sprintf("%s://%s:%d", protoWithDefault(p.Protocol), p.entrypoint.Address, p.elbPort()),
	}
}

// The following methods apply to external ports only

func (p *Port) nodePort() int {
	for _, port := range p.service.Spec.Ports {
		if port.Port == p.Number {
			return port.NodePort
		}
	}
	panic("Could not find NodePort")
}

func (p *Port) elbPort() int {
	if !p.PerInstance && p.ExternalNumber != 0 {
		return p.ExternalNumber
	}
	return p.nodePort()
}

// TODO like the comment above, this only applies when there is an EntrypointDomain
func (p *Port) addToELB() error {
	return p.entrypoint.AddPort(p.elbPort(), p.nodePort())
}

func (p *Port) removeFromELB() error {
	return p.entrypoint.RemovePort(p.elbPort())
}
