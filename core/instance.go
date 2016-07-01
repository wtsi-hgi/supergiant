package core

import (
	"fmt"
	"strconv"
	"time"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

type InstancesInterface interface {
	App() *AppResource
	Component() *ComponentResource
	Release() *ReleaseResource

	List() *InstanceList
	New(common.ID) *InstanceResource
	Get(common.ID) (*InstanceResource, error)
	Start(Resource) error
	Stop(Resource) error
	Delete(*InstanceResource) error
	DeleteVolumes(*InstanceResource) error
}

type InstanceCollection struct {
	core    *Core
	release *ReleaseResource
}

type InstanceResource struct {
	core *Core
	*common.Instance

	Collection InstancesInterface `json:"-"`

	serviceSet *ServiceSet
}

type InstanceList struct {
	Items []*InstanceResource `json:"items"`
}

func (c *InstanceCollection) App() *AppResource {
	return c.release.Component().App()
}

func (c *InstanceCollection) Component() *ComponentResource {
	return c.release.Component()
}

func (c *InstanceCollection) Release() *ReleaseResource {
	return c.release
}

// List returns an InstanceList.
func (c *InstanceCollection) List() *InstanceList {
	list := new(InstanceList)
	list.Items = make([]*InstanceResource, 0)
	for i := 0; i < c.release.InstanceCount; i++ {
		id := strconv.Itoa(i)
		list.Items = append(list.Items, c.New(&id))
	}
	return list
}

// New initializes an Instance with a pointer to the Collection.
func (c *InstanceCollection) New(id common.ID) *InstanceResource {
	r := &InstanceResource{
		core:       c.core,
		Collection: c,
		Instance: &common.Instance{
			ID: id,
		},
	}
	// TODO not consistent with the setter approach
	r.BaseName = common.StringID(r.Component().Name) + "-" + common.StringID(r.ID)
	r.Name = r.BaseName + common.StringID(r.Release().InstanceGroup)

	// TODO definitely have to take this panic out
	if err := r.decorate(); err != nil {
		panic(err)
	}
	return r
}

// Get takes an id and returns an InstanceResource if it exists.
func (c *InstanceCollection) Get(id common.ID) (*InstanceResource, error) {
	index, err := strconv.Atoi(common.StringID(id))
	if err != nil {
		return nil, err
	}
	maxIndex := c.release.InstanceCount - 1
	if index < 0 || index > maxIndex {
		return nil, fmt.Errorf("%d for Instance ID is out of range; Highest ID is %d", index, maxIndex)
	}
	return c.New(id), nil
}

func (c *InstanceCollection) Start(ri Resource) error {
	r := ri.(*InstanceResource)

	if err := r.prepareVolumes(); err != nil {
		return err
	}
	if err := r.serviceSet.provision(); err != nil {
		return err
	}

	if err := r.serviceSet.addNewPorts(); err != nil {
		return err
	}
	if err := r.serviceSet.removeOldPorts(); err != nil {
		return err
	}

	if err := r.provisionReplicationController(); err != nil {
		return err
	}
	return nil
}

func (c *InstanceCollection) Stop(ri Resource) error {
	r := ri.(*InstanceResource)

	if err := r.deleteReplicationControllerAndPod(); err != nil {
		return err
	}
	for _, vol := range r.Volumes() {
		if err := vol.waitForAvailable(); err != nil {
			return err
		}
	}

	// TODO need a wait in here (optional) for the pod to be deleted

	return nil
}

func (c *InstanceCollection) Delete(r *InstanceResource) (err error) {
	if err := r.deleteReplicationControllerAndPod(); err != nil {
		return err
	}
	if err := r.serviceSet.delete(); err != nil {
		return err
	}
	return
}

func (c *InstanceCollection) DeleteVolumes(r *InstanceResource) error {
	return r.deleteVolumes()
}

//------------------------------------------------------------------------------

// Key implements the Locatable interface.
func (c *InstanceCollection) locationKey() string {
	return "instances"
}

// Parent implements the Locatable interface.
func (c *InstanceCollection) parent() Locatable {
	return c.release
}

// Child implements the Locatable interface.
func (c *InstanceCollection) child(key string) Locatable {
	r, err := c.Get(common.IDString(key))
	if err != nil {
		panic(fmt.Errorf("No child with key %s for %T", key, c))
	}
	return r
}

// Key implements the Locatable interface.
func (r *InstanceResource) locationKey() string {
	return common.StringID(r.ID)
}

// Parent implements the Locatable interface.
func (r *InstanceResource) parent() Locatable {
	return r.Collection.(Locatable)
}

// Child implements the Locatable interface.
func (r *InstanceResource) child(key string) (l Locatable) {
	switch key {
	default:
		panic(fmt.Errorf("No child with key %s for %T", key, r))
	}
}

//------------------------------------------------------------------------------

// decorate implements the Resource interface
func (r *InstanceResource) decorate() error {
	var oldServiceSet *ServiceSet
	component := r.Release().Component()
	if component.TargetReleaseTimestamp != nil && component.CurrentReleaseTimestamp != nil && *component.TargetReleaseTimestamp == *r.Release().Timestamp {
		currentRelease, err := component.CurrentRelease()
		if err != nil {
			return err
		}
		var oldInstance *InstanceResource
		for _, instance := range currentRelease.Instances().List().Items {
			if *instance.ID == *r.ID {
				oldInstance = instance
				break
			}
		}
		if oldInstance != nil {
			oldServiceSet = oldInstance.serviceSet
		}
	}

	r.serviceSet = &ServiceSet{
		core:          r.core,
		release:       r.Release(),
		namespace:     common.StringID(r.App().Name),
		baseName:      r.BaseName,
		labelSelector: map[string]string{"instance_service": r.BaseName},
		portFilter:    func(port *common.Port) bool { return port.PerInstance },
		previous:      oldServiceSet,
	}

	pod, err := r.pod()
	if err != nil {
		return err
	}

	if pod != nil && pod.IsReady() {
		r.Status = common.InstanceStatusStarted
	} else {
		r.Status = common.InstanceStatusStopped
		return nil // don't get stats below
	}

	externalAddrs, err := r.externalAddresses()
	if err != nil {
		return err
	}
	internalAddrs, err := r.internalAddresses()
	if err != nil {
		return err
	}

	r.Addresses = &common.Addresses{
		External: externalAddrs,
		Internal: internalAddrs,
	}

	stats, err := pod.HeapsterStats()

	if err != nil {
		Log.Errorf("Could not load Heapster stats for pod %s", r.Name)
		// return err
	} else {

		// TODO repeated in Node

		cpuUsage := stats.Stats["cpu-usage"]
		memUsage := stats.Stats["memory-usage"]

		// NOTE we took out the limit displayed by Heapster, and replaced it with
		// the actual limit value assigned to the pod. Heapster was returning limit
		// values less than the usage, which was causing errors when calculating
		// percentages.

		if cpuUsage != nil && memUsage != nil {
			r.CPU = &common.ResourceMetrics{
				Usage: cpuUsage.Minute.Average,
				Limit: totalCpuLimit(pod).Millicores,
				// Limit: stats.Stats["cpu-limit"].Minute.Average,
			}
			r.RAM = &common.ResourceMetrics{
				Usage: memUsage.Minute.Average,
				Limit: int(totalRamLimit(pod).Bytes),
				// Limit: stats.Stats["memory-limit"].Minute.Average,
			}
		}
	}

	return nil
}

// TODO repeated code in component
func (r *InstanceResource) externalAddresses() (addrs []*common.PortAddress, err error) {
	ports, err := r.serviceSet.externalPorts()
	if err != nil {
		return nil, err
	}
	for _, port := range ports {
		addrs = append(addrs, port.externalAddress())
	}
	return addrs, nil
}

func (r *InstanceResource) internalAddresses() (addrs []*common.PortAddress, err error) {
	iPorts, err := r.serviceSet.internalPorts()
	if err != nil {
		return nil, err
	}
	ePorts, err := r.serviceSet.externalPorts() // external ports also have internal addresses
	if err != nil {
		return nil, err
	}
	ports := append(iPorts, ePorts...)
	for _, port := range ports {
		addrs = append(addrs, port.internalAddress())
	}
	return addrs, nil
}

func (r *InstanceResource) App() *AppResource {
	return r.Collection.App()
}

func (r *InstanceResource) Component() *ComponentResource {
	return r.Collection.Component()
}

func (r *InstanceResource) IsStarted() bool {
	return r.Status == common.InstanceStatusStarted
}

func (r *InstanceResource) IsStopped() bool {
	return r.Status == common.InstanceStatusStopped
}

// Delete tears down the instance
func (r *InstanceResource) Delete() error {
	return r.Collection.Delete(r)
}

// The following 2 are only diff from Provision() and Delete() in that they do
// not delete the create or delete the volumes.
func (r *InstanceResource) Start() error {
	return r.Collection.Start(r)
}

func (r *InstanceResource) Stop() error {
	return r.Collection.Stop(r)
}

// TODO we need a better way of initializing defaults on sub-resources
func (r *InstanceResource) DefaultContainerName() string {
	return asKubeContainer(r.Release().Containers[0], r).Name
}

// func (r *InstanceResource) Exec(container string, command string) (string, error) {
// 	pod, err := r.pod()
// 	if err != nil {
// 		return "", err
// 	}
//
// 	// TODO nil pointer possibility if pod is nil
//
// 	return pod.Exec(container, command)
// }

func (r *InstanceResource) Log() (string, error) {
	pod, err := r.pod()
	if err != nil {
		return "", err
	}

	// TODO nil pointer possibility if pod is nil

	return pod.Log(r.DefaultContainerName())
}

func (r *InstanceResource) Release() *ReleaseResource {
	return r.Collection.Release()
}

func (r *InstanceResource) Volumes() (vols []*AwsVolume) {
	for _, blueprint := range r.Release().Volumes {
		vol := &AwsVolume{
			core:      r.core,
			Blueprint: blueprint,
			Instance:  r,
		}
		vols = append(vols, vol)
	}
	return vols
}

func (r *InstanceResource) prepareVolumes() error {
	//resize volumes (concurrently) if needed
	c := make(chan error)
	// really, we should be resizing all or none. this is just so we don't wait
	// forever below when not resizing
	volsResizing := 0
	for _, vol := range r.Volumes() {
		if vol.needsResize() {
			volsResizing++
			go func(vol *AwsVolume) {
				c <- vol.resize()
			}(vol)
		}
	}
	for i := 0; i < volsResizing; i++ {
		if err := <-c; err != nil {
			return err
		}
	}
	return nil
}

func (r *InstanceResource) kubeVolumes() (vols []*guber.Volume, err error) {
	for _, vol := range r.Volumes() {
		kubeVol, err := asKubeVolume(vol)
		if err != nil {
			return nil, err
		}
		vols = append(vols, kubeVol)
	}
	return vols, nil
}

func (r *InstanceResource) kubeContainers() (containers []*guber.Container) {
	for _, blueprint := range r.Release().Containers {
		containers = append(containers, asKubeContainer(blueprint, r))
	}
	return containers
}

func (r *InstanceResource) replicationController() (*guber.ReplicationController, error) {
	return r.core.k8s.ReplicationControllers(common.StringID(r.App().Name)).Get(r.Name)
}

func (r *InstanceResource) waitForReplicationControllerReady() error {
	Log.Infof("Waiting for ReplicationController %s to start", r.Name)
	start := time.Now()
	maxWait := 5 * time.Minute
	for elapsed := time.Since(start); elapsed < maxWait; {
		rc, err := r.replicationController()
		if err != nil {
			return err
		} else if rc.Status.Replicas == 1 { // TODO this may not assert pod running
			return nil
		}
	}
	return fmt.Errorf("Timed out waiting for RC '%s' to start", r.Name)
}

func (r *InstanceResource) provisionReplicationController() error {
	if _, err := r.replicationController(); err == nil {
		return nil // already provisioned
	} else if !isKubeNotFoundErr(err) {
		return err
	}

	// We load them here because the repos may not exist, which needs to return error
	imagePullSecrets, err := r.Release().ImagePullSecrets()
	if err != nil {
		return err
	}

	kubeVolumes, err := r.kubeVolumes()
	if err != nil {
		return err
	}

	rc := &guber.ReplicationController{
		Metadata: &guber.Metadata{
			Name: r.Name,
		},
		Spec: &guber.ReplicationControllerSpec{
			Selector: map[string]string{
				"instance": r.Name,
			},
			Replicas: 1,
			Template: &guber.PodTemplate{
				Metadata: &guber.Metadata{
					Name: r.Name, // pod base name is same as RC
					Labels: map[string]string{
						"service":          common.StringID(r.Component().Name), // for Service
						"instance":         r.Name,                              // for RC (above)
						"instance_service": r.BaseName,                          // for Instance Service
					},
				},
				Spec: &guber.PodSpec{
					Volumes:                       kubeVolumes,
					Containers:                    r.kubeContainers(),
					ImagePullSecrets:              imagePullSecrets,
					TerminationGracePeriodSeconds: r.Release().TerminationGracePeriod,
				},
			},
		},
	}
	Log.Infof("Creating ReplicationController %s", r.Name)
	if _, err = r.core.k8s.ReplicationControllers(common.StringID(r.App().Name)).Create(rc); err != nil {
		return err
	}
	return r.waitForReplicationControllerReady()
}

func (r *InstanceResource) pod() (*guber.Pod, error) {
	q := &guber.QueryParams{
		LabelSelector: "instance=" + r.Name,
	}
	pods, err := r.core.k8s.Pods(common.StringID(r.App().Name)).Query(q)
	if err != nil {
		return nil, err // Not sure what the error might be here
	}

	if len(pods.Items) == 1 {
		return pods.Items[0], nil
	}

	// NOTE this does not return error if the pod cannot be found
	return nil, nil
}

func (r *InstanceResource) deleteReplicationControllerAndPod() error {
	Log.Infof("Deleting ReplicationController %s", r.Name)

	// TODO we call r.core.k8s.ReplicationControllers(r.App().Name)
	// nough to warrant its own method
	if err := r.core.k8s.ReplicationControllers(common.StringID(r.App().Name)).Delete(r.Name); err != nil && !isKubeNotFoundErr(err) {
		return err
	}

	pod, err := r.pod()
	if err != nil {
		return err
	}
	if pod != nil {
		if err := pod.Delete(); err != nil && !isKubeNotFoundErr(err) {
			return err
		}
	}

	return nil
}

// exposed for use in deploy_component.go
func (r *InstanceResource) deleteVolumes() error {
	for _, vol := range r.Volumes() {
		if err := vol.Delete(); err != nil { // NOTE this should not be a "not found" error -- since Volumes() will naturally do an existence check
			return err
		}
	}
	return nil
}
