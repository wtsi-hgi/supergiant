package ui

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/supergiant/supergiant/pkg/client"
	"github.com/supergiant/supergiant/pkg/models"
)

var templates map[string]*template.Template

func init() {
	templates = make(map[string]*template.Template)

	sharedViews, err := filepath.Glob("ui/views/_shared/*.html")
	if err != nil {
		panic(err)
	}

	fullViews, err := filepath.Glob("ui/views/[a-z]*/[a-z]*.html")
	if err != nil {
		panic(err)
	}

	for _, view := range fullViews {
		key := regexp.MustCompile("/([^/]+/[^/]+.html)$").FindStringSubmatch(view)[1]
		templates[key] = template.Must(template.ParseFiles(append(sharedViews, view)...))
	}
}

func NewRouter(sg *client.Client, baseRouter *mux.Router) *mux.Router {

	r := baseRouter.PathPrefix("/ui").Subrouter().StrictSlash(true)

	r.HandleFunc("/", handlerFunc(sg, Root)).Methods("GET")

	r.HandleFunc("/cloud_accounts/new", handlerFunc(sg, NewCloudAccount)).Methods("GET")
	r.HandleFunc("/cloud_accounts", handlerFunc(sg, CreateCloudAccount)).Methods("POST")
	r.HandleFunc("/cloud_accounts", handlerFunc(sg, ListCloudAccounts)).Methods("GET")
	r.HandleFunc("/cloud_accounts/{id}", handlerFunc(sg, GetCloudAccount)).Methods("GET")
	r.HandleFunc("/cloud_accounts/{id}/delete", handlerFunc(sg, DeleteCloudAccount)).Methods("PUT")

	r.HandleFunc("/kubes/new", handlerFunc(sg, NewKube)).Methods("GET")
	r.HandleFunc("/kubes", handlerFunc(sg, CreateKube)).Methods("POST")
	r.HandleFunc("/kubes", handlerFunc(sg, ListKubes)).Methods("GET")
	r.HandleFunc("/kubes/{id}", handlerFunc(sg, GetKube)).Methods("GET")
	r.HandleFunc("/kubes/{id}/delete", handlerFunc(sg, DeleteKube)).Methods("PUT")

	r.HandleFunc("/apps/new", handlerFunc(sg, NewApp)).Methods("GET")
	r.HandleFunc("/apps", handlerFunc(sg, CreateApp)).Methods("POST")
	r.HandleFunc("/apps", handlerFunc(sg, ListApps)).Methods("GET")
	r.HandleFunc("/apps/{id}", handlerFunc(sg, GetApp)).Methods("GET")
	r.HandleFunc("/apps/{id}/delete", handlerFunc(sg, DeleteApp)).Methods("PUT")

	r.HandleFunc("/components/new", handlerFunc(sg, NewComponent)).Methods("GET")
	r.HandleFunc("/components", handlerFunc(sg, CreateComponent)).Methods("POST")
	r.HandleFunc("/components", handlerFunc(sg, ListComponents)).Methods("GET")
	r.HandleFunc("/components/{id}", handlerFunc(sg, GetComponent)).Methods("GET")
	r.HandleFunc("/components/{id}/delete", handlerFunc(sg, DeleteComponent)).Methods("PUT")
	r.HandleFunc("/components/{id}/deploy", handlerFunc(sg, DeployComponent)).Methods("PUT")
	r.HandleFunc("/components/{id}/configure", handlerFunc(sg, ConfigureComponent)).Methods("GET")

	r.HandleFunc("/releases", handlerFunc(sg, CreateRelease)).Methods("POST")
	r.HandleFunc("/releases/{id}", handlerFunc(sg, UpdateRelease)).Methods("POST")

	r.HandleFunc("/instances", handlerFunc(sg, ListInstances)).Methods("GET")
	r.HandleFunc("/instances/{id}", handlerFunc(sg, GetInstance)).Methods("GET")

	r.HandleFunc("/volumes", handlerFunc(sg, ListVolumes)).Methods("GET")
	r.HandleFunc("/volumes/{id}", handlerFunc(sg, GetVolume)).Methods("GET")

	r.HandleFunc("/private_image_keys/new", handlerFunc(sg, NewPrivateImageKey)).Methods("GET")
	r.HandleFunc("/private_image_keys", handlerFunc(sg, CreatePrivateImageKey)).Methods("POST")
	r.HandleFunc("/private_image_keys", handlerFunc(sg, ListPrivateImageKeys)).Methods("GET")
	r.HandleFunc("/private_image_keys/{id}", handlerFunc(sg, GetPrivateImageKey)).Methods("GET")
	r.HandleFunc("/private_image_keys/{id}/delete", handlerFunc(sg, DeletePrivateImageKey)).Methods("PUT")

	r.HandleFunc("/entrypoints/new", handlerFunc(sg, NewEntrypoint)).Methods("GET")
	r.HandleFunc("/entrypoints", handlerFunc(sg, CreateEntrypoint)).Methods("POST")
	r.HandleFunc("/entrypoints", handlerFunc(sg, ListEntrypoints)).Methods("GET")
	r.HandleFunc("/entrypoints/{id}", handlerFunc(sg, GetEntrypoint)).Methods("GET")
	r.HandleFunc("/entrypoints/{id}/delete", handlerFunc(sg, DeleteEntrypoint)).Methods("PUT")

	r.HandleFunc("/nodes/new", handlerFunc(sg, NewNode)).Methods("GET")
	r.HandleFunc("/nodes", handlerFunc(sg, CreateNode)).Methods("POST")
	r.HandleFunc("/nodes", handlerFunc(sg, ListNodes)).Methods("GET")
	r.HandleFunc("/nodes/{id}", handlerFunc(sg, GetNode)).Methods("GET")
	r.HandleFunc("/nodes/{id}/delete", handlerFunc(sg, DeleteNode)).Methods("PUT")

	return baseRouter
}

func handlerFunc(sg *client.Client, fn func(*client.Client, http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO repeated in API, move to core helpers
		w.Header().Set("WWW-Authenticate", `Basic realm="supergiant"`)
		username, password, _ := r.BasicAuth()
		if username != sg.Username || password != sg.Password {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		if err := fn(sg, w, r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	}
}

//------------------------------------------------------------------------------

func parseID(r *http.Request) (*int64, error) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return nil, err
	}
	id64 := int64(id)
	return &id64, nil
}

func renderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) error {

	// TODO
	if mi := data["model"]; mi != nil {
		if model, isModel := mi.(models.Model); isModel {
			models.ZeroPrivateFields(model)
			data["model"] = model
		}
	}

	modelJSON, _ := json.Marshal(data["model"])
	data["modelJSON"] = string(modelJSON)

	if fields, ok := data["fields"]; ok {
		fieldsJSON, _ := json.Marshal(fields)
		data["fieldsJSON"] = string(fieldsJSON)
	}

	if err := templates[name].ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

func unmarshalFormInto(r *http.Request, out interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return json.Unmarshal([]byte(r.PostForm.Get("json_input")), out)
}

//------------------------------------------------------------------------------

func Root(sg *client.Client, w http.ResponseWriter, r *http.Request) error {
	var cloudAccounts []*models.CloudAccount
	if err := sg.CloudAccounts.List(&cloudAccounts); err != nil {
		return err
	}
	if len(cloudAccounts) == 0 {
		http.Redirect(w, r, "/ui/cloud_accounts/new", 302)
		return nil
	}

	var kubes []*models.Kube
	if err := sg.Kubes.List(&kubes); err != nil {
		return err
	}
	if len(kubes) == 0 {
		http.Redirect(w, r, "/ui/kubes/new", 302)
		return nil
	}

	http.Redirect(w, r, "/ui/apps", 302)
	return nil
}
