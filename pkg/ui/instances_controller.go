package ui

import (
	"net/http"

	"github.com/supergiant/supergiant/pkg/client"
	"github.com/supergiant/supergiant/pkg/models"
)

func ListInstances(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	fields := []map[string]interface{}{
		{
			"title": "Component ID",
			"type":  "field_value",
			"field": "component_id",
		},
		{
			"title": "Name",
			"type":  "field_value",
			"field": "name",
		},
		{
			"title":             "CPU Usage",
			"type":              "percentage",
			"numerator_field":   "cpu_usage",
			"denominator_field": "cpu_limit",
		},
		{
			"title":             "RAM usage",
			"type":              "percentage",
			"numerator_field":   "ram_usage",
			"denominator_field": "ram_limit",
		},
	}
	return renderTemplate(w, "instances/index.html", map[string]interface{}{
		"title":       "Instances",
		"uiBasePath":  "/ui/instances",
		"apiListPath": "/api/v0/instances",
		"fields":      fields,
		"showNewLink": false,
		// "batchActionPaths": map[string]string{
		// 	"Stop":  "/stop",
		// 	"Start": "/start",
		// },
	})
}

func GetInstance(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	item := new(models.Instance)
	if err := sg.Instances.Get(id, item); err != nil {
		return err
	}
	return renderTemplate(w, "instances/show.html", map[string]interface{}{
		"title": "Instances",
		"model": item,
	})
}
