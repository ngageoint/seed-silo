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
		"ListJobs",
		"GET",
		"/jobs",
	},
	Route{
		"Job",
		"GET",
		"/jobs/{id}",
	},
	Route{
		"Login",
		"POST",
		"/login",
	},
	Route{
		"User",
		"GET",
		"/user/{id}",
	},
	Route{
		"AddUser",
		"POST",
		"/user/add",
	},
	Route{
		"DeleteUser",
		"DELETE",
		"/user/delete/{id}",
	},
	Route{
		"ListUsers",
		"GET",
		"/users",
	},
	Route{
		"PreflightOptions",
		"OPTIONS",
		"/*",
	},
}

func GetRoutes() []Route {
	return routes
}
