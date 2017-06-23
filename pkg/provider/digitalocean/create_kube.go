package digitalocean

import (
	"bytes"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/digitalocean/godo"
	"github.com/supergiant/supergiant/bindata"
	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/model"
	"github.com/supergiant/supergiant/pkg/util"
)

// CreateKube creates a new DO kubernetes cluster.
func (p *Provider) CreateKube(m *model.Kube, action *core.Action) error {

	if m.DigitalOceanConfig.SSHPubKey == "" {
		m.DigitalOceanConfig.SSHPubKey = "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAklOUpkDHrfHY17SbrmTIpNLTGK9Tjom/BWDSUGPl+nafzlHDTYW7hdI4yZ5ew18JH4JW9jbhUFrviQzM7xlELEVf4h9lFX5QVkbPppSwg0cda3Pbv7kOdJ/MTyBlWXFCR+HAo3FXRitBqxiX1nKhXpHAZsMciLq8V6RjsNAQwdsdMFvSlVK/7XAt3FaoJoAsncM1Q9x5+3V0Ww68/eIFmb1zuUFljQJKprrX88XypNDvjYNby6vw/Pb0rwert/EnmZ+AW4OZPnTPI89ZPmVMLuayrD2cE86Z/il8b+gw3r3+1nKatmIkjn2so1d01QraTlMqVSsbxNrRFi9wrf+M7Q== schacon@mylaptop.local"
	}

	procedure := &core.Procedure{
		Core:   p.Core,
		Name:   "Create Kube",
		Model:  m,
		Action: action,
	}

	client := p.Client(m)

	// Default master count to 1
	if m.DigitalOceanConfig.KubeMasterCount == 0 {
		m.DigitalOceanConfig.KubeMasterCount = 1
	}

	// provision an etcd token
	url, err := etcdToken(strconv.Itoa(m.DigitalOceanConfig.KubeMasterCount))
	if err != nil {
		return err
	}

	err = p.Core.DB.Save(m)
	if err != nil {
		return err
	}
	// save the token
	m.DigitalOceanConfig.ETCDDiscoveryURL = url

	procedure.AddStep("creating global tags for Kube", func() error {
		// These are created once, and then attached by name to created resource
		globalTags := []string{
			"Kubernetes-Cluster",
			m.Name,
			m.Name + "-master",
			m.Name + "-minion",
		}
		for _, tag := range globalTags {
			createInput := &godo.TagCreateRequest{
				Name: tag,
			}
			if _, _, err := client.Tags.Create(createInput); err != nil {
				// TODO
				p.Core.Log.Warnf("Failed to create Digital Ocean tag '%s': %s", tag, err)
			}
		}
		return nil
	})

	for i := 1; i <= m.DigitalOceanConfig.KubeMasterCount; i++ {
		// Create master(s)
		count := strconv.Itoa(i)

		procedure.AddStep("Creating Kubernetes Master Node "+count+"...", func() error {

			// Master name
			name := m.Name + "-master" + "-" + strings.ToLower(util.RandomString(5))

			m.DigitalOceanConfig.MasterName = name

			// Build template
			masterUserdataTemplate, err := bindata.Asset("config/providers/digitalocean/master.yaml")
			if err != nil {
				return err
			}
			masterTemplate, err := template.New("master_template").Parse(string(masterUserdataTemplate))
			if err != nil {
				return err
			}
			var masterUserdata bytes.Buffer
			if err = masterTemplate.Execute(&masterUserdata, m); err != nil {
				return err
			}

			dropletRequest := &godo.DropletCreateRequest{
				Name:              name,
				Region:            m.DigitalOceanConfig.Region,
				Size:              m.MasterNodeSize,
				PrivateNetworking: true,
				UserData:          string(masterUserdata.Bytes()),
				SSHKeys: []godo.DropletCreateSSHKey{
					{
						Fingerprint: m.DigitalOceanConfig.SSHKeyFingerprint,
					},
				},
				Image: godo.DropletCreateImage{
					Slug: "coreos-stable",
				},
			}
			tags := []string{"Kubernetes-Cluster", m.Name, name}

			masterDroplet, _, err := p.createDroplet(client, action, dropletRequest, tags)
			if err != nil {
				return err
			}

			m.DigitalOceanConfig.MasterID = masterDroplet.ID

			return action.CancellableWaitFor("Kubernetes master launch", 10*time.Minute, 3*time.Second, func() (bool, error) {
				resp, _, serr := client.Droplets.Get(masterDroplet.ID)
				if serr != nil {
					return false, serr
				}

				// Save Master info when ready
				if resp.Status == "active" {
					m.DigitalOceanConfig.MasterNodes = append(m.DigitalOceanConfig.MasterNodes, resp.ID)
					m.DigitalOceanConfig.MasterPrivateIP, _ = resp.PrivateIPv4()
					m.MasterPublicIP, _ = resp.PublicIPv4()
					if serr := p.Core.DB.Save(m); serr != nil {
						return false, serr
					}
				}
				return resp.Status == "active", nil
			})
		})
	}

	procedure.AddStep("building Kubernetes minion", func() error {
		// Load Nodes to see if we've already created a minion
		// TODO -- I think we can get rid of a lot of this do-unless behavior if we
		// modify Procedure to save progess on Action (which is easy to implement).
		if err := p.Core.DB.Find(&m.Nodes, "kube_name = ?", m.Name); err != nil {
			return err
		}
		if len(m.Nodes) > 0 {
			return nil
		}

		node := &model.Node{
			KubeName: m.Name,
			Kube:     m,
			Size:     m.NodeSizes[0],
		}
		return p.Core.Nodes.Create(node)
	})

	// TODO repeated in provider_aws.go
	procedure.AddStep("waiting for Kubernetes", func() error {
		return action.CancellableWaitFor("Kubernetes API and first minion", 20*time.Minute, 3*time.Second, func() (bool, error) {
			k8s := p.Core.K8S(m)
			k8sNodes, err := k8s.ListNodes("")
			if err != nil {
				return false, nil
			}
			return len(k8sNodes) > 0, nil
		})
	})

	return procedure.Run()
}
