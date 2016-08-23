package ui

import (
	"fmt"
	"net/http"

	"github.com/supergiant/supergiant/pkg/client"
	"github.com/supergiant/supergiant/pkg/models"
)

func NewComponent(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	return renderTemplate(w, "components/new.html", map[string]interface{}{
		"title":      "Components",
		"formAction": "/ui/components",
		"model": map[string]interface{}{
			"app_id":             nil,
			"name":               "",
			"private_image_keys": []string{},
		},
	})
}

func CreateComponent(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	m := new(models.Component)
	if err := unmarshalFormInto(r, m); err != nil {
		return err
	}
	if err := sg.Components.Create(m); err != nil {
		return renderTemplate(w, "components/new.html", map[string]interface{}{
			"title":      "Components",
			"formAction": "/ui/components",
			"model":      m,
			"error":      err.Error(),
		})
	}
	http.Redirect(w, r, fmt.Sprintf("/ui/components/%d/configure", *m.ID), http.StatusTemporaryRedirect)
	return nil
}

func ListComponents(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	fields := []map[string]interface{}{
		{
			"title": "App ID",
			"type":  "field_value",
			"field": "app_id",
		},
		{
			"title": "Name",
			"type":  "field_value",
			"field": "name",
		},
	}
	return renderTemplate(w, "components/index.html", map[string]interface{}{
		"title":       "Components",
		"uiBasePath":  "/ui/components",
		"apiListPath": "/api/v0/components",
		"fields":      fields,
		"showNewLink": true,
		"actionPaths": map[string]string{
			"Configure": "/configure",
		},
		"batchActionPaths": map[string]string{
			"Deploy": "/deploy",
			"Delete": "/delete",
		},
	})
}

func GetComponent(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	item := new(models.Component)
	if err := sg.Components.Get(id, item); err != nil {
		return err
	}
	return renderTemplate(w, "components/show.html", map[string]interface{}{
		"title": "Components",
		"model": item,
	})
}

func DeployComponent(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	item := new(models.Component)
	item.ID = id
	if err := sg.Components.Deploy(item); err != nil {
		return err
	}
	// http.Redirect(w, r, "/ui/components", http.StatusTemporaryRedirect)
	return nil
}

func DeleteComponent(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	item := new(models.Component)
	item.ID = id
	if err := sg.Components.Delete(item); err != nil {
		return err
	}
	// http.Redirect(w, r, "/ui/components", http.StatusTemporaryRedirect)
	return nil
}

//------------------------------------------------------------------------------

// NOTE this is really just NewRelease, with Component auto-filled
func ConfigureComponent(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	item := new(models.Component)
	includes := []string{"TargetRelease", "CurrentRelease"}
	if err := sg.Components.GetWithIncludes(id, item, includes); err != nil {
		return err
	}

	var model interface{}
	var formAction string

	if item.TargetRelease != nil {
		model = item.TargetRelease
		formAction = fmt.Sprintf("/ui/releases/%d", *item.TargetReleaseID)

	} else if item.CurrentRelease != nil {
		release := item.CurrentRelease
		models.ZeroReadonlyFields(release)
		model = release
		formAction = "/ui/releases"

	} else {
		model = map[string]interface{}{
			"component_id":   id,
			"instance_count": 1,
			"config": map[string]interface{}{
				"volumes": []map[string]interface{}{},
				"containers": []map[string]interface{}{
					{
						"image":  "",
						"ports":  []interface{}{},
						"mounts": []interface{}{},
						"env":    []interface{}{},
					},
				},
			},
		}
		formAction = "/ui/releases"
	}

	return renderTemplate(w, "releases/new.html", map[string]interface{}{
		"title":      "Components",
		"formAction": formAction,
		"model":      model,
	})
}
