package digitalocean

import (
	"bytes"
	"strconv"
	"text/template"
	"time"

	"github.com/digitalocean/godo"
	"github.com/supergiant/supergiant/bindata"
	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/model"
)

// CreateNode creates a new minion on DO kubernetes cluster.
func (p *Provider) CreateNode(m *model.Node, action *core.Action) error {
	// Build template
	minionUserdataTemplate, err := bindata.Asset("config/providers/digitalocean/minion.yaml")
	if err != nil {
		return err
	}
	minionTemplate, err := template.New("minion_template").Parse(string(minionUserdataTemplate))
	if err != nil {
		return err
	}

	data := struct {
		*model.Node
		Token string
	}{
		m,
		m.Kube.CloudAccount.Credentials["token"],
	}

	var minionUserdata bytes.Buffer
	if err = minionTemplate.Execute(&minionUserdata, data); err != nil {
		return err
	}

	dropletRequest := &godo.DropletCreateRequest{
		Name:              m.Kube.Name + "-minion",
		Region:            m.Kube.DigitalOceanConfig.Region,
		Size:              m.Size,
		PrivateNetworking: true,
		UserData:          string(minionUserdata.Bytes()),
		SSHKeys: []godo.DropletCreateSSHKey{
			{
				Fingerprint: m.Kube.DigitalOceanConfig.SSHKeyFingerprint,
			},
		},
		Image: godo.DropletCreateImage{
			Slug: "coreos-stable",
		},
	}
	tags := []string{"Kubernetes-Cluster", m.Kube.Name, dropletRequest.Name}

	minionDroplet, publicIP, err := p.createDroplet(p.Client(m.Kube), action, dropletRequest, tags)
	if err != nil {
		return err
	}

	// Parse creation timestamp
	createdAt, err := time.Parse("2006-01-02T15:04:05Z", minionDroplet.Created)
	if err != nil {
		// TODO need to return on error here
		p.Core.Log.Warnf("Could not parse Droplet creation timestamp string '%s': %s", minionDroplet.Created, err)
	}

	// Save info before waiting on IP
	m.ProviderID = strconv.Itoa(minionDroplet.ID)
	m.ProviderCreationTimestamp = createdAt
	m.ExternalIP = publicIP
	m.Name = publicIP

	return p.Core.DB.Save(m)
}
