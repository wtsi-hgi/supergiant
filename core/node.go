package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

type NodesInterface interface {
	// Nodes are unique as a Resource because the collection in etcd is not
	// considered the "source of truth". We have to expect that users could
	// decide to manually increment the number on the autoscaling groups in the
	// AWS dashboard -- and if we want Helium to be able to delete Nodes through
	// our typical CRUD interface, a record must be created for that rogue node.
	//
	// Thus, populate() will gather all the relevant server IDs from AWS, and
	// create Nodes for those that do not exist. If there is a record that does
	// not have a matching node, it is deleted.
	populate() error

	List() (*NodeList, error)
	New() *NodeResource
	Create(*NodeResource) error
	Get(common.ID) (*NodeResource, error)
	Update(common.ID, *NodeResource) error
	Patch(common.ID, *NodeResource) error
	Delete(*NodeResource) error
}

type NodeCollection struct {
	core *Core
}

type NodeResource struct {
	core       *Core
	collection NodesInterface
	*common.Node
}

// NOTE this does not inherit from common like model does; all we need is a List
// object, internally, that has a slice of our composed model above.
type NodeList struct {
	Items []*NodeResource `json:"items"`
}

// initializeResource implements the Collection interface.
func (c *NodeCollection) initializeResource(in Resource) {
	r := in.(*NodeResource)
	r.collection = c
	r.core = c.core
}

// List returns an NodeList.
func (c *NodeCollection) List() (*NodeList, error) {
	list := new(NodeList)
	err := c.core.db.list(c, list)
	return list, err
}

// New initializes an Node with a pointer to the Collection.
func (c *NodeCollection) New() *NodeResource {
	r := &NodeResource{
		Node: &common.Node{
			Meta: common.NewMeta(),
		},
	}
	c.initializeResource(r)
	return r
}

// Create takes an Node and creates it in etcd.
func (c *NodeCollection) Create(r *NodeResource) error {

	// TODO
	//
	// This refreshes the entire db on every create. It should ideally poll
	// in the background.
	//
	if err := c.populate(); err != nil {
		return err
	}

	server, err := c.createServer(r.Class) // TODO move to AWS helpers
	if err != nil {
		return err
	}
	node, err := c.createNodeFromServer(server)
	if err != nil {
		return err
	}

	// NOTE we do this because we do not directly set attributes on r, we actually
	// create a new Node record... which should maybe be changed.
	*r = *node

	return nil
}

// Get takes a name and returns an NodeResource if it exists.
func (c *NodeCollection) Get(id common.ID) (*NodeResource, error) {
	r := c.New()
	if err := c.core.db.get(c, id, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Update updates the Node in etcd.
func (c *NodeCollection) Update(id common.ID, r *NodeResource) error {
	return c.core.db.update(c, id, r)
}

// Patch partially updates the App in etcd.
func (c *NodeCollection) Patch(name common.ID, r *NodeResource) error {
	return c.core.db.patch(c, name, r)
}

// Delete deletes the Node in etcd.
func (c *NodeCollection) Delete(r *NodeResource) error {
	if err := r.deleteServer(); err != nil {
		return err
	}
	return c.core.db.delete(c, r.ID)
}

func (c *NodeCollection) populate() error {
	nodes, err := c.List()
	if err != nil {
		return err
	}

	servers, err := ec2InstancesFromAutoscalingGroups(c.core)
	if err != nil {
		return err
	}

	existentNodeIDs := make(map[string]struct{})

	// Create new Nodes from newly discovered servers
	for _, server := range servers {
		exists := false
		for _, node := range nodes.Items {
			if *node.ID == *server.InstanceId {
				existentNodeIDs[*node.ID] = struct{}{}
				exists = true
				break
			}
		}
		if !exists {
			if _, err := c.createNodeFromServer(server); err != nil {
				return err
			}
		}
	}

	// Delete any Nodes which no longer exist
	for _, node := range nodes.Items {
		if _, exists := existentNodeIDs[*node.ID]; exists {
			continue
		}

		Log.Warnf("Deleting node with ID %s because it no longer exists in AWS", *node.ID)

		if err := node.Delete(); err != nil {
			return err
		}
	}

	return nil
}

//------------------------------------------------------------------------------

// Key implements the Locatable interface.
func (c *NodeCollection) locationKey() string {
	return "nodes"
}

// Parent implements the Locatable interface.
func (c *NodeCollection) parent() (l Locatable) {
	return
}

// Child implements the Locatable interface.
func (c *NodeCollection) child(key string) Locatable {
	r, err := c.Get(common.IDString(key))
	if err != nil {
		panic(fmt.Errorf("No child with key %s for %T", key, c))
	}
	return r
}

// Key implements the Locatable interface.
func (r *NodeResource) locationKey() string {
	return common.StringID(r.ID)
}

// Parent implements the Locatable interface.
func (r *NodeResource) parent() Locatable {
	return r.collection.(Locatable)
}

// Child implements the Locatable interface.
func (r *NodeResource) child(key string) (l Locatable) {
	switch key {
	default:
		panic(fmt.Errorf("No child with key %s for %T", key, r))
	}
}

// Action implements the Resource interface.
func (r *NodeResource) Action(name string) *Action {
	// var fn ActionPerformer
	switch name {
	default:
		panic(fmt.Errorf("No action %s for Node", name))
	}
	// return &Action{
	// 	ActionName: name,
	// 	core:       r.core,
	// 	resource:   r,
	// 	performer:  fn,
	// }
}

//------------------------------------------------------------------------------

// decorate implements the Resource interface
func (r *NodeResource) decorate() error {
	k8sNode, err := r.core.k8s.Nodes().Get(r.Name)

	// TODO
	// we need to check error type, and figure out what to actually do with status
	if err != nil {
		r.Status = "NOT_READY"
		return nil
	}

	// r.ServerUptime = int(time.Since(r.LaunchTime).Seconds())

	r.Status = "READY"
	r.ExternalIP = k8sNode.ExternalIP()
	r.OutOfDisk = k8sNode.IsOutOfDisk()

	stats, err := k8sNode.HeapsterStats()
	if err != nil {
		Log.Errorf("Could not load Heapster stats for node %s", r.Name)
		// return err
	} else {

		// NOTE Heapster apparently can return stats without all the needed fields

		cpuUsage := stats.Stats["cpu-usage"]
		memUsage := stats.Stats["memory-usage"]

		if cpuUsage != nil && memUsage != nil {
			r.CPU = &common.ResourceMetrics{
				Usage: cpuUsage.Minute.Average,
				Limit: stats.Stats["cpu-limit"].Minute.Average,
			}
			r.RAM = &common.ResourceMetrics{
				Usage: memUsage.Minute.Average,
				Limit: stats.Stats["memory-limit"].Minute.Average,
			}
		}
	}

	return nil
}

// Update is a proxy method to NodeCollection's Update.
func (r *NodeResource) Update() error {
	return r.collection.Update(r.ID, r)
}

// Patch is a proxy method to collection Patch.
func (r *NodeResource) Patch() error {
	return r.collection.Patch(r.ID, r)
}

// Delete is a proxy method to NodeCollection's Delete.
func (r *NodeResource) Delete() error {
	return r.collection.Delete(r)
}

func (c *NodeCollection) createNodeFromServer(server *ec2.Instance) (*NodeResource, error) {
	node := c.New()
	node.ID = server.InstanceId
	node.Name = *server.PrivateDnsName
	node.Class = *server.InstanceType
	node.ProviderCreationTimestamp = &common.Timestamp{Time: *server.LaunchTime}
	// ExternalIP:   *server.PublicIpAddress,

	// NOTE we can't call c.Create(node) here because that method utilizes
	// populate().
	if err := c.core.db.create(c, node.ID, node); err != nil {
		return nil, err
	}

	return node, nil
}

func (c *NodeCollection) createServer(class string) (*ec2.Instance, error) {
	group, err := autoscalingGroupByInstanceType(c.core, class)
	if err != nil {
		return nil, err
	}

	// TODO autoscaling group operations are not atomic by way of design; race
	// conditions can occur between fetching / updating the size fields.
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
	_, err = c.core.autoscaling.UpdateAutoScalingGroup(input)

	var serverID *string

	desc := fmt.Sprintf(
		"autoscaling group %s scale up from %d to %d instances",
		*group.AutoScalingGroupName,
		startingCapacity,
		desiredCapacity,
	)
	err = common.WaitFor(desc, 60*time.Second, 2*time.Second, func() (bool, error) {
		group, err := autoscalingGroup(c.core, group.AutoScalingGroupName)
		if err != nil {
			return false, err
		}

		Log.Debugf("Waiting for %s", desc)

		for _, instance := range group.Instances {
			if _, ok := origInstanceIDs[*instance.InstanceId]; !ok { // if this is a new instance ID
				serverID = instance.InstanceId
				return true, nil
			}
		}
		return false, nil
	})

	return ec2Instance(c.core, serverID)
}

func (r *NodeResource) deleteServer() error {

	// Do we need to delete Node in K8S? Any benefit to that?

	input := &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		InstanceId:                     r.ID,
		ShouldDecrementDesiredCapacity: aws.Bool(true),
	}
	_, err := r.core.autoscaling.TerminateInstanceInAutoScalingGroup(input)
	if err != nil && strings.Contains(err.Error(), "Instance Id not found") {
		return nil
	}
	return err
}

func (r *NodeResource) hasPodsWithReservedResources() (bool, error) {
	q := &guber.QueryParams{
		FieldSelector: "spec.nodeName=" + r.Name + ",status.phase=Running",
	}
	// TODO does this actually return pods in multiple namespaces?
	pods, err := r.core.k8s.Pods("").Query(q)
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
