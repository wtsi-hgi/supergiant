package api

import (
	"net/http"

	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/models"
)

func ListReleases(core *core.Core, r *http.Request) (*Response, error) {
	return handleList(core, r, new(models.Release))
}

func CreateRelease(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.Release)
	if err := decodeBodyInto(r, item); err != nil {
		return nil, err
	}
	if err := core.Releases.Create(item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusCreated)
}

func UpdateRelease(core *core.Core, r *http.Request) (*Response, error) {
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	item := new(models.Release)
	if err := decodeBodyInto(r, item); err != nil {
		return nil, err
	}
	if err := core.Releases.Update(id, new(models.Release), item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusAccepted)
}

func GetRelease(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.Release)
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	if err := core.Releases.Get(id, item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusOK)
}

func DeleteRelease(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.Release)
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	if err := core.Releases.Delete(id, item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusAccepted)
}
