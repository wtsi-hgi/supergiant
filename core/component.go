package core

import (
	"path"

	"github.com/supergiant/supergiant/common"
)

type ComponentsInterface interface {

	// A simple getter since App is an attribute on actual Collections implementing this
	App() *AppResource

	List() (*ComponentList, error)
	New() *ComponentResource
	Create(*ComponentResource) error
	Get(common.ID) (*ComponentResource, error)
	Update(common.ID, *ComponentResource) error
	Delete(*ComponentResource) error
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
}

type ComponentList struct {
	Items []*ComponentResource `json:"items"`
}

// etcdKey implements the Collection interface.
func (c *ComponentCollection) etcdKey(name common.ID) string {
	key := path.Join("/components", common.StringID(c.App().Name))
	if name != nil {
		key = path.Join(key, common.StringID(name))
	}
	return key
}

// initializeResource implements the Collection interface.
func (c *ComponentCollection) initializeResource(r Resource) error {
	resource := r.(*ComponentResource)
	resource.collection = c
	resource.core = c.core

	// TODO it seems wrong this is called here -- execessive to have to load the
	// current Release, Entrypoints, and Kube Services just to render a
	// Component.
	// However, it's rare a Component is loaded out of the context of its
	// Release. We will change this when we see issues.
	return resource.decorate()
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

func (c *ComponentCollection) Update(name common.ID, r *ComponentResource) error {
	return c.core.db.update(c, name, r)
}

func (c *ComponentCollection) Delete(r *ComponentResource) error {
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
	close(ch)
	for err := <-ch; err != nil; {
		return err
	}

	return c.core.db.delete(c, r.Name)
}

// Resource-level
//==============================================================================

// Update saves the Component in etcd through an update.
func (r *ComponentResource) Update() error {
	return r.collection.Update(r.Name, r)
}

// Delete cascades delete calls to current and target releases, and deletes the
// Component in etcd.
//
// TODO this should somehow stop any ongoing tasks related to the Component.
func (r *ComponentResource) Delete() error {
	return r.collection.Delete(r)
}

func (r *ComponentResource) App() *AppResource {
	return r.collection.App()
}

func (r *ComponentResource) Releases() *ReleaseCollection {
	return &ReleaseCollection{
		core:      r.core,
		Component: r,
	}
}

func (r *ComponentResource) CurrentRelease() (*ReleaseResource, error) {
	return r.Releases().Get(r.CurrentReleaseTimestamp)
}

func (r *ComponentResource) TargetRelease() (*ReleaseResource, error) {
	return r.Releases().Get(r.TargetReleaseTimestamp)
}

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
