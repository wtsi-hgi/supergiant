package digitalocean

import (
	"strconv"

	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/model"
)

// DeleteNode deletes a minsion on a DO kubernetes cluster.
func (p *Provider) DeleteNode(m *model.Node, action *core.Action) error {
	client := p.Client(m.Kube)

	intID, err := strconv.Atoi(m.ProviderID)
	if err != nil {
		return err
	}
	_, err = client.Droplets.Delete(intID)
	return err
}
