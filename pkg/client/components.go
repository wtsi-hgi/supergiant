package client

import "github.com/supergiant/supergiant/pkg/model"

type Components struct {
	Collection
}

func (c *Components) Deploy(m *model.Component) error {
	return c.client.request("POST", c.memberPath(m.ID)+"/deploy", nil, m, nil)
}
