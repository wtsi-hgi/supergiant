package packet

import (
	"strings"

	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/model"
)

// DeleteKube deletes a GCE kubernetes cluster.
func (p *Provider) DeleteKube(m *model.Kube, action *core.Action) error {
	// setup provider steps.
	procedure := &core.Procedure{
		Core:   p.Core,
		Name:   "Delete Kube",
		Model:  m,
		Action: action,
	}
	// fetch client.
	client, err := p.Client(m)
	if err != nil {
		return err
	}
	// Delete all master nodes.
	procedure.AddStep("Destroying Kubernetes Master(s)...", func() error {
		for _, master := range m.PACKConfig.MasterNodes {

			deviceID, err := getDevice(m, client, master)
			if err != nil {
				return err
			}
			_, err = client.Devices.Delete(deviceID)
			if err != nil {
				if strings.Contains(err.Error(), "404 Not found") {
					// it does not exist,
					return nil
				}
				return err
			}
		}
		return nil
	})

	procedure.AddStep("Destroying Kubernetes Minions...", func() error {
		for _, node := range m.Nodes {

			deviceID, err := getDevice(m, client, node.Name)
			if err != nil {
				return err
			}

			_, err = client.Devices.Delete(deviceID)
			if err != nil {
				if strings.Contains(err.Error(), "404 Not found") {
					// it does not exist,
					return nil
				}
				return err
			}
		}
		return nil
	})

	// Initialize steps
	return procedure.Run()
}
