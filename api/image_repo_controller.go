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

	// TODO we need a consistent way of stripping sensitive fields from API responses.
	// And we also need a validation that key is not saved empty -- because this could cause that if someone updates tags or something.
	for _, repo := range repos.Items {
		repo.Key = ""
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

	// TODO we need a consistent way of stripping sensitive fields from API responses.
	// And we also need a validation that key is not saved empty -- because this could cause that if someone updates tags or something.
	repo.Key = ""

	body, err := marshalBody(w, repo)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
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
