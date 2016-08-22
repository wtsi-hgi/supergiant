package ui

import (
	"net/http"

	"github.com/supergiant/supergiant/pkg/client"
	"github.com/supergiant/supergiant/pkg/models"
)

func ListVolumes(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	fields := []map[string]interface{}{
		{
			"title": "Kube ID",
			"type":  "field_value",
			"field": "kube_id",
		},
		{
			"title": "Instance ID",
			"type":  "field_value",
			"field": "instance_id",
		},
		{
			"title": "Name",
			"type":  "field_value",
			"field": "name",
		},
		{
			"title": "Type",
			"type":  "field_value",
			"field": "type",
		},
		{
			"title": "GB",
			"type":  "field_value",
			"field": "size",
		},
		{
			"title": "Provider ID",
			"type":  "field_value",
			"field": "provider_id",
		},
	}
	return renderTemplate(w, "volumes/index.html", map[string]interface{}{
		"title":       "Volumes",
		"uiBasePath":  "/ui/volumes",
		"apiListPath": "/api/v0/volumes",
		"fields":      fields,
		"showNewLink": false,
	})
}

func GetVolume(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	item := new(models.Volume)
	if err := sg.Volumes.Get(id, item); err != nil {
		return err
	}
	return renderTemplate(w, "volumes/show.html", map[string]interface{}{
		"title": "Volumes",
		"model": item,
	})
}
