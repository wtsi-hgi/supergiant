package api

import (
	"errors"
	"net/http"

	"github.com/supergiant/supergiant/core"
)

type ComponentController struct {
	core *core.Core
}

func (c *ComponentController) Create(w http.ResponseWriter, r *http.Request) {
	app, err := loadApp(c.core, w, r)
	if err != nil {
		return
	}

	component := app.Components().New()
	if err := unmarshalBodyInto(w, r, component); err != nil {
		return
	}

	core.ZeroReadonlyFields(component)

	err = app.Components().Create(component)
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, component)
	if err != nil {
		return
	}
	renderWithStatusCreated(w, body)
}

func (c *ComponentController) Index(w http.ResponseWriter, r *http.Request) {
	app, err := loadApp(c.core, w, r)
	if err != nil {
		return
	}

	components, err := app.Components().List()
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, components)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

func (c *ComponentController) Show(w http.ResponseWriter, r *http.Request) {
	component, err := loadComponent(c.core, w, r)
	if err != nil {
		return
	}

	body, err := marshalBody(w, component)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

func (c *ComponentController) Update(w http.ResponseWriter, r *http.Request) {
	component, err := loadComponent(c.core, w, r)
	if err != nil {
		return
	}

	if err := unmarshalBodyInto(w, r, component); err != nil {
		return
	}

	core.ZeroReadonlyFields(component)

	if err := component.Update(); err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, component)
	if err != nil {
		return
	}
	renderWithStatusAccepted(w, body)
}

func (c *ComponentController) Delete(w http.ResponseWriter, r *http.Request) {
	component, err := loadComponent(c.core, w, r)
	if err != nil {
		return
	}

	if err := component.Action("delete").Supervise(); err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, component)
	if err != nil {
		return
	}
	renderWithStatusAccepted(w, body)
}

func (c *ComponentController) Deploy(w http.ResponseWriter, r *http.Request) {
	component, err := loadComponent(c.core, w, r)
	if err != nil {
		return
	}

	if component.TargetReleaseTimestamp == nil {
		renderError(w, errors.New("Component does not have target Release"), http.StatusBadRequest)
		return
	}

	release, err := component.TargetRelease()
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	release.Committed = true
	if err := release.Update(); err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	if err := component.Action("deploy").Supervise(); err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, component)
	if err != nil {
		return
	}
	renderWithStatusAccepted(w, body)
}
