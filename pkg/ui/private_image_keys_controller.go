package ui

import (
	"net/http"

	"github.com/supergiant/supergiant/pkg/client"
	"github.com/supergiant/supergiant/pkg/models"
)

func NewPrivateImageKey(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	return renderTemplate(w, "private_image_keys/new.html", map[string]interface{}{
		"title":      "Private Image Keys",
		"formAction": "/ui/private_image_keys",
		"model": map[string]interface{}{
			"host":     "index.docker.io",
			"username": "",
			"email":    "",
			"password": "",
		},
	})
}

func CreatePrivateImageKey(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	m := new(models.PrivateImageKey)
	if err := unmarshalFormInto(r, m); err != nil {
		return err
	}
	if err := sg.PrivateImageKeys.Create(m); err != nil {
		return renderTemplate(w, "private_image_keys/new.html", map[string]interface{}{
			"title":      "Private Image Keys",
			"formAction": "/ui/private_image_keys",
			"model":      m,
			"error":      err.Error(),
		})
	}
	http.Redirect(w, r, "/ui/private_image_keys", http.StatusTemporaryRedirect)
	return nil
}

func ListPrivateImageKeys(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	fields := []map[string]interface{}{
		{
			"title": "Host",
			"type":  "field_value",
			"field": "host",
		},
		{
			"title": "Username",
			"type":  "field_value",
			"field": "username",
		},
	}
	return renderTemplate(w, "private_image_keys/index.html", map[string]interface{}{
		"title":       "Private Image Keys",
		"uiBasePath":  "/ui/private_image_keys",
		"apiListPath": "/api/v0/private_image_keys",
		"fields":      fields,
		"showNewLink": true,
		"batchActionPaths": map[string]string{
			"Delete": "/delete",
		},
	})
}

func GetPrivateImageKey(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	item := new(models.PrivateImageKey)
	if err := sg.PrivateImageKeys.Get(id, item); err != nil {
		return err
	}
	return renderTemplate(w, "private_image_keys/show.html", map[string]interface{}{
		"title": "Private Image Keys",
		"model": item,
	})
}

func DeletePrivateImageKey(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	item := new(models.PrivateImageKey)
	item.ID = id
	if err := sg.PrivateImageKeys.Delete(item); err != nil {
		return err
	}
	// http.Redirect(w, r, "/ui/private_image_keys", http.StatusTemporaryRedirect)
	return nil
}
