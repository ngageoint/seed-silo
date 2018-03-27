package route

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ngageoint/seed-silo/handlers"
	"github.com/ngageoint/seed-silo/logger"
)

var handler = map[string]http.HandlerFunc{
	"Index": handlers.Index,
	"AddRegistry": handlers.Validate([]string{"admin"}, handlers.AddRegistry),
	"DeleteRegistry": handlers.Validate([]string{"admin"}, handlers.DeleteRegistry),
	"ListRegistries": handlers.ListRegistries,
	"ScanRegistries": handlers.Validate([]string{"admin"}, handlers.ScanRegistries),
	"Registry": handlers.Registry,
	"ScanRegistry": handlers.Validate([]string{"admin"}, handlers.ScanRegistry),
	"ListImages": handlers.ListImages,
	"SearchImages": handlers.SearchImages,
	"SearchJobs": handlers.SearchJobs,
	"Image": handlers.Image,
	"ImageManifest": handlers.ImageManifest,
	"ListJobs": handlers.ListJobs,
	"Job": handlers.Job,
	"JobVersions": handlers.JobVersions,
	"ListJobVersions": handlers.ListJobVersions,
	"JobVersion": handlers.JobVersion,
	"Login": handlers.Login,
	"User": handlers.User,
	"AddUser": handlers.Validate([]string{"admin"}, handlers.AddUser),
	"DeleteUser": handlers.Validate([]string{"admin"}, handlers.DeleteUser),
	"ListUsers": handlers.ListUsers,
	"PreflightOptions": handlers.PreflightOptions,
}

func NewRouter() (*mux.Router, error) {
    routes := handlers.GetRoutes()
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
		options := logger.Logger(handler[po], po)
		router.Methods("OPTIONS").Handler(options)

		logHandler := logger.Logger(handler[route.Name], route.Name)
		router.
		Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(logHandler)
	}

	return router, err
}