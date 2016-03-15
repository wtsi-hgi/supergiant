package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"supergiant/core/model"
	"supergiant/core/storage"

	"github.com/gorilla/mux"
)

type AppController struct {
	db *storage.Client
}

type appListResponse struct {
	Items []*model.App `json:"items"`
}

func NewAppController(router *mux.Router, db *storage.Client) *AppController {
	controller := AppController{db: db}
	router.HandleFunc("/apps", controller.Create).Methods("POST")
	router.HandleFunc("/apps", controller.Index).Methods("GET")
	router.HandleFunc("/apps/{name}", controller.Show).Methods("GET")
	router.HandleFunc("/apps/{name}", controller.Delete).Methods("DELETE")
	return &controller
}

func (e *AppController) Create(w http.ResponseWriter, r *http.Request) {
	app := new(model.App)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(app)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app, err = e.db.AppStorage.Create(app)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out, err := json.Marshal(app)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, string(out))
}

func (e *AppController) Index(w http.ResponseWriter, r *http.Request) {
	apps, err := e.db.AppStorage.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out, err := json.Marshal(appListResponse{Items: apps})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(out))
}

func (e *AppController) Show(w http.ResponseWriter, r *http.Request) {
	appName := mux.Vars(r)["name"]
	app, err := e.db.AppStorage.Get(appName)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	out, err := json.Marshal(app)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(out))
}

func (e *AppController) Delete(w http.ResponseWriter, r *http.Request) {
	appName := mux.Vars(r)["name"]
	if err := e.db.AppStorage.Delete(appName); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// fmt.Fprint(w, string(out))
}
