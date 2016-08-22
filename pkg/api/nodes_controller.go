package api

import (
	"net/http"

	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/models"
)

func ListNodes(core *core.Core, r *http.Request) (*Response, error) {
	return handleList(core, r, new(models.Node))
}

func CreateNode(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.Node)
	if err := decodeBodyInto(r, item); err != nil {
		return nil, err
	}
	if err := core.Nodes.Create(item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusCreated)
}

func UpdateNode(core *core.Core, r *http.Request) (*Response, error) {
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	item := new(models.Node)
	if err := decodeBodyInto(r, item); err != nil {
		return nil, err
	}
	if err := core.Nodes.Update(id, new(models.Node), item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusAccepted)
}

func GetNode(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.Node)
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	if err := core.Nodes.Get(id, item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusOK)
}

func DeleteNode(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.Node)
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	if err := core.Nodes.Delete(id, item).Async(); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusAccepted)
}
