package ui

import (
	"fmt"
	"net/http"

	"github.com/supergiant/supergiant/pkg/client"
	"github.com/supergiant/supergiant/pkg/model"
)

func CreateRelease(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	m := new(model.Release)
	if err := unmarshalFormInto(r, m); err != nil {
		return err
	}
	if err := sg.Releases.Create(m); err != nil {
		return renderTemplate(w, "new", map[string]interface{}{
			"title":      "Components",
			"formAction": "/ui/releases",
			"formMethod": "POST",
			"model":      m,
			"error":      err.Error(),
		})
	}
	http.Redirect(w, r, "/ui/components", http.StatusFound)
	return nil
}

func UpdateRelease(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	m := new(model.Release)
	if err := unmarshalFormInto(r, m); err != nil {
		return err
	}
	if err := sg.Releases.Update(id, m); err != nil {
		return renderTemplate(w, "new", map[string]interface{}{
			"title":      "Components",
			"formAction": fmt.Sprintf("/ui/releases/%d", *id),
			"formMethod": "PUT",
			"model":      m,
			"error":      err.Error(),
		})
	}
	http.Redirect(w, r, "/ui/components", http.StatusFound)
	return nil
}
