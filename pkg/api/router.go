package api

import (
	"github.com/supergiant/supergiant/pkg/core"

	"github.com/gorilla/mux"
)

func NewRouter(core *core.Core) *mux.Router {
	r := mux.NewRouter()

	s := r.PathPrefix("/api/v0").Subrouter()

	s.HandleFunc("/cloud_accounts", handlerFunc(core, CreateCloudAccount)).Methods("POST")
	s.HandleFunc("/cloud_accounts", handlerFunc(core, ListCloudAccounts)).Methods("GET")
	s.HandleFunc("/cloud_accounts/{id}", handlerFunc(core, GetCloudAccount)).Methods("GET")
	s.HandleFunc("/cloud_accounts/{id}", handlerFunc(core, UpdateCloudAccount)).Methods("PATCH", "PUT")
	s.HandleFunc("/cloud_accounts/{id}", handlerFunc(core, DeleteCloudAccount)).Methods("DELETE")

	s.HandleFunc("/kubes", handlerFunc(core, CreateKube)).Methods("POST")
	s.HandleFunc("/kubes", handlerFunc(core, ListKubes)).Methods("GET")
	s.HandleFunc("/kubes/{id}", handlerFunc(core, GetKube)).Methods("GET")
	s.HandleFunc("/kubes/{id}", handlerFunc(core, UpdateKube)).Methods("PATCH", "PUT")
	s.HandleFunc("/kubes/{id}", handlerFunc(core, DeleteKube)).Methods("DELETE")

	s.HandleFunc("/apps", handlerFunc(core, CreateApp)).Methods("POST")
	s.HandleFunc("/apps", handlerFunc(core, ListApps)).Methods("GET")
	s.HandleFunc("/apps/{id}", handlerFunc(core, GetApp)).Methods("GET")
	s.HandleFunc("/apps/{id}", handlerFunc(core, UpdateApp)).Methods("PATCH", "PUT")
	s.HandleFunc("/apps/{id}", handlerFunc(core, DeleteApp)).Methods("DELETE")

	s.HandleFunc("/components", handlerFunc(core, CreateComponent)).Methods("POST")
	s.HandleFunc("/components", handlerFunc(core, ListComponents)).Methods("GET")
	s.HandleFunc("/components/{id}", handlerFunc(core, GetComponent)).Methods("GET")
	s.HandleFunc("/components/{id}", handlerFunc(core, UpdateComponent)).Methods("PATCH", "PUT")
	s.HandleFunc("/components/{id}", handlerFunc(core, DeleteComponent)).Methods("DELETE")
	s.HandleFunc("/components/{id}/deploy", handlerFunc(core, DeployComponent)).Methods("POST")

	s.HandleFunc("/releases", handlerFunc(core, CreateRelease)).Methods("POST")
	s.HandleFunc("/releases", handlerFunc(core, ListReleases)).Methods("GET")
	s.HandleFunc("/releases/{id}", handlerFunc(core, GetRelease)).Methods("GET")
	s.HandleFunc("/releases/{id}", handlerFunc(core, UpdateRelease)).Methods("PATCH", "PUT")
	s.HandleFunc("/releases/{id}", handlerFunc(core, DeleteRelease)).Methods("DELETE")

	s.HandleFunc("/instances", handlerFunc(core, ListInstances)).Methods("GET")
	s.HandleFunc("/instances/{id}", handlerFunc(core, GetInstance)).Methods("GET")
	s.HandleFunc("/instances/{id}/stop", handlerFunc(core, StopInstance)).Methods("POST")
	s.HandleFunc("/instances/{id}/start", handlerFunc(core, StartInstance)).Methods("POST")
	s.HandleFunc("/instances/{id}/log", handlerFunc(core, ViewInstanceLog)).Methods("GET")
	s.HandleFunc("/instances/{id}", handlerFunc(core, DeleteInstance)).Methods("DELETE")

	s.HandleFunc("/volumes", handlerFunc(core, ListVolumes)).Methods("GET")
	s.HandleFunc("/volumes/{id}", handlerFunc(core, GetVolume)).Methods("GET")

	s.HandleFunc("/private_image_keys", handlerFunc(core, CreatePrivateImageKey)).Methods("POST")
	s.HandleFunc("/private_image_keys", handlerFunc(core, ListPrivateImageKeys)).Methods("GET")
	s.HandleFunc("/private_image_keys/{id}", handlerFunc(core, GetPrivateImageKey)).Methods("GET")
	s.HandleFunc("/private_image_keys/{id}", handlerFunc(core, UpdatePrivateImageKey)).Methods("PATCH", "PUT")
	s.HandleFunc("/private_image_keys/{id}", handlerFunc(core, DeletePrivateImageKey)).Methods("DELETE")

	s.HandleFunc("/entrypoints", handlerFunc(core, CreateEntrypoint)).Methods("POST")
	s.HandleFunc("/entrypoints", handlerFunc(core, ListEntrypoints)).Methods("GET")
	s.HandleFunc("/entrypoints/{id}", handlerFunc(core, GetEntrypoint)).Methods("GET")
	s.HandleFunc("/entrypoints/{id}", handlerFunc(core, UpdateEntrypoint)).Methods("PATCH", "PUT")
	s.HandleFunc("/entrypoints/{id}", handlerFunc(core, DeleteEntrypoint)).Methods("DELETE")

	s.HandleFunc("/nodes", handlerFunc(core, CreateNode)).Methods("POST")
	s.HandleFunc("/nodes", handlerFunc(core, ListNodes)).Methods("GET")
	s.HandleFunc("/nodes/{id}", handlerFunc(core, GetNode)).Methods("GET")
	s.HandleFunc("/nodes/{id}", handlerFunc(core, UpdateNode)).Methods("PATCH", "PUT")
	s.HandleFunc("/nodes/{id}", handlerFunc(core, DeleteNode)).Methods("DELETE")

	s.HandleFunc("/log", logHandler(core)).Methods("GET")

	return r
}
