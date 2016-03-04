package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"supergiant/core/model"
	"supergiant/core/storage"

	"github.com/julienschmidt/httprouter"
)

type Environment struct {
	Storage *storage.Environment
}

type environmentListResponse struct {
	Items []*model.Environment `json:"items"`
}

func (e *Environment) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var raw []byte
	r.Body.Read(raw)
	environment := new(model.Environment)
	err := json.Unmarshal(raw, environment)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	environment, err = e.Storage.Create(environment)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	out, err := json.Marshal(environment)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusContinue)
	fmt.Fprint(w, string(out))
}

func (e *Environment) Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	environments, err := e.Storage.List()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	out, err := json.Marshal(environmentListResponse{Items: environments})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(out))
}

func (e *Environment) Show(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	environment, err := e.Storage.Get(ps.ByName("name"))
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	out, err := json.Marshal(environment)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(out))
}

func (e *Environment) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := e.Storage.Delete(ps.ByName("name")); err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// fmt.Fprint(w, string(out))
}
