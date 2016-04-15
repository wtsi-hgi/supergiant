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

	// Relations
	ComponentsInterface ComponentsInterface
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
func (c *AppCollection) initializeResource(in Resource) error {
	r := in.(*AppResource)
	r.collection = c
	r.ComponentsInterface = &ComponentCollection{
		core: c.core,
		app:  r,
	}
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
	r := &AppResource{
		App: &common.App{
			Meta: common.NewMeta(),
		},
	}
	c.initializeResource(r)
	return r
}

// Create takes an App and creates it in etcd. It also creates a Kubernetes
// Namespace with the name of the App.
func (c *AppCollection) Create(r *AppResource) error {
	if err := c.core.db.create(c, r.Name, r); err != nil {
		return err
	}
	// TODO for error handling and retries, we may want to do this in a task and
	// utilize a Status field
	if err := r.createNamespace(); err != nil {
		return err
	}
	return nil
}

// Get takes a name and returns an AppResource if it exists.
func (c *AppCollection) Get(name common.ID) (*AppResource, error) {
	r := c.New()
	if err := c.core.db.get(c, name, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Update updates the App in etcd.
func (c *AppCollection) Update(name common.ID, r *AppResource) error {
	return c.core.db.update(c, name, r)
}

// Delete deletes the App in etcd, and deletes the namespace and all Components.
//
// NOTE I know it's weird that we take an App here and not an ID, but we do that
// to prevent double lookup when component.Delete() is called. And while it may
// be weird, the resource-level Delete() is the natural approach to calling
// Delete, and so it would be rare to have to manually say
// components.Delete(component). The reason we put this logic here and not
// on the Component itself is because we want to isolate shared Resource
// behavior (CRUD) from Resources, preventing Resources from having any
// operational logic (like deleting volumes and such) that we want to mock.
// It is difficult to approach mocking Resources, because if they are returned
// as interfaces from methods like collection.Get(), we no longer have access
// to the attributes of the Resource without type casting.
func (c *AppCollection) Delete(r *AppResource) error {
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
	return c.core.db.delete(c, r.Name)
}

// Resource-level

// Update is a proxy method to AppCollection's Update.
func (r *AppResource) Update() error {
	return r.collection.Update(r.Name, r)
}

// Delete is a proxy method to AppCollection's Delete.
func (r *AppResource) Delete() error {
	return r.collection.Delete(r)
}

// Components returns a ComponentsInterface with a pointer to the AppResource.
func (r *AppResource) Components() ComponentsInterface {
	// TODO this is now just a getter
	return r.ComponentsInterface
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
