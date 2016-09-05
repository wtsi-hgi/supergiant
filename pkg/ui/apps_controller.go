package ui

import (
	"net/http"

	"github.com/supergiant/supergiant/pkg/client"
	"github.com/supergiant/supergiant/pkg/model"
)

func NewApp(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	return renderTemplate(w, "new", map[string]interface{}{
		"title":      "Apps",
		"formAction": "/ui/apps",
		"formMethod": "POST",
		"model": map[string]interface{}{
			"kube_id": nil,
			"name":    "",
		},
	})
}

func CreateApp(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	m := new(model.App)
	if err := unmarshalFormInto(r, m); err != nil {
		return err
	}
	if err := sg.Apps.Create(m); err != nil {
		return renderTemplate(w, "new", map[string]interface{}{
			"title":      "Apps",
			"formAction": "/ui/apps",
			"formMethod": "POST",
			"model":      m,
			"error":      err.Error(),
		})
	}
	http.Redirect(w, r, "/ui/apps", http.StatusFound)
	return nil
}

func ListApps(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	fields := []map[string]interface{}{
		{
			"title": "Kube ID",
			"type":  "field_value",
			"field": "kube_id",
		},
		{
			"title": "Name",
			"type":  "field_value",
			"field": "name",
		},
	}
	return renderTemplate(w, "index", map[string]interface{}{
		"title":       "Apps",
		"uiBasePath":  "/ui/apps",
		"apiListPath": "/api/v0/apps",
		"fields":      fields,
		"showNewLink": true,
		"batchActionPaths": map[string]string{
			"Delete": "/delete",
		},
	})
}

func GetApp(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	item := new(model.App)
	if err := sg.Apps.Get(id, item); err != nil {
		return err
	}
	return renderTemplate(w, "show", map[string]interface{}{
		"title": "Apps",
		"model": item,
	})
}

func DeleteApp(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	item := new(model.App)
	if err := sg.Apps.Delete(id, item); err != nil {
		return err
	}
	// http.Redirect(w, r, "/ui/apps", http.StatusFound)
	return nil
}
