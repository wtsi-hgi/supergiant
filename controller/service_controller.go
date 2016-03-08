package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"supergiant/core/model"
	"supergiant/core/storage"

	"github.com/gorilla/mux"
)

type ServiceController struct {
	db *storage.Client
}

type serviceListResponse struct {
	Items []*model.Service `json:"items"`
}

func NewServiceController(router *mux.Router, db *storage.Client) *ServiceController {
	controller := ServiceController{db: db}
	router.HandleFunc("/environments/{env_name}/services", controller.Create).Methods("POST")
	router.HandleFunc("/environments/{env_name}/services", controller.Index).Methods("GET")
	router.HandleFunc("/environments/{env_name}/services/{name}", controller.Show).Methods("GET")
	router.HandleFunc("/environments/{env_name}/services/{name}", controller.Delete).Methods("DELETE")
	return &controller
}

func (s *ServiceController) loadEnvironment(envName string) (*model.Environment, error) {
	return s.db.EnvironmentStorage.Get(envName)
}

func (s *ServiceController) Create(w http.ResponseWriter, r *http.Request) {
	envName := mux.Vars(r)["env_name"]
	env, err := s.loadEnvironment(envName)
	if err != nil {
		http.Error(w, "Environment does not exist", http.StatusBadRequest)
		return
	}

	var raw []byte
	r.Body.Read(raw)
	service := new(model.Service)
	err = json.Unmarshal(raw, service)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	service, err = s.db.ServiceStorage.Create(env.Name, service)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out, err := json.Marshal(service)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusContinue)
	fmt.Fprint(w, string(out))
}

func (s *ServiceController) Index(w http.ResponseWriter, r *http.Request) {
	envName := mux.Vars(r)["env_name"]
	env, err := s.loadEnvironment(envName)
	if err != nil {
		http.Error(w, "Environment does not exist", http.StatusBadRequest)
		return
	}

	services, err := s.db.ServiceStorage.List(env.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out, err := json.Marshal(serviceListResponse{Items: services})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(out))
}

func (s *ServiceController) Show(w http.ResponseWriter, r *http.Request) {
	urlVars := mux.Vars(r)
	envName := urlVars["env_name"]
	svcName := urlVars["name"]
	env, err := s.loadEnvironment(envName)
	if err != nil {
		http.Error(w, "Environment does not exist", http.StatusBadRequest)
		return
	}

	service, err := s.db.ServiceStorage.Get(env.Name, svcName)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	out, err := json.Marshal(service)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(out))
}

func (s *ServiceController) Delete(w http.ResponseWriter, r *http.Request) {
	urlVars := mux.Vars(r)
	envName := urlVars["env_name"]
	svcName := urlVars["name"]
	env, err := s.loadEnvironment(envName)
	if err != nil {
		http.Error(w, "Environment does not exist", http.StatusBadRequest)
		return
	}

	if err := s.db.ServiceStorage.Delete(env.Name, svcName); err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// fmt.Fprint(w, string(out))
}
