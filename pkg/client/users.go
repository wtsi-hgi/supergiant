package client

import "github.com/supergiant/supergiant/pkg/model"

type Users struct {
	Collection
}

func (c *Users) RegenerateAPIToken(id interface{}, m *model.User) error {
	return c.client.request("POST", c.memberPath(id)+"/regenerate_api_token", nil, m, nil)
}
