package api

import (
	"net/http"

	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/models"
)

func ListApps(core *core.Core, r *http.Request) (*Response, error) {
	return handleList(core, r, new(models.App))
}

func CreateApp(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.App)
	if err := decodeBodyInto(r, item); err != nil {
		return nil, err
	}
	if err := core.Apps.Create(item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusCreated)
}

func UpdateApp(core *core.Core, r *http.Request) (*Response, error) {
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	item := new(models.App)
	if err := decodeBodyInto(r, item); err != nil {
		return nil, err
	}
	if err := core.Apps.Update(id, new(models.App), item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusAccepted)
}

func GetApp(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.App)
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	if err := core.Apps.Get(id, item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusOK)
}

func DeleteApp(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.App)
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	if err := core.Apps.Delete(id, item).Async(); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusAccepted)
}
