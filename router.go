package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

var handler = map[string]http.HandlerFunc{
	"Index": Index,
	"Registry": Registry,
	"AddRegistry": Validate([]string{"admin"}, AddRegistry),
	"DeleteRegistry": Validate([]string{"admin"}, DeleteRegistry),
	"ListRegistries": ListRegistries,
	"ScanRegistries": Validate([]string{"admin"}, ScanRegistries),
	"ScanRegistry": Validate([]string{"admin"}, ScanRegistry),
	"ListImages": ListImages,
	"SearchImages": SearchImages,
	"Image": Image,
	"ImageManifest": ImageManifest,
	"Login": Login,
	"User": User,
	"AddUser": Validate([]string{"admin"}, AddUser),
	"DeleteUser": Validate([]string{"admin"}, DeleteUser),
	"ListUsers": ListUsers,
	"PreflightOptions": PreflightOptions,
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
			msg := fmt.Sprintf("Unable to find handler for route %s", route.Name)
			err = errors.New(msg)
			continue
		}

		// handle all CORS preflight options
		po := "PreflightOptions"
		options := Logger(handler[po], po)
		router.Methods("OPTIONS").Handler(options)

		logHandler := Logger(handler[route.Name], route.Name)
		router.
		Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(logHandler)
	}

	return router, err
}