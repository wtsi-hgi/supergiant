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
	// StrictSlash will redirect /apps to /apps/
	// otherwise mux will simply not match /apps/
	router.StrictSlash(true)

	db := storage.NewClient([]string{"http://localhost:2379"})

	controller.NewAppController(router, db)
	controller.NewComponentController(router, db)
	controller.NewDeploymentController(router, db)
	controller.NewInstanceController(router, db)
	controller.NewReleaseController(router, db)

	log.Fatal(http.ListenAndServe(":8080", router))
}
