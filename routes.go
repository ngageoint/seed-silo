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
		"Registry",
		"GET",
		"/registry/{id}",
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
		"Image",
		"GET",
		"/images/{id}",
	},
	Route{
		"ImageManifest",
		"GET",
		"/images/{id}/manifest",
	},
	Route{
		"Login",
		"POST",
		"/login",
	},
	Route{
		"AddUser",
		"POST",
		"/user/add",
	},
}

func GetRoutes() []Route {
	return routes
}
