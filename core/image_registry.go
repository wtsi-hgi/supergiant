package core

import "github.com/supergiant/supergiant/common"

type ImageRegistriesInterface interface {
	List() (*ImageRegistryList, error)
	New() *ImageRegistryResource
	Create(*ImageRegistryResource) error
	Get(common.ID) (*ImageRegistryResource, error)
	Update(common.ID, *ImageRegistryResource) error
	Delete(*ImageRegistryResource) error
}

type ImageRegistryCollection struct {
	core *Core
}

type ImageRegistryResource struct {
	core       *Core
	collection ImageRegistriesInterface
	*common.ImageRegistry

	// Relations
	ImageReposInterface ImageReposInterface `json:"-"`
}

// NOTE this does not inherit from common like model does; all we need is a List
// object, internally, that has a slice of our composed model above.
type ImageRegistryList struct {
	Items []*ImageRegistryResource `json:"items"`
}

// initializeResource implements the Collection interface.
func (c *ImageRegistryCollection) initializeResource(in Resource) {
	r := in.(*ImageRegistryResource)
	r.collection = c
	r.core = c.core
	r.ImageReposInterface = &ImageRepoCollection{
		core:     c.core,
		registry: r,
	}
}

// List returns an ImageRegistryList.
func (c *ImageRegistryCollection) List() (*ImageRegistryList, error) {
	list := new(ImageRegistryList)
	err := c.core.db.list(c, list)
	return list, err
}

// New initializes an ImageRegistry with a pointer to the Collection.
func (c *ImageRegistryCollection) New() *ImageRegistryResource {
	r := &ImageRegistryResource{
		ImageRegistry: &common.ImageRegistry{
			Meta: common.NewMeta(),
		},
	}
	c.initializeResource(r)
	return r
}

// Create takes an ImageRegistry and creates it in etcd.
func (c *ImageRegistryCollection) Create(r *ImageRegistryResource) error {
	return c.core.db.create(c, r.Name, r)
}

// Get takes a name and returns an ImageRegistryResource if it exists.
func (c *ImageRegistryCollection) Get(name common.ID) (*ImageRegistryResource, error) {
	r := c.New()
	if err := c.core.db.get(c, name, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Update updates the ImageRegistry in etcd.
func (c *ImageRegistryCollection) Update(name common.ID, r *ImageRegistryResource) error {
	return c.core.db.update(c, name, r)
}

// Delete deletes the ImageRegistry in etcd.
func (c *ImageRegistryCollection) Delete(r *ImageRegistryResource) error {
	return c.core.db.delete(c, r.Name)
}

//------------------------------------------------------------------------------

// Key implements the Locatable interface.
func (c *ImageRegistryCollection) locationKey() string {
	return "registries"
}

// Parent implements the Locatable interface.
func (c *ImageRegistryCollection) parent() (l Locatable) {
	return
}

// Child implements the Locatable interface.
func (c *ImageRegistryCollection) child(key string) Locatable {
	r, err := c.Get(common.IDString(key))
	if err != nil {
		Log.Panicf("No child with key %s for %T", key, c)
	}
	return r
}

// Key implements the Locatable interface.
func (r *ImageRegistryResource) locationKey() string {
	return common.StringID(r.Name)
}

// Parent implements the Locatable interface.
func (r *ImageRegistryResource) parent() Locatable {
	return r.collection.(Locatable)
}

// Child implements the Locatable interface.
func (r *ImageRegistryResource) child(key string) (l Locatable) {
	switch key {
	case "repos":
		l = r.ImageRepos().(Locatable)
	default:
		Log.Panicf("No child with key %s for %T", key, r)
	}
	return
}

// Action implements the Resource interface.
func (r *ImageRegistryResource) Action(name string) *Action {
	var fn ActionPerformer
	switch name {
	default:
		Log.Panicf("No action %s for ImageRegistry", name)
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
func (r *ImageRegistryResource) decorate() (err error) {
	return
}

// Update is a proxy method to ImageRegistryCollection's Update.
func (r *ImageRegistryResource) Update() error {
	return r.collection.Update(r.Name, r)
}

// Delete is a proxy method to ImageRegistryCollection's Delete.
func (r *ImageRegistryResource) Delete() error {
	return r.collection.Delete(r)
}

// ImageRepos returns a ImageReposInterface with a pointer to the ImageRegistryResource.
func (r *ImageRegistryResource) ImageRepos() ImageReposInterface {
	// TODO this is now just a getter
	return r.ImageReposInterface
}
