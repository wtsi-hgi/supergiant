package api

import (
	"net/http"

	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/models"
)

func ListPrivateImageKeys(core *core.Core, r *http.Request) (*Response, error) {
	return handleList(core, r, new(models.PrivateImageKey))
}

func CreatePrivateImageKey(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.PrivateImageKey)
	if err := decodeBodyInto(r, item); err != nil {
		return nil, err
	}
	if err := core.PrivateImageKeys.Create(item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusCreated)
}

func UpdatePrivateImageKey(core *core.Core, r *http.Request) (*Response, error) {
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	item := new(models.PrivateImageKey)
	if err := decodeBodyInto(r, item); err != nil {
		return nil, err
	}
	if err := core.PrivateImageKeys.Update(id, new(models.PrivateImageKey), item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusAccepted)
}

func GetPrivateImageKey(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.PrivateImageKey)
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	if err := core.PrivateImageKeys.Get(id, item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusOK)
}

func DeletePrivateImageKey(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.PrivateImageKey)
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	if err := core.PrivateImageKeys.Delete(id, item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusAccepted)
}
