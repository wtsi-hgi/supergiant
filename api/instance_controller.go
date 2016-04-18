package api

import (
	"net/http"

	"github.com/supergiant/supergiant/core"
)

type InstanceController struct {
	core *core.Core
}

func (c *InstanceController) Index(w http.ResponseWriter, r *http.Request) {
	release, err := loadRelease(c.core, w, r)
	if err != nil {
		return
	}

	instances := release.Instances().List()

	body, err := marshalBody(w, instances)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

func (c *InstanceController) Show(w http.ResponseWriter, r *http.Request) {
	instance, err := loadInstance(c.core, w, r)
	if err != nil {
		return
	}

	body, err := marshalBody(w, instance)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

// TODO this is not JSON
func (c *InstanceController) Log(w http.ResponseWriter, r *http.Request) {
	instance, err := loadInstance(c.core, w, r)
	if err != nil {
		return
	}

	log, err := instance.Log()
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	renderWithStatusOK(w, log)
}

func (c *InstanceController) Start(w http.ResponseWriter, r *http.Request) {
	instance, err := loadInstance(c.core, w, r)
	if err != nil {
		return
	}

	if err := instance.Action("start").Supervise(); err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, instance)
	if err != nil {
		return
	}
	renderWithStatusAccepted(w, body)
}

func (c *InstanceController) Stop(w http.ResponseWriter, r *http.Request) {
	instance, err := loadInstance(c.core, w, r)
	if err != nil {
		return
	}

	if err := instance.Action("stop").Supervise(); err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, instance)
	if err != nil {
		return
	}
	renderWithStatusAccepted(w, body)
}
