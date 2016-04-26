package api

import (
	"net/http"

	"github.com/supergiant/supergiant/core"
)

type ImageRepoController struct {
	core *core.Core
}

func (c *ImageRepoController) Create(w http.ResponseWriter, r *http.Request) {
	repo := c.core.ImageRepos().New()

	if err := unmarshalBodyInto(w, r, repo); err != nil {
		return
	}

	core.ZeroReadonlyFields(repo)

	err := c.core.ImageRepos().Create(repo)
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, repo)
	if err != nil {
		return
	}
	renderWithStatusCreated(w, body)
}

func (c *ImageRepoController) Index(w http.ResponseWriter, r *http.Request) {
	repos, err := c.core.ImageRepos().List()
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	// TODO this _could_ be stuffed in marshalBody... but List is what breaks it.
	// It's like we need a separate method core.ZeroPrivateFieldsOnList
	for _, repo := range repos.Items {
		core.ZeroPrivateFields(repo)
	}

	body, err := marshalBody(w, repos)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

func (c *ImageRepoController) Show(w http.ResponseWriter, r *http.Request) {
	repo, err := loadImageRepo(c.core, w, r)
	if err != nil {
		return
	}

	// see TODO above
	core.ZeroPrivateFields(repo)

	body, err := marshalBody(w, repo)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

func (c *ImageRepoController) Update(w http.ResponseWriter, r *http.Request) {
	repo, err := loadImageRepo(c.core, w, r)
	if err != nil {
		return
	}

	if err := unmarshalBodyInto(w, r, repo); err != nil {
		return
	}

	core.ZeroReadonlyFields(repo)

	if err := repo.Update(); err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	core.ZeroPrivateFields(repo)

	body, err := marshalBody(w, repo)
	if err != nil {
		return
	}
	renderWithStatusAccepted(w, body)
}

func (c *ImageRepoController) Delete(w http.ResponseWriter, r *http.Request) {
	repo, err := loadImageRepo(c.core, w, r)
	if err != nil {
		return
	}
	if err = repo.Delete(); err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}
}
