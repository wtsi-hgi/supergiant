package core

import (
	"strings"

	"github.com/supergiant/supergiant/common"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elb"
)

type EntrypointsInterface interface {
	List() (*EntrypointList, error)
	New() *EntrypointResource
	Create(*EntrypointResource) error
	Get(common.ID) (*EntrypointResource, error)
	Update(common.ID, *EntrypointResource) error
	Delete(*EntrypointResource) error
}

type EntrypointCollection struct {
	core *Core
}

type EntrypointResource struct {
	core       *Core
	collection EntrypointsInterface
	*common.Entrypoint
}

// NOTE this does not inherit from common like model does; all we need is a List
// object, internally, that has a slice of our composed model above.
type EntrypointList struct {
	Items []*EntrypointResource `json:"items"`
}

// initializeResource implements the Collection interface.
func (c *EntrypointCollection) initializeResource(in Resource) {
	r := in.(*EntrypointResource)
	r.collection = c
	r.core = c.core
}

// List returns an EntrypointList.
func (c *EntrypointCollection) List() (*EntrypointList, error) {
	list := new(EntrypointList)
	err := c.core.db.list(c, list)
	return list, err
}

// New initializes an Entrypoint with a pointer to the Collection.
func (c *EntrypointCollection) New() *EntrypointResource {
	r := &EntrypointResource{
		Entrypoint: &common.Entrypoint{
			Meta: common.NewMeta(),
		},
	}
	c.initializeResource(r)
	return r
}

// Create takes an Entrypoint and creates it in etcd, and creates an AWS ELB.
func (c *EntrypointCollection) Create(r *EntrypointResource) error {
	if err := c.core.db.create(c, r.Domain, r); err != nil {
		return err
	}

	// TODO for error handling and retries, we may want to do this in a task and
	// utilize a Status field
	address, err := r.createELB()
	if err != nil {
		return err
	}
	if err := r.attachELBToScalingGroups(); err != nil {
		return err
	}
	if err := r.configureELBHealthCheck(); err != nil {
		return err
	}

	r.Address = *address
	r.Update()

	return nil
}

// Get takes a name and returns an EntrypointResource if it exists.
func (c *EntrypointCollection) Get(domain common.ID) (*EntrypointResource, error) {
	r := c.New()
	if err := c.core.db.get(c, domain, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Update saves the Entrypoint in etcd through an update.
func (c *EntrypointCollection) Update(domain common.ID, r *EntrypointResource) error {
	return c.core.db.update(c, domain, r)
}

// Delete cascades deletes to all Components, deletes the Kube Namespace, and
// deletes the Entrypoint in etcd.
func (c *EntrypointCollection) Delete(r *EntrypointResource) error {
	if err := r.detachELBFromScalingGroups(); err != nil {
		return err
	}
	if err := r.deleteELB(); err != nil {
		return err
	}
	return c.core.db.delete(c, r.Domain)
}

//------------------------------------------------------------------------------

// Key implements the Locatable interface.
func (c *EntrypointCollection) locationKey() string {
	return "entrypoints"
}

// Parent implements the Locatable interface.
func (c *EntrypointCollection) parent() (l Locatable) {
	return
}

// Child implements the Locatable interface.
func (c *EntrypointCollection) child(key string) Locatable {
	r, err := c.Get(common.IDString(key))
	if err != nil {
		Log.Panicf("No child with key %s for %T", key, c)
	}
	return r
}

// Key implements the Locatable interface.
func (r *EntrypointResource) locationKey() string {
	return common.StringID(r.Domain)
}

// Parent implements the Locatable interface.
func (r *EntrypointResource) parent() Locatable {
	return r.collection.(Locatable)
}

// Child implements the Locatable interface.
func (r *EntrypointResource) child(key string) (l Locatable) {
	switch key {
	default:
		Log.Panicf("No child with key %s for %T", key, r)
	}
	return
}

// Action implements the Resource interface.
func (r *EntrypointResource) Action(name string) *Action {
	var fn ActionPerformer
	switch name {
	default:
		Log.Panicf("No action %s for Entrypoint", name)
	}
	return &Action{
		ActionName: name,
		core:       r.core,
		resource:   r,
		performer:  fn,
	}
}

//------------------------------------------------------------------------------

// decorate implements the Resource interface
func (r *EntrypointResource) decorate() (err error) {
	return
}

// Update is a proxy method to EntrypointCollection's Update.
func (r *EntrypointResource) Update() error {
	return r.collection.Update(r.Domain, r)
}

// Delete is a proxy method to EntrypointCollection's Delete.
func (r *EntrypointResource) Delete() error {
	return r.collection.Delete(r)
}

// AddPort creates a listener on the ELB.
func (r *EntrypointResource) AddPort(elbPort int, instancePort int) error {
	params := &elb.CreateLoadBalancerListenersInput{
		LoadBalancerName: r.awsName(),
		Listeners: []*elb.Listener{
			{
				InstancePort:     aws.Int64(int64(instancePort)),
				LoadBalancerPort: aws.Int64(int64(elbPort)),
				Protocol:         aws.String("TCP"),
				// InstanceProtocol: aws.String("Protocol"),
				// SSLCertificateId: aws.String("SSLCertificateId"),
			},
		},
	}

	Log.Infof("Adding port %d:%d to ELB %s", elbPort, instancePort, *r.awsName())

	_, err := r.core.elb.CreateLoadBalancerListeners(params)
	return err
}

// RemovePort removes a listener from the ELB.
func (r *EntrypointResource) RemovePort(elbPort int) error {
	params := &elb.DeleteLoadBalancerListenersInput{
		LoadBalancerName: r.awsName(),
		LoadBalancerPorts: []*int64{
			aws.Int64(int64(elbPort)),
		},
	}

	Log.Infof("Removing port %d from ELB %s", elbPort, *r.awsName())

	_, err := r.core.elb.DeleteLoadBalancerListeners(params)
	return err
}

func (r *EntrypointResource) awsName() *string {
	// TODO add a unique cloud ID, load instead of return from func
	// Also this is just really crazy. Should probably just specify name/domain
	// separately.
	suffix := strings.Replace(common.StringID(r.Domain), ".", "-", -1)
	return aws.String("supergiant-" + suffix)
}

func (r *EntrypointResource) createELB() (*string, error) {
	params := &elb.CreateLoadBalancerInput{
		Listeners: []*elb.Listener{ // Required
			{
				InstancePort:     aws.Int64(8080),    // Required
				LoadBalancerPort: aws.Int64(8080),    // Required
				Protocol:         aws.String("HTTP"), // Required
				// InstanceProtocol: aws.String("Protocol"),
				// SSLCertificateId: aws.String("SSLCertificateId"),
			},
		},
		LoadBalancerName: r.awsName(),
		Scheme:           aws.String("internet-facing"),
		SecurityGroups: []*string{
			aws.String(r.core.AwsSgID),
		},
		Subnets: []*string{
			aws.String(r.core.AwsSubnetID),
		},
	}
	resp, err := r.core.elb.CreateLoadBalancer(params)
	if err != nil {
		return nil, err
	}
	return resp.DNSName, nil
}

func (r *EntrypointResource) configureELBHealthCheck() error {
	params := &elb.ConfigureHealthCheckInput{
		LoadBalancerName: r.awsName(),
		HealthCheck: &elb.HealthCheck{
			Target:             aws.String("HTTPS:10250/healthz"),
			HealthyThreshold:   aws.Int64(2),
			UnhealthyThreshold: aws.Int64(10),
			Interval:           aws.Int64(30),
			Timeout:            aws.Int64(5),
		},
	}
	_, err := r.core.elb.ConfigureHealthCheck(params)
	return err
}

func (r *EntrypointResource) attachELBToScalingGroups() error {
	groups, err := autoscalingGroups(r.core)
	if err != nil {
		return err
	}

	for _, group := range groups {
		params := &autoscaling.AttachLoadBalancersInput{
			AutoScalingGroupName: group.AutoScalingGroupName,
			LoadBalancerNames:    []*string{r.awsName()},
		}
		if _, err := r.core.autoscaling.AttachLoadBalancers(params); err != nil {
			return err
		}
	}
	return nil
}

func (r *EntrypointResource) detachELBFromScalingGroups() error {
	groups, err := autoscalingGroups(r.core)
	if err != nil {
		return err
	}

	for _, group := range groups {
		params := &autoscaling.DetachLoadBalancersInput{
			AutoScalingGroupName: group.AutoScalingGroupName,
			LoadBalancerNames:    []*string{r.awsName()},
		}
		if _, err := r.core.autoscaling.DetachLoadBalancers(params); err != nil {
			// TODO should find the error type
			if strings.Contains(err.Error(), "Trying to remove Load Balancers that are not part of the group") {
				continue
			} else {
				return err
			}
		}
	}
	return nil
}

func (r *EntrypointResource) deleteELB() error {
	params := &elb.DeleteLoadBalancerInput{
		LoadBalancerName: r.awsName(),
	}
	_, err := r.core.elb.DeleteLoadBalancer(params)
	return err
}
