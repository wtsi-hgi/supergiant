package api

import (
	"net/http"

	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/models"
)

func ListKubes(core *core.Core, r *http.Request) (*Response, error) {
	return handleList(core, r, new(models.Kube))
}

func CreateKube(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.Kube)
	if err := decodeBodyInto(r, item); err != nil {
		return nil, err
	}
	if err := core.Kubes.Create(item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusCreated)
}

func UpdateKube(core *core.Core, r *http.Request) (*Response, error) {
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	item := new(models.Kube)
	if err := decodeBodyInto(r, item); err != nil {
		return nil, err
	}
	if err := core.Kubes.Update(id, new(models.Kube), item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusAccepted)
}

func GetKube(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.Kube)
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	if err := core.Kubes.Get(id, item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusOK)
}

func DeleteKube(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.Kube)
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	if err := core.Kubes.Delete(id, item).Async(); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusAccepted)
}
