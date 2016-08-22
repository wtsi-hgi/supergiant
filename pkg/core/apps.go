package core

import (
	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/pkg/models"
)

type Apps struct {
	Collection
}

func (c *Apps) Create(m *models.App) error {
	if err := c.Collection.Create(m); err != nil {
		return err
	}

	provision := &Action{
		Status: &models.ActionStatus{
			Description: "provisioning",
			MaxRetries:  5,
		},
		core:       c.core,
		resourceID: m.UUID,
		model:      m,
		fn: func(_ *Action) error {
			return c.createNamespace(m)
		},
	}
	return provision.Async()
}

func (c *Apps) Delete(id *int64, m *models.App) *Action {
	return &Action{
		Status: &models.ActionStatus{
			Description: "deleting",
			MaxRetries:  20,
		},
		core:  c.core,
		scope: c.core.DB.Preload("Kube.CloudAccount").Preload("Components"),
		model: m,
		id:    id,
		fn: func(_ *Action) error {
			for _, component := range m.Components {
				if err := c.core.Components.Delete(component.ID, component).Now(); err != nil {
					return err
				}
			}
			// TODO
			// Ideally we would delete namespace first, because it quickly tears down all
			// kube resources. However, in order to remove ports from ELBs, we currently
			// need the K8S service to stick around so we know the assigned NodePort
			// value. A solution may be simply to store the port assignment.
			if err := c.deleteNamespace(m); err != nil {
				return err
			}
			return c.Collection.Delete(id, m)
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
// Private methods                                                            //
////////////////////////////////////////////////////////////////////////////////

func (c *Apps) createNamespace(m *models.App) error {
	namespace := &guber.Namespace{
		Metadata: &guber.Metadata{
			Name: m.Name,
		},
	}
	_, err := c.core.K8S(m.Kube).Namespaces().Create(namespace)
	if err != nil && !isKubeAlreadyExistsErr(err) {
		return err
	}
	return nil
}

func (c *Apps) deleteNamespace(m *models.App) error {
	err := c.core.K8S(m.Kube).Namespaces().Delete(m.Name)
	if err != nil && !isKubeNotFoundErr(err) {
		return err
	}
	return nil
}
