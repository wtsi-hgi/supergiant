package digitalocean

import (
	"strconv"
	"strings"

	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/model"
)

func (p *Provider) DeleteKube(m *model.Kube, action *core.Action) error {
	// New Client
	client := p.Client(m)
	// Step procedure
	procedure := &core.Procedure{
		Core:   p.Core,
		Name:   "Delete Kube",
		Model:  m,
		Action: action,
	}

	procedure.AddStep("deleting master", func() error {
		if m.MasterID == "" {
			return nil
		}
		for _, master := range m.MasterNodes {
			imaster, err := strconv.Atoi(master)
			if err != nil {
				return err
			}
			if _, err := client.Droplets.Delete(imaster); err != nil && !strings.Contains(err.Error(), "404") {
				return err
			}
		}

		m.MasterID = ""
		return nil
	})

	return procedure.Run()
}
