package main

import (
	"log"
	"net/http"
	"supergiant/core/controller"
	"supergiant/core/storage"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	// StrictSlash will redirect /environments to /environments/
	// otherwise mux will simply not match /environments/
	router.StrictSlash(true)

	db := storage.NewClient([]string{"http://localhost:2379"})

	controller.NewEnvironmentController(router, db)
	controller.NewServiceController(router, db)

	log.Fatal(http.ListenAndServe(":8080", router))
}
