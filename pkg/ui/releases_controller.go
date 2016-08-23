package ui

import (
	"fmt"
	"net/http"

	"github.com/supergiant/supergiant/pkg/client"
	"github.com/supergiant/supergiant/pkg/models"
)

func CreateRelease(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	m := new(models.Release)
	if err := unmarshalFormInto(r, m); err != nil {
		return err
	}
	if err := sg.Releases.Create(m); err != nil {
		return renderTemplate(w, "releases/new.html", map[string]interface{}{
			"title":      "Releases",
			"formAction": "/ui/releases",
			"model":      m,
			"error":      err.Error(),
		})
	}
	http.Redirect(w, r, "/ui/components", http.StatusTemporaryRedirect)
	return nil
}

func UpdateRelease(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	m := new(models.Release)
	if err := unmarshalFormInto(r, m); err != nil {
		return err
	}
	m.ID = id
	if err := sg.Releases.Update(m); err != nil {
		return renderTemplate(w, "releases/new.html", map[string]interface{}{
			"title":      "Releases",
			"formAction": fmt.Sprintf("/ui/releases/%d", *m.ID),
			"model":      m,
			"error":      err.Error(),
		})
	}
	http.Redirect(w, r, "/ui/components", http.StatusTemporaryRedirect)
	return nil
}
