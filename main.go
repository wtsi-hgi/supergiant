package main

import (
	"guber"
	"log"
	"net/http"
	"os"
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

	var (
		kHost = os.Getenv("K_HOST")
		kUser = os.Getenv("K_USER")
		kPass = os.Getenv("K_PASS")
	)
	kube := guber.NewClient(kHost, kUser, kPass)

	controller.NewAppController(router, db)
	controller.NewComponentController(router, db)
	controller.NewDeploymentController(router, db)
	controller.NewInstanceController(router, db)
	controller.NewReleaseController(router, db)

	log.Fatal(http.ListenAndServe(":8080", router))
}
