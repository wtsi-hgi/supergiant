package api

import (
	"github.com/supergiant/supergiant/pkg/core"

	"github.com/gorilla/mux"
)

func NewRouter(core *core.Core) *mux.Router {
	r := mux.NewRouter()

	s := r.PathPrefix("/api/v0").Subrouter()

	// Login request can't be authenticated
	s.HandleFunc("/sessions", openHandler(core, CreateSession)).Methods("POST")

	s.HandleFunc("/sessions/{id}", restrictedHandler(core, GetSession)).Methods("GET")
	s.HandleFunc("/sessions", restrictedHandler(core, ListSessions)).Methods("GET")
	s.HandleFunc("/sessions/{id}", restrictedHandler(core, DeleteSession)).Methods("DELETE")

	s.HandleFunc("/users", restrictedHandler(core, CreateUser)).Methods("POST")
	s.HandleFunc("/users", restrictedHandler(core, ListUsers)).Methods("GET")
	s.HandleFunc("/users/{id}", restrictedHandler(core, GetUser)).Methods("GET")
	s.HandleFunc("/users/{id}", restrictedHandler(core, UpdateUser)).Methods("PATCH", "PUT")
	s.HandleFunc("/users/{id}", restrictedHandler(core, DeleteUser)).Methods("DELETE")
	s.HandleFunc("/users/{id}/regenerate_api_token", restrictedHandler(core, RegenerateUserAPIToken)).Methods("POST")

	s.HandleFunc("/cloud_accounts", restrictedHandler(core, CreateCloudAccount)).Methods("POST")
	s.HandleFunc("/cloud_accounts", restrictedHandler(core, ListCloudAccounts)).Methods("GET")
	s.HandleFunc("/cloud_accounts/{id}", restrictedHandler(core, GetCloudAccount)).Methods("GET")
	s.HandleFunc("/cloud_accounts/{id}", restrictedHandler(core, UpdateCloudAccount)).Methods("PATCH", "PUT")
	s.HandleFunc("/cloud_accounts/{id}", restrictedHandler(core, DeleteCloudAccount)).Methods("DELETE")

	s.HandleFunc("/kubes", restrictedHandler(core, CreateKube)).Methods("POST")
	s.HandleFunc("/kubes", restrictedHandler(core, ListKubes)).Methods("GET")
	s.HandleFunc("/kubes/{id}", restrictedHandler(core, GetKube)).Methods("GET")
	s.HandleFunc("/kubes/{id}", restrictedHandler(core, UpdateKube)).Methods("PATCH", "PUT")
	s.HandleFunc("/kubes/{id}", restrictedHandler(core, DeleteKube)).Methods("DELETE")

	s.HandleFunc("/apps", restrictedHandler(core, CreateApp)).Methods("POST")
	s.HandleFunc("/apps", restrictedHandler(core, ListApps)).Methods("GET")
	s.HandleFunc("/apps/{id}", restrictedHandler(core, GetApp)).Methods("GET")
	s.HandleFunc("/apps/{id}", restrictedHandler(core, UpdateApp)).Methods("PATCH", "PUT")
	s.HandleFunc("/apps/{id}", restrictedHandler(core, DeleteApp)).Methods("DELETE")

	s.HandleFunc("/components", restrictedHandler(core, CreateComponent)).Methods("POST")
	s.HandleFunc("/components", restrictedHandler(core, ListComponents)).Methods("GET")
	s.HandleFunc("/components/{id}", restrictedHandler(core, GetComponent)).Methods("GET")
	s.HandleFunc("/components/{id}", restrictedHandler(core, UpdateComponent)).Methods("PATCH", "PUT")
	s.HandleFunc("/components/{id}", restrictedHandler(core, DeleteComponent)).Methods("DELETE")
	s.HandleFunc("/components/{id}/deploy", restrictedHandler(core, DeployComponent)).Methods("POST")

	s.HandleFunc("/releases", restrictedHandler(core, CreateRelease)).Methods("POST")
	s.HandleFunc("/releases", restrictedHandler(core, ListReleases)).Methods("GET")
	s.HandleFunc("/releases/{id}", restrictedHandler(core, GetRelease)).Methods("GET")
	s.HandleFunc("/releases/{id}", restrictedHandler(core, UpdateRelease)).Methods("PATCH", "PUT")
	s.HandleFunc("/releases/{id}", restrictedHandler(core, DeleteRelease)).Methods("DELETE")

	s.HandleFunc("/instances", restrictedHandler(core, ListInstances)).Methods("GET")
	s.HandleFunc("/instances/{id}", restrictedHandler(core, GetInstance)).Methods("GET")
	s.HandleFunc("/instances/{id}/stop", restrictedHandler(core, StopInstance)).Methods("POST")
	s.HandleFunc("/instances/{id}/start", restrictedHandler(core, StartInstance)).Methods("POST")
	s.HandleFunc("/instances/{id}/log", restrictedHandler(core, ViewInstanceLog)).Methods("GET")
	s.HandleFunc("/instances/{id}", restrictedHandler(core, DeleteInstance)).Methods("DELETE")

	s.HandleFunc("/volumes", restrictedHandler(core, ListVolumes)).Methods("GET")
	s.HandleFunc("/volumes/{id}", restrictedHandler(core, GetVolume)).Methods("GET")

	s.HandleFunc("/private_image_keys", restrictedHandler(core, CreatePrivateImageKey)).Methods("POST")
	s.HandleFunc("/private_image_keys", restrictedHandler(core, ListPrivateImageKeys)).Methods("GET")
	s.HandleFunc("/private_image_keys/{id}", restrictedHandler(core, GetPrivateImageKey)).Methods("GET")
	s.HandleFunc("/private_image_keys/{id}", restrictedHandler(core, UpdatePrivateImageKey)).Methods("PATCH", "PUT")
	s.HandleFunc("/private_image_keys/{id}", restrictedHandler(core, DeletePrivateImageKey)).Methods("DELETE")

	s.HandleFunc("/entrypoints", restrictedHandler(core, CreateEntrypoint)).Methods("POST")
	s.HandleFunc("/entrypoints", restrictedHandler(core, ListEntrypoints)).Methods("GET")
	s.HandleFunc("/entrypoints/{id}", restrictedHandler(core, GetEntrypoint)).Methods("GET")
	s.HandleFunc("/entrypoints/{id}", restrictedHandler(core, UpdateEntrypoint)).Methods("PATCH", "PUT")
	s.HandleFunc("/entrypoints/{id}", restrictedHandler(core, DeleteEntrypoint)).Methods("DELETE")

	s.HandleFunc("/nodes", restrictedHandler(core, CreateNode)).Methods("POST")
	s.HandleFunc("/nodes", restrictedHandler(core, ListNodes)).Methods("GET")
	s.HandleFunc("/nodes/{id}", restrictedHandler(core, GetNode)).Methods("GET")
	s.HandleFunc("/nodes/{id}", restrictedHandler(core, UpdateNode)).Methods("PATCH", "PUT")
	s.HandleFunc("/nodes/{id}", restrictedHandler(core, DeleteNode)).Methods("DELETE")

	s.HandleFunc("/log", logHandler(core)).Methods("GET")

	return r
}
