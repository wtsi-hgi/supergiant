package core

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/supergiant/supergiant/pkg/models"
)

type Entrypoints struct {
	Collection
}

func (c *Entrypoints) Create(m *models.Entrypoint) error {
	if err := c.Collection.Create(m); err != nil {
		return err
	}

	// Load Kube and CloudAccount
	if err := c.core.DB.Preload("Nodes").Preload("CloudAccount").First(m.Kube, m.KubeID); err != nil {
		return err
	}

	provision := &Action{
		Status: &models.ActionStatus{
			Description: "provisioning",
			MaxRetries:  5,
		},
		core:       c.core,
		resourceID: m.UUID,
		model:      m,
		fn: func(_ *Action) error {
			return c.createELB(m)
		},
	}
	return provision.Async()
}

func (c *Entrypoints) Delete(id *int64, m *models.Entrypoint) *Action {
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
			if err := c.deleteELB(m); err != nil {
				return err
			}
			return c.Collection.Delete(id, m)
		},
	}
}

func (c *Entrypoints) SetPort(id *int64, m *models.Entrypoint, elbPort int, instancePort int) error {
	action := &Action{
		Status: &models.ActionStatus{
			Description: fmt.Sprintf("setting port %d:%d", elbPort, instancePort),
			MaxRetries:  5,
		},
		core:       c.core,
		scope:      c.core.DB.Preload("Kube.CloudAccount"),
		model:      m,
		id:         id,
		resourceID: m.UUID,
		fn: func(_ *Action) error {
			params := &elb.CreateLoadBalancerListenersInput{
				LoadBalancerName: aws.String(m.ProviderID),
				Listeners: []*elb.Listener{
					{
						InstancePort:     aws.Int64(int64(instancePort)),
						LoadBalancerPort: aws.Int64(int64(elbPort)),
						Protocol:         aws.String("TCP"),
					},
				},
			}
			_, err := c.core.CloudAccounts.elb(m.Kube.CloudAccount, m.Kube.Config.Region).CreateLoadBalancerListeners(params)
			return err
		},
	}
	return action.Now()
}

func (c *Entrypoints) RemovePort(id *int64, m *models.Entrypoint, elbPort int) error {
	action := &Action{
		Status: &models.ActionStatus{
			Description: fmt.Sprintf("removing port %d", elbPort),
			MaxRetries:  5,
		},
		core:       c.core,
		scope:      c.core.DB.Preload("Kube.CloudAccount"),
		model:      m,
		id:         id,
		resourceID: m.UUID,
		fn: func(_ *Action) error {
			params := &elb.DeleteLoadBalancerListenersInput{
				LoadBalancerName: aws.String(m.ProviderID),
				LoadBalancerPorts: []*int64{
					aws.Int64(int64(elbPort)),
				},
			}
			_, err := c.core.CloudAccounts.elb(m.Kube.CloudAccount, m.Kube.Config.Region).DeleteLoadBalancerListeners(params)
			if isErrAndNotAWSNotFound(err) {
				return err
			}
			return nil
		},
	}
	return action.Now()
}

////////////////////////////////////////////////////////////////////////////////
// Private methods                                                            //
////////////////////////////////////////////////////////////////////////////////

func (c *Entrypoints) createELB(m *models.Entrypoint) error {
	params := &elb.CreateLoadBalancerInput{
		Listeners: []*elb.Listener{ // NOTE we must provide at least 1 listener, it is currently arbitrary
			{
				InstancePort:     aws.Int64(420),
				LoadBalancerPort: aws.Int64(420),
				Protocol:         aws.String("TCP"),
			},
		},
		LoadBalancerName: aws.String(m.ProviderID),
		Scheme:           aws.String("internet-facing"),
		SecurityGroups: []*string{
			aws.String(m.Kube.Config.ELBSecurityGroupID),
		},
		Subnets: []*string{
			aws.String(m.Kube.Config.PublicSubnetID),
		},
	}
	resp, err := c.core.CloudAccounts.elb(m.Kube.CloudAccount, m.Kube.Config.Region).CreateLoadBalancer(params)
	if err != nil {
		return err
	}

	// Save Address
	m.Address = *resp.DNSName
	if err := c.core.DB.Save(m); err != nil {
		return err
	}

	if err := c.registerNodes(m, m.Kube.Nodes...); err != nil {
		return err
	}

	// Configure health check
	healthParams := &elb.ConfigureHealthCheckInput{
		LoadBalancerName: aws.String(m.ProviderID),
		HealthCheck: &elb.HealthCheck{
			Target:             aws.String("HTTPS:10250/healthz"),
			HealthyThreshold:   aws.Int64(2),
			UnhealthyThreshold: aws.Int64(10),
			Interval:           aws.Int64(30),
			Timeout:            aws.Int64(5),
		},
	}
	_, err = c.core.CloudAccounts.elb(m.Kube.CloudAccount, m.Kube.Config.Region).ConfigureHealthCheck(healthParams)
	return err
}

func (c *Entrypoints) registerNodes(m *models.Entrypoint, nodes ...*models.Node) error {
	var elbInstances []*elb.Instance
	for _, node := range nodes {
		elbInstances = append(elbInstances, &elb.Instance{
			InstanceId: aws.String(node.ProviderID),
		})
	}
	input := &elb.RegisterInstancesWithLoadBalancerInput{
		LoadBalancerName: aws.String(m.ProviderID),
		Instances:        elbInstances,
	}
	_, err := c.core.CloudAccounts.elb(m.Kube.CloudAccount, m.Kube.Config.Region).RegisterInstancesWithLoadBalancer(input)
	return err
}

func (c *Entrypoints) deleteELB(m *models.Entrypoint) error {
	// Delete ELB
	params := &elb.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(m.ProviderID),
	}
	_, err := c.core.CloudAccounts.elb(m.Kube.CloudAccount, m.Kube.Config.Region).DeleteLoadBalancer(params)
	if isErrAndNotAWSNotFound(err) {
		return err
	}
	return nil
}
