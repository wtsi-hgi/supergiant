package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"supergiant/core/model"
	"supergiant/core/storage"

	"github.com/gorilla/mux"
)

type EnvironmentController struct {
	db *storage.Client
}

type environmentListResponse struct {
	Items []*model.Environment `json:"items"`
}

func NewEnvironmentController(router *mux.Router, db *storage.Client) *EnvironmentController {
	controller := EnvironmentController{db: db}
	router.HandleFunc("/environments", controller.Create).Methods("POST")
	router.HandleFunc("/environments", controller.Index).Methods("GET")
	router.HandleFunc("/environments/{name}", controller.Show).Methods("GET")
	router.HandleFunc("/environments/{name}", controller.Delete).Methods("DELETE")
	return &controller
}

func (e *EnvironmentController) Create(w http.ResponseWriter, r *http.Request) {
	var raw []byte
	r.Body.Read(raw)
	environment := new(model.Environment)
	err := json.Unmarshal(raw, environment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	environment, err = e.db.EnvironmentStorage.Create(environment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out, err := json.Marshal(environment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusContinue)
	fmt.Fprint(w, string(out))
}

func (e *EnvironmentController) Index(w http.ResponseWriter, r *http.Request) {
	environments, err := e.db.EnvironmentStorage.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out, err := json.Marshal(environmentListResponse{Items: environments})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(out))
}

func (e *EnvironmentController) Show(w http.ResponseWriter, r *http.Request) {
	envName := mux.Vars(r)["name"]
	environment, err := e.db.EnvironmentStorage.Get(envName)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	out, err := json.Marshal(environment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(out))
}

func (e *EnvironmentController) Delete(w http.ResponseWriter, r *http.Request) {
	envName := mux.Vars(r)["name"]
	if err := e.db.EnvironmentStorage.Delete(envName); err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// fmt.Fprint(w, string(out))
}
