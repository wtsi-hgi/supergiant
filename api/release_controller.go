package api

import (
	"errors"
	"net/http"

	"github.com/supergiant/supergiant/core"
)

type ReleaseController struct {
	core *core.Core
}

func (c *ReleaseController) Create(w http.ResponseWriter, r *http.Request) {
	component, err := loadComponent(c.core, w, r)
	if err != nil {
		return
	}

	release := component.Releases().New()
	if err := unmarshalBodyInto(w, r, release); err != nil {
		return
	}

	release, err = component.Releases().Create(release)
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	component.TargetReleaseTimestamp = release.Timestamp
	// (this may should be elsewhere)
	// Set the task ID of the deploy on Component
	// component.DeployTaskID = task.ID
	if err := component.Save(); err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, release)
	if err != nil {
		return
	}
	renderWithStatusCreated(w, body)
}

func (c *ReleaseController) Index(w http.ResponseWriter, r *http.Request) {
	component, err := loadComponent(c.core, w, r)
	if err != nil {
		return
	}

	releases, err := component.Releases().List()
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, releases)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

func (c *ReleaseController) Show(w http.ResponseWriter, r *http.Request) {
	release, err := loadRelease(c.core, w, r)
	if err != nil {
		return
	}

	body, err := marshalBody(w, release)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

func (c *ReleaseController) Update(w http.ResponseWriter, r *http.Request) {
	release, err := loadRelease(c.core, w, r)
	if err != nil {
		return
	}

	// TODO need some consolidated validation logic
	if release.Committed {
		renderError(w, errors.New("Release is committed"), http.StatusBadRequest)
		return

	}

	if err := unmarshalBodyInto(w, r, release); err != nil {
		return
	}

	if err := release.Save(); err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, release)
	if err != nil {
		return
	}
	renderWithStatusAccepted(w, body)
}

func (c *ReleaseController) Current(w http.ResponseWriter, r *http.Request) {
	component, err := loadComponent(c.core, w, r)
	if err != nil {
		return
	}

	if component.CurrentReleaseTimestamp == nil {
		renderError(w, errors.New("No current release"), http.StatusNotFound)
		return
	}

	release, err := component.CurrentRelease()
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, release)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

func (c *ReleaseController) Target(w http.ResponseWriter, r *http.Request) {
	component, err := loadComponent(c.core, w, r)
	if err != nil {
		return
	}

	if component.TargetReleaseTimestamp == nil {
		renderError(w, errors.New("No target release"), http.StatusNotFound)
		return
	}
	release, err := component.TargetRelease()
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, release)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}
