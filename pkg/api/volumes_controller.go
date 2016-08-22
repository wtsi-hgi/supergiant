package api

import (
	"net/http"

	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/models"
)

func ListVolumes(core *core.Core, r *http.Request) (*Response, error) {
	return handleList(core, r, new(models.Volume))
}

func GetVolume(core *core.Core, r *http.Request) (*Response, error) {
	item := new(models.Volume)
	id, err := parseID(r)
	if err != nil {
		return nil, err
	}
	if err := core.Volumes.Get(id, item); err != nil {
		return nil, err
	}
	return itemResponse(core, item, http.StatusOK)
}
