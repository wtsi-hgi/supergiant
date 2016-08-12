package core

import (
	"errors"

	"github.com/imdario/mergo"
	"github.com/supergiant/supergiant/pkg/models"
)

type Releases struct {
	Collection
}

func (c *Releases) Create(m *models.Release) error {
	// load Component (you can't have it preloaded here)
	m.Component = new(models.Component)
	if err := c.core.DB.Preload("CurrentRelease").First(m.Component, *m.ComponentID); err != nil {
		return err
	}

	if m.Component.TargetReleaseID != nil {
		return errors.New("Component already has a target Release")
	}

	// If there's a current Release, merge its attributes into this new Release
	if m.Component.CurrentRelease != nil {
		if err := mergo.Merge(m, m.Component.CurrentRelease); err != nil {
			return err
		}
		m.ID = nil
		m.InstanceGroup = nil
		m.InUse = false
	}

	if m.InstanceGroup == nil {
		m.InstanceGroup = m.ID
	} else if *m.InstanceGroup != *m.ID && *m.InstanceGroup != *m.Component.CurrentReleaseID {
		return errors.New("Release InstanceGroup field can only be set to either the current or target Release's Timestamp value.")
	}

	if err := c.Collection.Create(m); err != nil {
		return err
	}

	m.Component.TargetReleaseID = m.ID
	return c.core.DB.Save(m.Component)
}

// TODO prevent updating / deleting active release in controller
