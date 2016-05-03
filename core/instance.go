package core

import (
	"fmt"
	"strconv"
	"time"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

type InstancesInterface interface {
	// Release() *ReleaseResource

	List() *InstanceList
	New(common.ID) *InstanceResource
	Get(common.ID) (*InstanceResource, error)
	Start(Resource) error
	Stop(Resource) error
}

type InstanceCollection struct {
	core    *Core
	release *ReleaseResource
}

type InstanceResource struct {
	core       *Core
	collection *InstanceCollection
	*common.Instance
}

type InstanceList struct {
	Items []*InstanceResource `json:"items"`
}

func (c *InstanceCollection) app() *AppResource {
	return c.release.Component().App()
}

func (c *InstanceCollection) component() *ComponentResource {
	return c.release.Component()
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
		collection: c,
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
		Log.Panicf("No child with key %s for %T", key, c)
	}
	return r
}

// Key implements the Locatable interface.
func (r *InstanceResource) locationKey() string {
	return common.StringID(r.ID)
}

// Parent implements the Locatable interface.
func (r *InstanceResource) parent() Locatable {
	return r.collection
}

// Child implements the Locatable interface.
func (r *InstanceResource) child(key string) (l Locatable) {
	switch key {
	default:
		Log.Panicf("No child with key %s for %T", key, r)
	}
	return
}

// Action implements the Resource interface.
func (r *InstanceResource) Action(name string) *Action {
	var fn ActionPerformer
	switch name {
	case "start":
		fn = ActionPerformer(r.collection.Start)
	case "stop":
		fn = ActionPerformer(r.collection.Stop)
	default:
		Log.Panicf("No action %s for Instance", name)
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
func (r *InstanceResource) decorate() error {
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

func (r *InstanceResource) App() *AppResource {
	return r.collection.app()
}

func (r *InstanceResource) Component() *ComponentResource {
	return r.collection.component()
}

func (r *InstanceResource) IsStarted() bool {
	return r.Status == common.InstanceStatusStarted
}

func (r *InstanceResource) IsStopped() bool {
	return r.Status == common.InstanceStatusStopped
}

// Delete tears down the instance
func (r *InstanceResource) Delete() (err error) {
	if err = r.deleteReplicationControllerAndPod(); err != nil {
		return err
	}
	if err = r.DeleteVolumes(); err != nil {
		return err
	}
	return nil
}

// The following 2 are only diff from Provision() and Delete() in that they do
// not delete the create or delete the volumes.
func (r *InstanceResource) Start() error {
	return r.collection.Start(r)
}

func (r *InstanceResource) Stop() error {
	return r.collection.Stop(r)
}

func (r *InstanceResource) Log() (string, error) {
	pod, err := r.pod()
	if err != nil {
		return "", err
	}

	// TODO nil pointer possibility if pod is nil

	// TODO we need a better way of initializing defaults on sub-resources
	containerName := asKubeContainer(r.Release().Containers[0], r).Name

	return pod.Log(containerName)
}

func (r *InstanceResource) Release() *ReleaseResource {
	return r.collection.release
}

func (r *InstanceResource) Volumes() (vols []*AwsVolume) {
	for _, blueprint := range r.Release().Volumes {
		vol := &AwsVolume{
			core:      r.collection.core,
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
	return r.collection.core.k8s.ReplicationControllers(common.StringID(r.App().Name)).Get(r.Name)
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
						"service":  common.StringID(r.Component().Name), // for Service
						"instance": r.Name,                              // for RC (above)
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
	if _, err = r.collection.core.k8s.ReplicationControllers(common.StringID(r.App().Name)).Create(rc); err != nil {
		return err
	}
	return r.waitForReplicationControllerReady()
}

func (r *InstanceResource) pod() (*guber.Pod, error) {
	q := &guber.QueryParams{
		LabelSelector: "instance=" + r.Name,
	}
	pods, err := r.collection.core.k8s.Pods(common.StringID(r.App().Name)).Query(q)
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
	// TODO we call r.collection.core.k8s.ReplicationControllers(r.App().Name)
	// nough to warrant its own method
	if err := r.collection.core.k8s.ReplicationControllers(common.StringID(r.App().Name)).Delete(r.Name); err != nil && !isKubeNotFoundErr(err) {
		return err
	}
	pod, err := r.pod()
	if err != nil {
		return err
	}
	if pod != nil {
		// _ is found bool, we don't care if it was found or not, just don't want an error
		if err := pod.Delete(); err != nil && !isKubeNotFoundErr(err) {
			return err
		}
	}
	return nil
}

// exposed for use in deploy_component.go
func (r *InstanceResource) DeleteVolumes() error {
	for _, vol := range r.Volumes() {
		if err := vol.Delete(); err != nil { // NOTE this should not be a "not found" error -- since Volumes() will naturally do an existence check
			return err
		}
	}
	return nil
}
