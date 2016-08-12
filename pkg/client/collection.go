package client

import (
	"fmt"
	"strings"

	"github.com/supergiant/supergiant/pkg/models"
)

type Collection struct {
	client   *Client
	basePath string
}

func (c *Collection) List(items interface{}) error {
	return c.client.request("GET", c.basePath, nil, items, nil)
}

func (c *Collection) Get(id *int64, item models.Model) error {
	return c.client.request("GET", c.memberPath(id), nil, item, nil)
}

func (c *Collection) GetWithIncludes(id *int64, item models.Model, includes []string) error {
	queryValues := map[string]string{"includes": strings.Join(includes, " ")}
	return c.client.request("GET", c.memberPath(id), nil, item, queryValues)
}

func (c *Collection) Create(item models.Model) error {
	return c.client.request("POST", c.basePath, item, item, nil)
}

func (c *Collection) Update(item models.Model) error {
	return c.client.request("PATCH", c.memberPath(item.GetID()), item, item, nil)
}

func (c *Collection) Delete(item models.Model) error {
	return c.client.request("DELETE", c.memberPath(item.GetID()), item, item, nil)
}

////////////////////////////////////////////////////////////////////////////////
// Private methods                                                            //
////////////////////////////////////////////////////////////////////////////////

func (c *Collection) memberPath(id *int64) string {
	return fmt.Sprintf("%s/%d", c.basePath, *id)
}
