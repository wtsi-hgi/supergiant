package api

import (
	"net/http"

	"github.com/supergiant/supergiant/core"
)

type TaskController struct {
	core *core.Core
}

func (c *TaskController) Index(w http.ResponseWriter, r *http.Request) {
	tasks, err := c.core.Tasks().List()
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, tasks)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

func (c *TaskController) Show(w http.ResponseWriter, r *http.Request) {
	task, err := loadTask(c.core, w, r)
	if err != nil {
		return
	}

	body, err := marshalBody(w, task)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

func (c *TaskController) Delete(w http.ResponseWriter, r *http.Request) {
	task, err := loadTask(c.core, w, r)
	if err != nil {
		return
	}

	if err := task.Delete(); err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, task)
	if err != nil {
		return
	}
	renderWithStatusAccepted(w, body)
}
