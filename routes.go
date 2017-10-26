package main

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

func GetRoutes() []Route {
	return routes
}
