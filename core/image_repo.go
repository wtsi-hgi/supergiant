package core

import (
	"fmt"

	"github.com/supergiant/supergiant/common"
)

type ImageReposInterface interface {
	List() (*ImageRepoList, error)
	New() *ImageRepoResource
	Create(*ImageRepoResource) error
	Get(common.ID) (*ImageRepoResource, error)
	Update(common.ID, *ImageRepoResource) error
	Patch(common.ID, *ImageRepoResource) error
	Delete(*ImageRepoResource) error
}

type ImageRepoCollection struct {
	core     *Core
	registry *ImageRegistryResource
}

type ImageRepoResource struct {
	core       *Core
	collection ImageReposInterface
	*common.ImageRepo
}

// NOTE this does not inherit from common like model does; all we need is a List
// object, internally, that has a slice of our composed model above.
type ImageRepoList struct {
	Items []*ImageRepoResource `json:"items"`
}

// initializeResource implements the Collection interface.
func (c *ImageRepoCollection) initializeResource(in Resource) {
	r := in.(*ImageRepoResource)
	r.collection = c
	r.core = c.core
}

// List returns an ImageRepoList.
func (c *ImageRepoCollection) List() (*ImageRepoList, error) {
	list := new(ImageRepoList)
	err := c.core.db.list(c, list)
	return list, err
}

// New initializes an ImageRepo with a pointer to the Collection.
func (c *ImageRepoCollection) New() *ImageRepoResource {
	r := &ImageRepoResource{
		ImageRepo: &common.ImageRepo{
			Meta: common.NewMeta(),
		},
	}
	c.initializeResource(r)
	return r
}

// Create takes an ImageRepo and creates it in etcd.
func (c *ImageRepoCollection) Create(r *ImageRepoResource) error {
	return c.core.db.create(c, r.Name, r)
}

// Get takes a name and returns an ImageRepoResource if it exists.
func (c *ImageRepoCollection) Get(name common.ID) (*ImageRepoResource, error) {
	r := c.New()
	if err := c.core.db.get(c, name, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Update updates the ImageRepo in etcd.
func (c *ImageRepoCollection) Update(name common.ID, r *ImageRepoResource) error {
	return c.core.db.update(c, name, r)
}

// Patch partially updates the App in etcd.
func (c *ImageRepoCollection) Patch(name common.ID, r *ImageRepoResource) error {
	return c.core.db.patch(c, name, r)
}

// Delete deletes the ImageRepo in etcd.
func (c *ImageRepoCollection) Delete(r *ImageRepoResource) error {
	return c.core.db.delete(c, r.Name)
}

//------------------------------------------------------------------------------

// Key implements the Locatable interface.
func (c *ImageRepoCollection) locationKey() string {
	return "repos"
}

// Parent implements the Locatable interface.
func (c *ImageRepoCollection) parent() Locatable {
	return c.registry
}

// Child implements the Locatable interface.
func (c *ImageRepoCollection) child(key string) Locatable {
	r, err := c.Get(common.IDString(key))
	if err != nil {
		panic(fmt.Errorf("No child with key %s for %T", key, c))
	}
	return r
}

// Key implements the Locatable interface.
func (r *ImageRepoResource) locationKey() string {
	return common.StringID(r.Name)
}

// Parent implements the Locatable interface.
func (r *ImageRepoResource) parent() Locatable {
	return r.collection.(Locatable)
}

// Child implements the Locatable interface.
func (r *ImageRepoResource) child(key string) (l Locatable) {
	switch key {
	default:
		panic(fmt.Errorf("No child with key %s for %T", key, r))
	}
}

// Action implements the Resource interface.
func (r *ImageRepoResource) Action(name string) *Action {
	// var fn ActionPerformer
	switch name {
	default:
		panic(fmt.Errorf("No action %s for ImageRepo", name))
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
func (r *ImageRepoResource) decorate() (err error) {
	return
}

// Update is a proxy method to ImageRepoCollection's Update.
func (r *ImageRepoResource) Update() error {
	return r.collection.Update(r.Name, r)
}

// Patch is a proxy method to collection Patch.
func (r *ImageRepoResource) Patch() error {
	return r.collection.Patch(r.Name, r)
}

// Delete is a proxy method to ImageRepoCollection's Delete.
func (r *ImageRepoResource) Delete() error {
	return r.collection.Delete(r)
}
