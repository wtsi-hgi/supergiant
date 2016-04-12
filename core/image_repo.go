package core

import (
	"path"

	"github.com/supergiant/supergiant/common"
)

type ImageRepoCollection struct {
	core *Core
}

type ImageRepoResource struct {
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
func (c *ImageRepoCollection) initializeResource(r Resource) error {
	resource := r.(*ImageRepoResource)
	resource.collection = c
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
	return &ImageRepoResource{
		collection: c,
		ImageRepo: &common.ImageRepo{
			Meta: common.NewMeta(),
		},
	}
}

// Create takes an ImageRepo and creates it in etcd. It also creates a Kubernetes
// Namespace with the name of the ImageRepo.
func (c *ImageRepoCollection) Create(r *ImageRepoResource) (*ImageRepoResource, error) {
	if err := c.core.db.create(c, r.Name, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Get takes a name and returns an ImageRepoResource if it exists.
func (c *ImageRepoCollection) Get(name common.ID) (*ImageRepoResource, error) {
	r := c.New()
	if err := c.core.db.get(c, name, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Resource-level
//==============================================================================

// Delete deletes the ImageRepo in etcd.
func (r *ImageRepoResource) Delete() error {
	return r.collection.core.db.delete(r.collection, r.Name)
}
