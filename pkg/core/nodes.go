package core

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"regexp"
	"strconv"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/pkg/models"
)

type Nodes struct {
	Collection
}

func (c *Nodes) Create(m *models.Node) error {
	if err := c.Collection.Create(m); err != nil {
		return err
	}

	provision := &Action{
		Status: &models.ActionStatus{
			Description: "provisioning",

			// TODO
			// This resource has an issue with retryable provisioning -- which in this
			// context means creating an remote asset from the local record.
			//
			// Apps, for example, which use their user-set Name field as the actual
			// identifier for the provisioned Kubernetes Namespace. That makes the
			// creation of the Namespace retryable, because it is IDEMPOTENT.
			//
			// The problem here, is that WE CANNOT SET AN IDENTIFIER UP FRONT. The ID
			// is given to us upon successful creation of the remote asset.
			//
			// We currently do not have a great solution in place for this problem.
			// In the meantime, MaxRetries is set low to prevent creating several
			// duplicate, billable assets in the user's cloud account. If there is an
			// error, the user will know about it quickly, instead of after 20 retries.
			MaxRetries: 0,
		},
		core:  c.core,
		scope: c.core.DB.Preload("Kube.CloudAccount").Preload("Kube.Entrypoints.Kube.CloudAccount"),
		model: m,
		id:    m.ID,
		fn: func(_ *Action) error {
			server, err := c.createServer(m)
			if err != nil {
				return err
			}
			c.setAttrsFromServer(m, server)
			if err := c.core.DB.Save(m); err != nil {
				return err
			}
			for _, entrypoint := range m.Kube.Entrypoints {
				if err := c.core.Entrypoints.registerNodes(entrypoint, m); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return provision.Async()
}

func (c *Nodes) Delete(id *int64, m *models.Node) *Action {
	return &Action{
		Status: &models.ActionStatus{
			Description: "deleting",
			MaxRetries:  5,
		},
		core:  c.core,
		scope: c.core.DB.Preload("Kube.CloudAccount"),
		model: m,
		id:    id,
		fn: func(_ *Action) error {
			if m.ProviderID == "" {
				c.core.Log.Warnf("Deleting Node %d which has no provider_id", *m.ID)
			} else {
				if err := c.deleteServer(m); err != nil {
					return err
				}
			}
			return c.Collection.Delete(id, m)
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
// Private methods                                                            //
////////////////////////////////////////////////////////////////////////////////

func (c *Nodes) setAttrsFromServer(m *models.Node, server *ec2.Instance) {
	m.ProviderID = *server.InstanceId
	m.Name = *server.PrivateDnsName
	m.Class = *server.InstanceType
	m.ProviderCreationTimestamp = *server.LaunchTime
}

func (c *Nodes) createServer(m *models.Node) (*ec2.Instance, error) {

	// TODO move to init outside of func
	userdataTemplate, err := ioutil.ReadFile("config/minion_userdata.txt")
	if err != nil {
		return nil, err
	}
	template, err := template.New("minion_template").Parse(string(userdataTemplate))
	if err != nil {
		return nil, err
	}
	var userdata bytes.Buffer
	if err = template.Execute(&userdata, m.Kube); err != nil {
		return nil, err
	}
	encodedUserdata := base64.StdEncoding.EncodeToString(userdata.Bytes())

	input := &ec2.RunInstancesInput{
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		InstanceType: aws.String(m.Class),
		ImageId:      aws.String(AWSMasterAMIs[m.Kube.Config.Region]),
		EbsOptimized: aws.Bool(true),
		KeyName:      aws.String(m.Kube.Name + "-key"),
		SecurityGroupIds: []*string{
			aws.String(m.Kube.Config.NodeSecurityGroupID),
		},
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Name: aws.String("kubernetes-minion"),
		},
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			// root device ? TODO (do we need the one after this?)
			{
				DeviceName: aws.String("/dev/sda1"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeType:          aws.String("gp2"),
					VolumeSize:          aws.Int64(80),
					DeleteOnTermination: aws.Bool(true),
				},
			},
			&ec2.BlockDeviceMapping{
				DeviceName: aws.String("/dev/xvdb"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeType:          aws.String("gp2"),
					VolumeSize:          aws.Int64(80),
					DeleteOnTermination: aws.Bool(true),
				},
			},
		},
		UserData: aws.String(encodedUserdata),
		SubnetId: aws.String(m.Kube.Config.PublicSubnetID),
	}

	ec2S := c.core.CloudAccounts.ec2(m.Kube.CloudAccount, m.Kube.Config.Region)

	resp, err := ec2S.RunInstances(input)
	if err != nil {
		return nil, err
	}

	server := resp.Instances[0]

	err = tagAWSResource(ec2S, *server.InstanceId, map[string]string{
		"KubernetesCluster": m.Kube.Name,
		"Name":              m.Kube.Name + "-minion",
		"Role":              m.Kube.Name + "-minion",
	})
	if err != nil {
		// TODO
		c.core.Log.Error("Failed to tag EC2 Instance " + *server.InstanceId)
	}

	return server, nil
}

func (c *Nodes) deleteServer(m *models.Node) error {

	// TODO
	if m.Kube == nil {
		c.core.Log.Warnf("Deleting Node %d without deleting server because Kube is nil", *m.ID)
		return nil
	}

	input := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(m.ProviderID)},
	}
	_, err := c.core.CloudAccounts.ec2(m.Kube.CloudAccount, m.Kube.Config.Region).TerminateInstances(input)
	if isErrAndNotAWSNotFound(err) {
		return err
	}
	return nil
}

func (c *Nodes) hasPodsWithReservedResources(m *models.Node) (bool, error) {
	q := &guber.QueryParams{
		FieldSelector: "spec.nodeName=" + m.Name + ",status.phase=Running",
	}
	pods, err := c.core.K8S(m.Kube).Pods("").Query(q)
	if err != nil {
		return false, err
	}

	rxp := regexp.MustCompile("[0-9]+")

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			reqs := container.Resources.Requests

			if reqs == nil {
				continue
			}

			values := [2]string{
				reqs.CPU,
				reqs.Memory,
			}

			for _, val := range values {
				numstr := rxp.FindString(val)
				num := 0
				var err error
				if numstr != "" {
					num, err = strconv.Atoi(numstr)
					if err != nil {
						return false, err
					}
				}

				if num > 0 {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
