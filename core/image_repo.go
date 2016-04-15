package core

import (
	"path"

	"github.com/supergiant/supergiant/common"
)

type ImageRepoCollection struct {
	core *Core
}

type ImageRepoResource struct {
	core       *Core
	collection *ImageRepoCollection
	*common.ImageRepo
}

// NOTE this does not inherit from common like model does; all we need is a List
// object, internally, that has a slice of our composed model above.
type ImageRepoList struct {
	Items []*ImageRepoResource `json:"items"`
}

// etcdKey implements the Collection interface.
func (c *ImageRepoCollection) etcdKey(name common.ID) string {
	key := "/image_repos/dockerhub"
	if name != nil {
		key = path.Join(key, common.StringID(name))
	}
	return key
}

// initializeResource implements the Collection interface.
func (c *ImageRepoCollection) initializeResource(in Resource) error {
	r := in.(*ImageRepoResource)
	r.collection = c
	r.core = c.core
	return nil
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
		collection: c,
		ImageRepo: &common.ImageRepo{
			Meta: common.NewMeta(),
		},
	}
	c.initializeResource(r)
	return r
}

// Create takes an ImageRepo and creates it in etcd. It also creates a Kubernetes
// Namespace with the name of the ImageRepo.
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

// Delete deletes the ImageRepo in etcd.
func (c *ImageRepoCollection) Delete(r *ImageRepoResource) error {
	return c.core.db.delete(c, r.Name)
}

// Resource-level

// Update is a proxy method to ImageRepoCollection's Update.
func (r *ImageRepoResource) Update() error {
	return r.collection.Update(r.Name, r)
}

// Delete is a proxy method to ImageRepoCollection's Delete.
func (r *ImageRepoResource) Delete() error {
	return r.collection.Delete(r)
}
