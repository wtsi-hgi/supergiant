package ui

import (
	"net/http"

	"github.com/supergiant/supergiant/pkg/client"
	"github.com/supergiant/supergiant/pkg/models"
)

func NewKube(sg *client.Client, w http.ResponseWriter, r *http.Request) error {

	// TODO we shouldn't use a map for the model, because attr ordering is screwed
	// up. But it's difficult to initialize blank values with omitemptys (which
	// are needed for certain things), so we should probably have special structs.

	return renderTemplate(w, "kubes/new.html", map[string]interface{}{
		"title":      "Kubes",
		"formAction": "/ui/kubes",
		"model": map[string]interface{}{
			"cloud_account_id": nil,
			"name":             "",
			"config": map[string]interface{}{
				"region":               "us-east-1",
				"availability_zone":    "us-east-1b",
				"master_instance_type": "m4.large",
				"instance_types": []string{
					"m4.large",
					"m4.xlarge",
					"m4.2xlarge",
					"m4.4xlarge",
				},
				"vpc_ip_range":           "172.20.0.0/16",
				"public_subnet_ip_range": "172.20.0.0/24",
				"master_private_ip":      "172.20.0.9",
			},
		},
	})
}

func CreateKube(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	m := new(models.Kube)
	if err := unmarshalFormInto(r, m); err != nil {
		return err
	}
	if err := sg.Kubes.Create(m); err != nil {
		return renderTemplate(w, "kubes/new.html", map[string]interface{}{
			"title":      "Kubes",
			"formAction": "/ui/kubes",
			"model":      m,
			"error":      err.Error(),
		})
	}
	http.Redirect(w, r, "/ui/kubes", http.StatusFound)
	return nil
}

func ListKubes(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	fields := []map[string]interface{}{
		{
			"title": "Name",
			"type":  "field_value",
			"field": "name",
		},
		{
			"title": "Availability Zone",
			"type":  "field_value",
			"field": "config.availability_zone",
		},
		{
			"title": "Master Size",
			"type":  "field_value",
			"field": "config.master_instance_type",
		},
	}
	return renderTemplate(w, "kubes/index.html", map[string]interface{}{
		"title":       "Kubes",
		"uiBasePath":  "/ui/kubes",
		"apiListPath": "/api/v0/kubes",
		"fields":      fields,
		"showNewLink": true,
		"batchActionPaths": map[string]string{
			"Delete": "/delete",
		},
	})
}

func GetKube(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	item := new(models.Kube)
	if err := sg.Kubes.Get(id, item); err != nil {
		return err
	}
	return renderTemplate(w, "kubes/show.html", map[string]interface{}{
		"title": "Kubes",
		"model": item,
	})
}

func DeleteKube(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	id, err := parseID(r)
	if err != nil {
		return err
	}
	item := new(models.Kube)
	item.ID = id
	if err := sg.Kubes.Delete(item); err != nil {
		return err
	}
	// http.Redirect(w, r, "/ui/kubes", http.StatusFound)
	return nil
}
