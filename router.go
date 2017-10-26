package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

var handler = map[string]http.HandlerFunc{
	"Index": Index,
	"AddRegistry": AddRegistry,
	"DeleteRegistry": DeleteRegistry,
	"ListRegistries": ListRegistries,
	"ScanRegistries": ScanRegistries,
	"ScanRegistry": ScanRegistry,
	"ListImages": ListImages,
	"SearchImages": SearchImages,
	"ImageManifest": ImageManifest,
}

func NewRouter() (*mux.Router, error) {

	if len(handler) != len(routes) {
		fmt.Println("Error creating router. Please check that there is a handler for each route.")
		fmt.Println("Handlers:")
		fmt.Println(handler)
		fmt.Println("Routes:")
		fmt.Println(routes)
		return nil, errors.New("Router/Handler length mismatch")
	}

	var err error = nil
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		if handler[route.Name] == nil {
			fmt.Println("Unable to find handler for route %s", route.Name)
			err = errors.New("Unable to find handler for route")
			continue
		}

		logHandler := Logger(handler[route.Name], route.Name)
		router.
		Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(logHandler)
	}

	return router, err
}