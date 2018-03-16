package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ngageoint/seed-common/util"
)

type Route struct {
	Name    string
	Method  string
	Pattern string
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
		"/registries/add",
	},
	Route{
		"DeleteRegistry",
		"DELETE",
		"/registries/delete/{id}",
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
		"Registry",
		"GET",
		"/registries/{id}",
	},
	Route{
		"ScanRegistry",
		"GET",
		"/registries/{id}/scan",
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
		"SearchJobs",
		"GET",
		"/jobs/search/{query}",
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
		"JobVersions",
		"GET",
		"/jobs/{id}/job-versions",
	},
	Route{
		"ListJobVersions",
		"GET",
		"/job-versions",
	},
	Route{
		"JobVersion",
		"GET",
		"/job-versions/{id}",
	},
	Route{
		"Login",
		"POST",
		"/login",
	},
	Route{
		"User",
		"GET",
		"/users/{id}",
	},
	Route{
		"AddUser",
		"POST",
		"/users/add",
	},
	Route{
		"DeleteUser",
		"DELETE",
		"/users/delete/{id}",
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

func Index(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, GetRoutes())
}

func PreflightOptions(w http.ResponseWriter, r *http.Request) {
	// return 200 OK for all preflight CORS requests
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
	return
}

func checkError(err error, url, username, password string) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	humanError := ""

	if strings.Contains(errStr, "status=401") {
		if username == "" || password == "" {
			humanError = "The specified registry requires a login.  Please try again with a username (-u) and password (-p)."
		} else {
			humanError = "Incorrect username/password."
		}
	} else if strings.Contains(errStr, "status=404") {
		humanError = "Connected to registry but received a 404 error. Please check the url and try again."
	} else {
		humanError = "Could not connect to the specified registry. Please check the url and try again."
	}
	return humanError
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)

	util.PrintUtil("Response: %s\n", response)
}
