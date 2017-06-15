package packet

import (
	"strings"

	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/model"
)

// DeleteNode deletes a minsion on a DO kubernetes cluster.
func (p *Provider) DeleteNode(m *model.Node, action *core.Action) error {
	// setup provider steps.
	procedure := &core.Procedure{
		Core:   p.Core,
		Name:   "Delete Node",
		Model:  m,
		Action: action,
	}

	// fetch client.
	client, err := p.Client(m.Kube)
	if err != nil {
		return err
	}

	procedure.AddStep("Destroying Kubernetes node...", func() error {
		_, err = client.Devices.Delete(m.ProviderID)
		if err != nil {
			if strings.Contains(err.Error(), "404 Not found") {
				// it does not exist,
				return nil
			}
			return err
		}
		return nil
	})
	return procedure.Run()
}
