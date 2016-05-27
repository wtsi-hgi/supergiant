package core

import (
	"fmt"

	"github.com/supergiant/supergiant/common"
	"github.com/supergiant/supergiant/deploy"
)

type ComponentsInterface interface {
	// A simple getter since App is an attribute on actual Collections implementing this
	App() *AppResource

	List() (*ComponentList, error)
	New() *ComponentResource
	Create(*ComponentResource) error
	Get(common.ID) (*ComponentResource, error)
	Update(common.ID, *ComponentResource) error
	Patch(common.ID, *ComponentResource) error
	Deploy(Resource) error
	Delete(Resource) error
}

// ComponentCollection implements ComponentsInterface.
type ComponentCollection struct {
	core *Core
	app  *AppResource
}

type ComponentResource struct {
	core       *Core
	collection ComponentsInterface
	*common.Component

	// Relations
	ReleasesInterface ReleasesInterface `json:"-"`
}

type ComponentList struct {
	Items []*ComponentResource `json:"items"`
}

// initializeResource implements the Collection interface.
func (c *ComponentCollection) initializeResource(in Resource) {
	r := in.(*ComponentResource)
	r.collection = c
	r.core = c.core
	if r.ReleasesInterface == nil { // don't want to reset for testing purposes
		r.ReleasesInterface = &ReleaseCollection{
			core:      c.core,
			component: r,
		}
	}
}

func (c *ComponentCollection) App() *AppResource {
	return c.app
}

// List returns an ComponentList.
func (c *ComponentCollection) List() (*ComponentList, error) {
	list := new(ComponentList)
	err := c.core.db.list(c, list)
	return list, err
}

// New initializes an Component with a pointer to the Collection.
func (c *ComponentCollection) New() *ComponentResource {
	r := &ComponentResource{
		Component: &common.Component{
			Meta: common.NewMeta(),
		},
	}
	c.initializeResource(r)
	return r
}

// Create takes an Component and creates it in etcd.
func (c *ComponentCollection) Create(r *ComponentResource) error {
	return c.core.db.create(c, r.Name, r)
}

// Get takes a name and returns an ComponentResource if it exists.
func (c *ComponentCollection) Get(name common.ID) (*ComponentResource, error) {
	r := c.New()
	if err := c.core.db.get(c, name, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Update saves the Component in etcd through an update.
func (c *ComponentCollection) Update(name common.ID, r *ComponentResource) error {
	return c.core.db.update(c, name, r)
}

// Patch partially updates the App in etcd.
func (c *ComponentCollection) Patch(name common.ID, r *ComponentResource) error {
	return c.core.db.patch(c, name, r)
}

// Delete cascades delete calls to current and target releases, and deletes the
// Component in etcd.
//
// TODO this should somehow stop any ongoing tasks related to the Component.
func (c *ComponentCollection) Delete(ri Resource) error {
	r := ri.(*ComponentResource)

	releases, err := r.Releases().List()
	if err != nil {
		return err
	}

	// NOTE we delete releases concurrently, because when target and current both
	// exist (i.e. a deploy was still running), then one may hang waiting on the
	// other to release an asset like volumes.

	ch := make(chan error)
	for _, release := range releases.Items {
		go func(release *ReleaseResource) {
			ch <- release.Delete()
		}(release)
	}
	for i := 0; i < len(releases.Items); i++ {
		for err := <-ch; err != nil; {
			return err
		}
	}

	return c.core.db.delete(c, r.Name)
}

func (c *ComponentCollection) Deploy(ri Resource) (err error) {
	r := ri.(*ComponentResource)

	var currentRelease *ReleaseResource
	if r.CurrentReleaseTimestamp != nil {
		currentRelease, err = r.CurrentRelease()
		if err != nil {
			return err
		}
	}
	// There should always be a target release at this point
	targetRelease, err := r.TargetRelease()
	if err != nil {
		return err
	}

	// This sets up all the necessary dependencies (the only thing needed past the
	// first release is volumes for new instances)
	if err := targetRelease.Provision(); err != nil {
		return err
	}

	if currentRelease != nil {
		targetRelease.AddNewPorts(currentRelease)
	}

	if customDeploy := r.CustomDeployScript; customDeploy != nil {
		if err := RunCustomDeployment(c.core, r); err != nil {
			return err
		}
	} else {
		// This goes to the deploy/ folder which uses the client package.
		if err := deploy.Deploy(c.app.Name, r.Name); err != nil {
			return err
		}
	}

	// Make sure old release (current) has been fully stopped, and the new release
	// (target) has been fully started.
	// It doesn't matter on the first deploy, though.
	if currentRelease != nil && *currentRelease.InstanceGroup != *targetRelease.InstanceGroup {
		if !currentRelease.IsStopped() {
			return fmt.Errorf("Current Release for Component %s:%s is not completely stopped.", common.StringID(c.app.Name), common.StringID(r.Name))
		}
	}
	if !targetRelease.IsStarted() {
		return fmt.Errorf("Target Release for Component %s:%s is not completely started.", common.StringID(c.app.Name), common.StringID(r.Name))
	}

	// TODO really sloppy
	// Stopping instances doesn't remove volumes. So, user-defined deploys, when
	// removing instances, can't control the volumes, which need to be deleted.
	if currentRelease != nil && targetRelease.InstanceCount < currentRelease.InstanceCount {
		instancesRemoving := currentRelease.InstanceCount - targetRelease.InstanceCount
		instances := currentRelease.Instances().List().Items
		for _, instance := range instances[len(instances)-instancesRemoving:] { // TODO test that this works correctly
			instance.DeleteVolumes()
		}
	}

	if currentRelease != nil {
		targetRelease.RemoveOldPorts(currentRelease)

		currentRelease.Retired = true
		currentRelease.Update()
	}

	// If we're all good, we set target to current, and remove target.
	r.CurrentReleaseTimestamp = r.TargetReleaseTimestamp
	r.TargetReleaseTimestamp = nil
	return c.core.db.update(c, r.Name, r)
}

//------------------------------------------------------------------------------

// Key implements the Locatable interface.
func (c *ComponentCollection) locationKey() string {
	return "components"
}

// Parent implements the Locatable interface.
func (c *ComponentCollection) parent() Locatable {
	return c.app
}

// Child implements the Locatable interface.
func (c *ComponentCollection) child(key string) Locatable {
	r, err := c.Get(common.IDString(key))
	if err != nil {
		panic(fmt.Errorf("No child with key %s for %T", key, c))
	}
	return r
}

// Key implements the Locatable interface.
func (r *ComponentResource) locationKey() string {
	return common.StringID(r.Name)
}

// Parent implements the Locatable interface.
func (r *ComponentResource) parent() Locatable {
	return r.collection.(Locatable)
}

// Child implements the Locatable interface.
func (r *ComponentResource) child(key string) (l Locatable) {
	switch key {
	case "releases":
		l = r.Releases().(Locatable)
	default:
		panic(fmt.Errorf("No child with key %s for %T", key, r))
	}
	return
}

// Action implements the Resource interface.
func (r *ComponentResource) Action(name string) *Action {
	var fn ActionPerformer
	switch name {
	case "deploy":
		fn = ActionPerformer(r.collection.Deploy)
	case "delete":
		fn = ActionPerformer(r.collection.Delete)
	default:
		panic(fmt.Errorf("No action %s for Component", name))
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
func (r *ComponentResource) decorate() error {
	if r.CurrentReleaseTimestamp == nil {
		return nil
	}

	externalAddrs, err := r.externalAddresses()
	if err != nil {
		return err
	}
	internalAddrs, err := r.internalAddresses()
	if err != nil {
		return err
	}

	r.Addresses = &common.ComponentAddresses{
		External: externalAddrs,
		Internal: internalAddrs,
	}

	return nil
}

// Update is a proxy method to ComponentCollection's Update.
func (r *ComponentResource) Update() error {
	return r.collection.Update(r.Name, r)
}

// Patch is a proxy method to collection Patch.
func (r *ComponentResource) Patch() error {
	return r.collection.Patch(r.Name, r)
}

// Delete is a proxy method to ComponentCollection's Delete.
func (r *ComponentResource) Delete() error {
	return r.collection.Delete(r)
}

func (r *ComponentResource) App() *AppResource {
	return r.collection.App()
}

// Releases returns a ReleasesInterface with a pointer to the AppResource.
func (r *ComponentResource) Releases() ReleasesInterface {
	// TODO this is now just a getter
	return r.ReleasesInterface
}

func (r *ComponentResource) CurrentRelease() (*ReleaseResource, error) {
	return r.Releases().Get(r.CurrentReleaseTimestamp)
}

func (r *ComponentResource) TargetRelease() (*ReleaseResource, error) {
	return r.Releases().Get(r.TargetReleaseTimestamp)
}

func (r *ComponentResource) externalAddresses() (addrs []*common.PortAddress, err error) {
	release, err := r.CurrentRelease()
	if err != nil {
		return nil, err
	}
	for _, port := range release.ExternalPorts() {
		addrs = append(addrs, port.address())
	}
	return addrs, nil
}

func (r *ComponentResource) internalAddresses() (addrs []*common.PortAddress, err error) {
	release, err := r.CurrentRelease()
	if err != nil {
		return nil, err
	}
	for _, port := range release.InternalPorts() {
		addrs = append(addrs, port.address())
	}
	return addrs, nil
}
