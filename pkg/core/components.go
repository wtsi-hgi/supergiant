package core

import (
	"errors"
	"fmt"

	"github.com/supergiant/supergiant/pkg/deploy"
	"github.com/supergiant/supergiant/pkg/models"
)

type Components struct {
	Collection
}

func (c *Components) Deploy(id *int64, m *models.Component) *Action {
	return &Action{
		Status: &models.ActionStatus{
			Description: "deploying",
			MaxRetries:  5,
		},
		core:  c.core,
		scope: c.core.DB.Preload("App.Kube.CloudAccount").Preload("PrivateImageKeys.Key").Preload("CurrentRelease").Preload("TargetRelease").Preload("Instances"),
		model: m,
		id:    id,
		fn: func(_ *Action) error {

			// TODO something is causing Kube to not exist by time it gets to
			// serviceSet provision, there's a note below.
			kube := *m.App.Kube

			// TODO this should not be handled async
			if m.TargetRelease == nil {
				return errors.New("Component does not have target Release")
			}

			m.TargetRelease.InUse = true
			if err := c.core.DB.Save(m.TargetRelease); err != nil {
				return err
			}

			// Create Instances
			for n := 0; n < m.InstanceCount(); n++ {
				if m.InstanceByNum(n) != nil {
					continue
				}

				instance := &models.Instance{
					Component:   m,
					ComponentID: m.ID,
					Release:     m.TargetRelease,
					ReleaseID:   m.TargetReleaseID,
					Num:         n,
					Name:        fmt.Sprintf("%s-%d", m.Name, n),
				}
				if err := c.core.Instances.Create(instance); err != nil {
					return err
				}
				m.Instances = append(m.Instances, instance)
			}

			// Create Volumes in parallel
			if m.TargetRelease.Config.Volumes != nil {
				err := c.core.Instances.inParallel(m.Instances, func(mi interface{}) error {
					instance := mi.(*models.Instance)
					return c.core.Instances.CreateVolumes(instance.ID, instance)
				})
				if err != nil {
					return err
				}
			}

			// TODO see above.... something unsets Preloaded Kube from App
			m.App.Kube = &kube

			// Provision Secrets
			for _, imageKey := range m.PrivateImageKeys {
				if err := provisionSecret(c.core, m.App, imageKey.Key); err != nil {
					return err
				}
			}

			// Provision Services
			serviceSet, err := c.serviceSet(m)
			if err != nil {
				return err
			}
			if err := serviceSet.provision(); err != nil {
				return err
			}

			// Add new ports to existing service, if there is one, and there are any.
			if m.CurrentRelease != nil {
				if err := serviceSet.addNewPorts(); err != nil {
					return err
				}
			}

			// Run "inner" deployment
			if m.CustomDeployScript != nil {
				if err := RunCustomDeployment(c.core, m); err != nil {
					return err
				}
			} else {
				// This goes to the deploy/ folder which uses the client package.
				if err := deploy.Deploy(c.core.NewAPIClient(), m.ID); err != nil {
					return err
				}
			}

			// Reload Instances
			if err := c.core.DB.Where("component_id = ?", m.ID).Find(&m.Instances); err != nil {
				return err
			}

			// Make sure all Instances (that haven't been deleted) have been restarted
			for _, instance := range m.Instances {
				if instance.Num <= m.TargetRelease.InstanceCount && *instance.ReleaseID != *m.TargetReleaseID {
					return fmt.Errorf("Not all Instances for Component %d have been started with the target Release", m.ID)
				}
			}

			if m.CurrentRelease != nil {
				// Remove old ports from service if there are any
				if err := serviceSet.removeOldPorts(); err != nil {
					return err
				}

				// Mark old Release as retired
				m.CurrentRelease.InUse = false
				if err := c.core.DB.Save(m.CurrentRelease); err != nil {
					return err
				}
			}

			// Save addresses to Component
			externalAddrs, err := serviceSet.externalAddresses()
			if err != nil {
				return err
			}
			internalAddrs, err := serviceSet.internalAddresses()
			if err != nil {
				return err
			}
			m.Addresses = &models.Addresses{
				External: externalAddrs,
				Internal: internalAddrs,
			}

			// If we're all good, we set target to current, and remove target.
			m.CurrentRelease = m.TargetRelease
			m.CurrentReleaseID = m.TargetReleaseID
			m.TargetRelease = nil
			m.TargetReleaseID = nil
			return c.core.DB.Save(m)
		},
	}
}

func (c *Components) Delete(id *int64, m *models.Component) *Action {
	return &Action{
		Status: &models.ActionStatus{
			Description: "deleting",
			MaxRetries:  20,
		},
		core:           c.core,
		scope:          c.core.DB.Preload("App.Kube.CloudAccount").Preload("PrivateImageKeys").Preload("TargetRelease").Preload("CurrentRelease").Preload("Releases").Preload("Instances"),
		model:          m,
		id:             id,
		cancelExisting: true,
		fn: func(action *Action) error {
			// Delete Instances
			err := c.inParallel(m.Instances, func(mi interface{}) error {
				instance := mi.(*models.Instance)
				return c.core.Instances.Delete(instance.ID, instance).Now()
			})
			if err != nil {
				return err
			}
			// Delete shared services
			if m.TargetReleaseID != nil || m.CurrentReleaseID != nil {
				serviceSet, err := c.serviceSet(m)
				if err != nil {
					return err
				}
				if serviceSet.delete(); err != nil {
					return err
				}
			}
			// Delete Releases
			for _, release := range m.Releases {
				if err := c.core.DB.Delete(release); err != nil {
					return err
				}
			}
			// Delete ComponentPrivateImageKeys (many2many)
			for _, keyAssoc := range m.PrivateImageKeys {
				if err := c.core.DB.Delete(keyAssoc); err != nil {
					return err
				}
			}
			// Delete self
			return c.core.DB.Delete(m)
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
// Private methods                                                            //
////////////////////////////////////////////////////////////////////////////////

func (c *Components) serviceSet(m *models.Component) (*ServiceSet, error) {
	return NewServiceSet(c.core, m, m.TargetRelease, m.Name, map[string]string{"service": m.Name}, nil)
}
