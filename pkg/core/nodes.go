package core

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/pkg/models"
	"github.com/supergiant/supergiant/pkg/util"
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
		scope: c.core.DB.Preload("Kube.CloudAccount"),
		model: m,
		id:    m.ID,
		fn: func(_ *Action) error {
			server, err := c.createServer(m)
			if err != nil {
				return err
			}
			c.setAttrsFromServer(m, server)
			return c.core.DB.Save(m)
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
	group, err := c.core.Kubes.autoScalingGroup(m.Kube, m.Class)
	if err != nil {
		return nil, err
	}

	// TODO race conditions can occur between fetching / updating the size fields.
	// We need a mutex of some sort.

	origInstanceIDs := make(map[string]struct{})
	for _, instance := range group.Instances {
		origInstanceIDs[*instance.InstanceId] = struct{}{}
	}

	startingCapacity := *group.DesiredCapacity
	desiredCapacity := startingCapacity + 1

	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: group.AutoScalingGroupName,
		DesiredCapacity:      aws.Int64(desiredCapacity),
		MaxSize:              aws.Int64(desiredCapacity),
	}
	_, err = c.core.CloudAccounts.autoscaling(m.Kube.CloudAccount, m.Kube.Config.Region).UpdateAutoScalingGroup(input)
	if err != nil {
		return nil, err
	}

	var serverID string

	desc := fmt.Sprintf(
		"autoscaling group %s scale up from %d to %d instances",
		*group.AutoScalingGroupName,
		startingCapacity,
		desiredCapacity,
	)
	err = util.WaitFor(desc, 60*time.Second, 2*time.Second, func() (bool, error) {
		group, err := c.core.Kubes.autoScalingGroup(m.Kube, m.Class)
		if err != nil {
			return false, err
		}

		c.core.Log.Debugf("Waiting for %s", desc)

		for _, instance := range group.Instances {
			if _, ok := origInstanceIDs[*instance.InstanceId]; !ok { // if this is a new instance ID
				serverID = *instance.InstanceId
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return c.core.Kubes.server(m.Kube, serverID)
}

func (c *Nodes) deleteServer(m *models.Node) error {

	// TODO
	if m.Kube == nil {
		c.core.Log.Warnf("Deleting Node %d without deleting server because Kube is nil", *m.ID)
		return nil
	}

	input := &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		InstanceId:                     aws.String(m.ProviderID),
		ShouldDecrementDesiredCapacity: aws.Bool(true),
	}
	_, err := c.core.CloudAccounts.autoscaling(m.Kube.CloudAccount, m.Kube.Config.Region).TerminateInstanceInAutoScalingGroup(input)
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
