package api

import (
	"net/http"

	"github.com/supergiant/supergiant/core"
)

type NodeController struct {
	core *core.Core
}

func (c *NodeController) Create(w http.ResponseWriter, r *http.Request) {
	node := c.core.Nodes().New()

	if err := unmarshalBodyInto(w, r, node); err != nil {
		return
	}

	core.ZeroReadonlyFields(node)

	err := c.core.Nodes().Create(node)
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, node)
	if err != nil {
		return
	}
	renderWithStatusCreated(w, body)
}

func (c *NodeController) Index(w http.ResponseWriter, r *http.Request) {
	apps, err := c.core.Nodes().List()
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, apps)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

func (c *NodeController) Show(w http.ResponseWriter, r *http.Request) {
	node, err := loadNode(c.core, w, r)
	if err != nil {
		return
	}

	body, err := marshalBody(w, node)
	if err != nil {
		return
	}
	renderWithStatusOK(w, body)
}

func (c *NodeController) Update(w http.ResponseWriter, r *http.Request) {
	node, err := loadNode(c.core, w, r)
	if err != nil {
		return
	}

	if err := unmarshalBodyInto(w, r, node); err != nil {
		return
	}

	core.ZeroReadonlyFields(node)

	if err := node.Update(); err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, node)
	if err != nil {
		return
	}
	renderWithStatusAccepted(w, body)
}

func (c *NodeController) Delete(w http.ResponseWriter, r *http.Request) {
	node, err := loadNode(c.core, w, r)
	if err != nil {
		return
	}

	if err := node.Delete(); err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	body, err := marshalBody(w, node)
	if err != nil {
		return
	}
	renderWithStatusAccepted(w, body)
}
