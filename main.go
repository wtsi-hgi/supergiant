package main

import (
	"log"
	"net/http"
	"supergiant/core/controller"
	"supergiant/core/storage"

	"github.com/julienschmidt/httprouter"
)

func main() {
	storageClient := &storage.Client{Endpoints: []string{"http://localhost:2379"}}
	envStorage := &storage.Environment{Client: storageClient}
	controller := &controller.Environment{Storage: envStorage}

	router := httprouter.New()
	router.POST("/environments", controller.Create)
	router.GET("/environments", controller.Index)
	router.GET("/environments/:name", controller.Show)
	router.DELETE("/environments/:name", controller.Delete)

	log.Fatal(http.ListenAndServe(":8080", router))
}
