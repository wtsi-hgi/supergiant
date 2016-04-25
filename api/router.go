package api

import (
	"github.com/supergiant/supergiant/core"

	"github.com/gorilla/mux"
)

func NewRouter(core *core.Core) *mux.Router {
	r := mux.NewRouter()

	s := r.PathPrefix("/v0").Subrouter()

	// this redirects /apps/ to /apps
	s.StrictSlash(true)

	imageRepos := &ImageRepoController{core}
	entrypoints := &EntrypointController{core}
	apps := &AppController{core}
	components := &ComponentController{core}
	releases := &ReleaseController{core}
	instances := &InstanceController{core}
	tasks := &TaskController{core}
	nodes := &NodeController{core}

	s.HandleFunc("/registries/dockerhub/repos", imageRepos.Create).Methods("POST")
	s.HandleFunc("/registries/dockerhub/repos", imageRepos.Index).Methods("GET")
	s.HandleFunc("/registries/dockerhub/repos/{name}", imageRepos.Show).Methods("GET")
	s.HandleFunc("/registries/dockerhub/repos/{name}", imageRepos.Delete).Methods("DELETE")

	s.HandleFunc("/nodes", nodes.Create).Methods("POST")
	s.HandleFunc("/nodes", nodes.Index).Methods("GET")
	s.HandleFunc("/nodes/{node_id}", nodes.Show).Methods("GET")
	s.HandleFunc("/nodes/{node_id}", nodes.Update).Methods("PUT")
	s.HandleFunc("/nodes/{node_id}", nodes.Delete).Methods("DELETE")

	s.HandleFunc("/entrypoints", entrypoints.Create).Methods("POST")
	s.HandleFunc("/entrypoints", entrypoints.Index).Methods("GET")
	s.HandleFunc("/entrypoints/{domain}", entrypoints.Show).Methods("GET")
	s.HandleFunc("/entrypoints/{domain}", entrypoints.Delete).Methods("DELETE")

	s.HandleFunc("/apps", apps.Create).Methods("POST")
	s.HandleFunc("/apps", apps.Index).Methods("GET")
	s.HandleFunc("/apps/{app_name}", apps.Show).Methods("GET")
	s.HandleFunc("/apps/{app_name}", apps.Update).Methods("PUT")
	s.HandleFunc("/apps/{app_name}", apps.Delete).Methods("DELETE")

	s.HandleFunc("/apps/{app_name}/components", components.Create).Methods("POST")
	s.HandleFunc("/apps/{app_name}/components", components.Index).Methods("GET")
	s.HandleFunc("/apps/{app_name}/components/{comp_name}", components.Show).Methods("GET")
	s.HandleFunc("/apps/{app_name}/components/{comp_name}", components.Update).Methods("PUT")
	s.HandleFunc("/apps/{app_name}/components/{comp_name}", components.Delete).Methods("DELETE")

	s.HandleFunc("/apps/{app_name}/components/{comp_name}/releases", releases.Create).Methods("POST")
	s.HandleFunc("/apps/{app_name}/components/{comp_name}/releases", releases.MergeCreate).Methods("PATCH")
	s.HandleFunc("/apps/{app_name}/components/{comp_name}/releases", releases.Index).Methods("GET")
	s.HandleFunc("/apps/{app_name}/components/{comp_name}/releases/{release_timestamp}", releases.Show).Methods("GET")
	s.HandleFunc("/apps/{app_name}/components/{comp_name}/releases/{release_timestamp}", releases.Update).Methods("PUT")
	s.HandleFunc("/apps/{app_name}/components/{comp_name}/releases/{release_timestamp}", releases.Delete).Methods("DELETE")

	s.HandleFunc("/apps/{app_name}/components/{comp_name}/deploy", components.Deploy).Methods("POST")

	// Integration

	s.HandleFunc("/apps/{app_name}/components/{comp_name}/releases/{release_timestamp}/instances", instances.Index).Methods("GET")
	s.HandleFunc("/apps/{app_name}/components/{comp_name}/releases/{release_timestamp}/instances/{instance_id}", instances.Show).Methods("GET")

	s.HandleFunc("/apps/{app_name}/components/{comp_name}/releases/{release_timestamp}/instances/{instance_id}/start", instances.Start).Methods("POST")
	s.HandleFunc("/apps/{app_name}/components/{comp_name}/releases/{release_timestamp}/instances/{instance_id}/stop", instances.Stop).Methods("POST")

	s.HandleFunc("/apps/{app_name}/components/{comp_name}/releases/{release_timestamp}/instances/{instance_id}/log", instances.Log).Methods("GET")

	// Misc

	s.HandleFunc("/tasks", tasks.Index).Methods("GET")
	s.HandleFunc("/tasks/{id}", tasks.Show).Methods("GET")
	s.HandleFunc("/tasks/{id}", tasks.Delete).Methods("DELETE")

	return r
}
