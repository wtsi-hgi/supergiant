package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/supergiant/supergiant/pkg/models"
	"github.com/supergiant/supergiant/pkg/util"
)

type Instances struct {
	Collection
}

// func (c *Instances) Log(m *models.Instance) error {
// 	return c.client.request("Get", c.memberPath(m.ID)+"/log", nil, m, nil) --------------------- can't decode JSON to string
// }

func (c *Instances) Start(m *models.Instance) error {
	return c.client.request("POST", c.memberPath(m.ID)+"/start", nil, m, nil)
}

func (c *Instances) Stop(m *models.Instance) error {
	return c.client.request("POST", c.memberPath(m.ID)+"/stop", nil, m, nil)
}

func (c *Instances) WaitForStarted(m *models.Instance) error {
	return util.WaitFor(fmt.Sprintf("Instance %d to start", m.ID), 4*time.Hour, 3*time.Second, func() (bool, error) {
		if err := c.Get(m.ID, m); err != nil {
			return false, err
		}

		// NOTE this is essential for cancelling deploys on Component delete, the
		// reason being that this is generally the longest part of deploys, because
		// instance Start() is what resizes volumes, which can take forever.
		if m.Status != nil && m.Status.Cancelled {
			return false, errors.New("Instance start was cancelled")
		}

		return m.Started, nil
	})
}

func (c *Instances) WaitForStopped(m *models.Instance) error {
	return util.WaitFor(fmt.Sprintf("Instance %d to stop", m.ID), 5*time.Minute, 3*time.Second, func() (bool, error) {
		if err := c.Get(m.ID, m); err != nil {
			return false, err
		}
		return !m.Started, nil
	})
}

func (c *Instances) WaitForDeleted(m *models.Instance) error {
	return util.WaitFor(fmt.Sprintf("Instance %d to terminate", m.ID), 5*time.Minute, 3*time.Second, func() (bool, error) {
		err := c.Get(m.ID, m)
		return err != nil, nil // TODO ------- check for 404
	})
}
