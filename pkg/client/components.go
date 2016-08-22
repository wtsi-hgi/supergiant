package client

import "github.com/supergiant/supergiant/pkg/models"

type Components struct {
	Collection
}

func (c *Components) Deploy(m *models.Component) error {
	return c.client.request("POST", c.memberPath(m.ID)+"/deploy", nil, m, nil)
}
