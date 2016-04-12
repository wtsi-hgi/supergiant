package core

import (
	"path"
	"strings"

	"github.com/supergiant/supergiant/common"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elb"
)

type EntrypointCollection struct {
	core *Core
}

type EntrypointResource struct {
	collection *EntrypointCollection
	*common.Entrypoint
}

// NOTE this does not inherit from common like model does; all we need is a List
// object, internally, that has a slice of our composed model above.
type EntrypointList struct {
	Items []*EntrypointResource `json:"items"`
}

// etcdKey implements the Collection interface.
func (c *EntrypointCollection) etcdKey(domain common.ID) string {
	key := "/entrypoints"
	if domain != nil {
		key = path.Join(key, common.StringID(domain))
	}
	return key
}

// initializeResource implements the Collection interface.
func (c *EntrypointCollection) initializeResource(r Resource) error {
	resource := r.(*EntrypointResource)
	resource.collection = c
	return nil
}

// List returns an EntrypointList.
func (c *EntrypointCollection) List() (*EntrypointList, error) {
	list := new(EntrypointList)
	err := c.core.db.list(c, list)
	return list, err
}

// New initializes an Entrypoint with a pointer to the Collection.
func (c *EntrypointCollection) New() *EntrypointResource {
	return &EntrypointResource{
		collection: c,
		Entrypoint: &common.Entrypoint{
			Meta: common.NewMeta(),
		},
	}
}

// Create takes an Entrypoint and creates it in etcd. It also creates a Kubernetes
// Namespace with the name of the Entrypoint.
func (c *EntrypointCollection) Create(r *EntrypointResource) (*EntrypointResource, error) {
	if err := c.core.db.create(c, r.Domain, r); err != nil {
		return nil, err
	}

	// TODO for error handling and retries, we may want to do this in a task and
	// utilize a Status field
	address, err := r.createELB()
	if err != nil {
		return nil, err
	}
	if err := r.attachELBToScalingGroups(); err != nil {
		return nil, err
	}
	if err := r.configureELBHealthCheck(); err != nil {
		return nil, err
	}

	r.Address = *address
	r.Update()

	return r, nil
}

// Get takes a name and returns an EntrypointResource if it exists.
func (c *EntrypointCollection) Get(domain common.ID) (*EntrypointResource, error) {
	r := c.New()
	if err := c.core.db.get(c, domain, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Resource-level
//==============================================================================

// Update saves the Entrypoint in etcd through an update.
func (r *EntrypointResource) Update() error {
	return r.collection.core.db.update(r.collection, r.Domain, r)
}

// Delete cascades deletes to all Components, deletes the Kube Namespace, and
// deletes the Entrypoint in etcd.
func (r *EntrypointResource) Delete() error {
	if err := r.detachELBFromScalingGroups(); err != nil {
		return err
	}
	if err := r.deleteELB(); err != nil {
		return err
	}
	return r.collection.core.db.delete(r.collection, r.Domain)
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

	_, err := r.collection.core.elb.CreateLoadBalancerListeners(params)
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

	_, err := r.collection.core.elb.DeleteLoadBalancerListeners(params)
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
			aws.String(AwsSgID),
		},
		Subnets: []*string{
			aws.String(AwsSubnetID),
		},
	}
	resp, err := r.collection.core.elb.CreateLoadBalancer(params)
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
	_, err := r.collection.core.elb.ConfigureHealthCheck(params)
	return err
}

func (r *EntrypointResource) autoscalingGroups() (groups []*autoscaling.Group, err error) {
	params := &autoscaling.DescribeAutoScalingGroupsInput{
		// VPCZoneIdentifier: aws.String(AwsSubnetID),
		// NOTE I think we have to just filter this client-side? Seems weird
		MaxRecords: aws.Int64(100),
	}
	resp, err := r.collection.core.autoscaling.DescribeAutoScalingGroups(params)
	if err != nil {
		return nil, err
	}

	for _, group := range resp.AutoScalingGroups {
		if *group.VPCZoneIdentifier == AwsSubnetID {
			groups = append(groups, group)
		}
	}
	return groups, nil
}

func (r *EntrypointResource) attachELBToScalingGroups() error {
	groups, err := r.autoscalingGroups()
	if err != nil {
		return err
	}

	for _, group := range groups {
		params := &autoscaling.AttachLoadBalancersInput{
			AutoScalingGroupName: group.AutoScalingGroupName,
			LoadBalancerNames:    []*string{r.awsName()},
		}
		if _, err := r.collection.core.autoscaling.AttachLoadBalancers(params); err != nil {
			return err
		}
	}
	return nil
}

func (r *EntrypointResource) detachELBFromScalingGroups() error {
	groups, err := r.autoscalingGroups()
	if err != nil {
		return err
	}

	for _, group := range groups {
		params := &autoscaling.DetachLoadBalancersInput{
			AutoScalingGroupName: group.AutoScalingGroupName,
			LoadBalancerNames:    []*string{r.awsName()},
		}
		if _, err := r.collection.core.autoscaling.DetachLoadBalancers(params); err != nil {
			// TODO is this bad practice?
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
	_, err := r.collection.core.elb.DeleteLoadBalancer(params)
	return err
}
