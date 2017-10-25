package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
}

var routes = []Route{
	Route{
		"Index",
		"GET",
		"/",
	},
	Route{
		"AddRegistry",
		"POST",
		"/registry/add",
	},
	Route{
		"DeleteRegistry",
		"DELETE",
		"/registry/delete/{id}",
	},
	Route{
		"ListRegistries",
		"GET",
		"/registries",
	},
	Route{
		"ScanRegistries",
		"GET",
		"/registries/scan",
	},
	Route{
		"ScanRegistry",
		"GET",
		"/registry/{id}/scan",
	},
	Route{
		"ListImages",
		"GET",
		"/images",
	},
	Route{
		"SearchImages",
		"GET",
		"/images/search/{query}",
	},
	Route{
		"ImageManifest",
		"GET",
		"/images/{id}/manifest",
	},
}

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

func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler[route.Name])
	}

	return router
}


func GetRoutes() []Route {
	return routes
}
