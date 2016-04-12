package core

import (
	"path"

	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"
)

type AppCollection struct {
	core *Core
}

type AppResource struct {
	collection *AppCollection
	*common.App
}

// NOTE this does not inherit from common like model does; all we need is a List
// object, internally, that has a slice of our composed model above.
type AppList struct {
	Items []*AppResource `json:"items"`
}

// etcdKey implements the Collection interface.
func (c *AppCollection) etcdKey(name common.ID) string {
	if name == nil {
		return "/apps"
	}
	return path.Join("/apps", common.StringID(name))
}

// initializeResource implements the Collection interface.
func (c *AppCollection) initializeResource(r Resource) error {
	resource := r.(*AppResource)
	resource.collection = c
	return nil
}

// List returns an AppList.
func (c *AppCollection) List() (*AppList, error) {
	list := new(AppList)
	err := c.core.db.list(c, list)
	return list, err
}

// New initializes an App with a pointer to the Collection.
func (c *AppCollection) New() *AppResource {
	return &AppResource{
		App: &common.App{
			Meta: common.NewMeta(),
		},
	}
}

// Create takes an App and creates it in etcd. It also creates a Kubernetes
// Namespace with the name of the App.
func (c *AppCollection) Create(r *AppResource) (*AppResource, error) {
	if err := c.core.db.create(c, r.Name, r); err != nil {
		return nil, err
	}

	// TODO for error handling and retries, we may want to do this in a task and
	// utilize a Status field
	if err := r.createNamespace(); err != nil {
		return nil, err
	}
	return r, nil
}

// Get takes a name and returns an AppResource if it exists.
func (c *AppCollection) Get(name common.ID) (*AppResource, error) {
	r := c.New()
	if err := c.core.db.get(c, name, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Resource-level

// Save saves the App in etcd through an update.
func (r *AppResource) Save() error {
	return r.collection.core.db.update(r.collection, r.Name, r)
}

// Delete cascades deletes to all Components, deletes the Kube Namespace, and
// deletes the App in etcd.
func (r *AppResource) Delete() error {
	components, err := r.Components().List()
	if err != nil {
		return err
	}
	if err := r.deleteNamespace(); err != nil {
		return err
	}
	for _, component := range components.Items {
		if err := component.Delete(); err != nil {
			return err
		}
	}
	return r.collection.core.db.delete(r.collection, r.Name)
}

// Components returns a ComponentCollection with a pointer to the AppResource.
func (r *AppResource) Components() *ComponentCollection {
	return &ComponentCollection{
		core: r.collection.core,
		App:  r,
	}
}

func (r *AppResource) createNamespace() error {
	namespace := &guber.Namespace{
		Metadata: &guber.Metadata{
			Name: common.StringID(r.Name),
		},
	}
	_, err := r.collection.core.k8s.Namespaces().Create(namespace)
	return err
}

func (r *AppResource) deleteNamespace() error {
	_, err := r.collection.core.k8s.Namespaces().Delete(common.StringID(r.Name))
	return err
}

func (r *AppResource) provisionSecret(repo *ImageRepoResource) error {
	// TODO not sure i've been consistent with error handling -- this strategy is
	// useful when there could be multiple common of errors, alongside the
	// expectation of an error when something doesn't exist
	secret, err := r.collection.core.k8s.Secrets(common.StringID(r.Name)).Get(common.StringID(repo.Name))

	if err != nil {
		return err
	} else if secret != nil {
		return nil
	}
	_, err = r.collection.core.k8s.Secrets(common.StringID(r.Name)).Create(asKubeSecret(repo))
	return err
}
