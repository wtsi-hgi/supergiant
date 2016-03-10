package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"supergiant/core/model"
	"supergiant/core/storage"

	"github.com/gorilla/mux"
)

type ComponentController struct {
	db *storage.Client
}

type componentListResponse struct {
	Items []*model.Component `json:"items"`
}

func NewComponentController(router *mux.Router, db *storage.Client) *ComponentController {
	controller := ComponentController{db: db}
	router.HandleFunc("/apps/{app_name}/components", controller.Create).Methods("POST")
	router.HandleFunc("/apps/{app_name}/components", controller.Index).Methods("GET")
	router.HandleFunc("/apps/{app_name}/components/{name}", controller.Show).Methods("GET")
	router.HandleFunc("/apps/{app_name}/components/{name}", controller.Delete).Methods("DELETE")
	return &controller
}

func (s *ComponentController) loadApp(appName string) (*model.App, error) {
	return s.db.AppStorage.Get(appName)
}

func (s *ComponentController) Create(w http.ResponseWriter, r *http.Request) {
	appName := mux.Vars(r)["app_name"]
	app, err := s.loadApp(appName)
	if err != nil {
		http.Error(w, "App does not exist", http.StatusBadRequest)
		return
	}

	component := new(model.Component)
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(component)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	component, err = s.db.ComponentStorage.Create(app.Name, component)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out, err := json.Marshal(component)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, string(out))
}

func (s *ComponentController) Index(w http.ResponseWriter, r *http.Request) {
	appName := mux.Vars(r)["app_name"]
	app, err := s.loadApp(appName)
	if err != nil {
		http.Error(w, "App does not exist", http.StatusBadRequest)
		return
	}

	components, err := s.db.ComponentStorage.List(app.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out, err := json.Marshal(componentListResponse{Items: components})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(out))
}

func (s *ComponentController) Show(w http.ResponseWriter, r *http.Request) {
	urlVars := mux.Vars(r)
	appName := urlVars["app_name"]
	compName := urlVars["name"]
	app, err := s.loadApp(appName)
	if err != nil {
		http.Error(w, "App does not exist", http.StatusBadRequest)
		return
	}

	component, err := s.db.ComponentStorage.Get(app.Name, compName)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	out, err := json.Marshal(component)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(out))
}

func (s *ComponentController) Delete(w http.ResponseWriter, r *http.Request) {
	urlVars := mux.Vars(r)
	appName := urlVars["app_name"]
	compName := urlVars["name"]
	app, err := s.loadApp(appName)
	if err != nil {
		http.Error(w, "App does not exist", http.StatusBadRequest)
		return
	}

	if err := s.db.ComponentStorage.Delete(app.Name, compName); err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// fmt.Fprint(w, string(out))
}
