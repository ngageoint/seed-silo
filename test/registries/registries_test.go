package handlers_registries_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/ngageoint/seed-common/util"
	"github.com/ngageoint/seed-silo/models"
	"github.com/ngageoint/seed-silo/database"
	"github.com/ngageoint/seed-silo/route"
)

var token = ""
var db *sql.DB
var router *mux.Router

func TestMain(m *testing.M) {
	var err error
	os.Remove("./silo-test.db")
	db = database.InitDB("./silo-test.db")
	router, err = route.NewRouter()
	if err != nil {
		os.Remove("./silo-test.db")
		os.Exit(-1)
	}

	util.InitPrinter(util.PrintErr)
	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)

	token, err = login("admin", "spicy-pickles17!")
	if err != nil {
		os.Remove("./silo-test.db")
		os.Exit(-1)
	}

	code := m.Run()

	os.Remove("./silo-test.db")

	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	clearTable()

	cases := []struct {
		urlStr string
	}{
		{"/registries"},
		{"/images"},
		{"/jobs"},
		{"/job-versions"},
	}

	for _, c := range cases {

		req, _ := http.NewRequest("GET", c.urlStr, nil)
		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)

		if body := response.Body.String(); body != "[]" {
			t.Errorf("Expected an empty array. Got %s", body)
		}
	}
}

func TestGetNonExistentItem(t *testing.T) {
	clearTable()

	cases := []struct {
		urlStr   string
		code     int
		auth     bool
		errorMsg string
	}{
		{"/registries/1/scan", 401, false, "Missing authorization token"},
		{"/registries/1/scan", 404, true, "No registry found with that ID"},
		{"/registries/1", 404, true, "No registry found with that ID"},
		{"/registries/badid", 400, true, "Invalid ID"},
		{"/images/1", 404, false, "No image found with that ID"},
		{"/images/badid", 400, false, "Invalid ID"},
		{"/images/1/manifest", 404, false, "No image found with that ID"},
		{"/users/2", 404, false, "No user found with that ID"},
		{"/users/badid", 400, false, "Invalid ID"},
		{"/jobs/1", 404, false, "No job found with that ID"},
		{"/jobs/badid", 400, false, "Invalid ID"},
		{"/job-versions/1", 404, false, "No job version found with that ID"},
		{"/job-versions/badid", 400, false, "Invalid ID"},
	}

	for _, c := range cases {
		req, _ := http.NewRequest("GET", c.urlStr, nil)
		if c.auth {
			req.Header.Set("Authorization", "Token: "+token)
		}
		response := executeRequest(req)

		checkResponseCode(t, c.code, response.Code)

		var m map[string]string
		json.Unmarshal(response.Body.Bytes(), &m)
		if m["error"] != c.errorMsg {
			t.Errorf("Expected the 'error' key of the response to be set to %s. Got '%s'", c.errorMsg, m["error"])
		}
	}
}

func TestAddRegistry(t *testing.T) {
	clearTable()

	payload := []byte(`{"name":"dockerhub", "url":"https://hub.docker.com", "org":"johnptobe", "username":"", "password": ""}`)

	req, _ := http.NewRequest("POST", "/registries/add", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Token: "+token)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["ID"] != 1.0 {
		t.Errorf("Expected registry ID to be '1'. Got '%v'", m["ID"])
	}

	if m["Name"] != "dockerhub" {
		t.Errorf("Expected registry name to be 'dockerhub'. Got '%v'", m["Name"])
	}

	if m["Url"] != "https://hub.docker.com" {
		t.Errorf("Expected url to be 'https://hub.docker.com'. Got '%v'", m["Url"])
	}

	if m["Org"] != "johnptobe" {
		t.Errorf("Expected org to be 'johnptobe'. Got '%v'", m["Org"])
	}
}

func TestDeleteRegistry(t *testing.T) {
	clearTable()

	addRegistry()

	payload := []byte(``)
	req, _ := http.NewRequest("DELETE", "/registries/delete/1", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Token: "+token)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/registries/1/scan", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Token: "+token)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	errorStr := "No registry found with that ID"
	if m["error"] != errorStr {
		t.Errorf("Expected error to be '%s'. Got '%v'", errorStr, m["error"])
	}

	req, _ = http.NewRequest("DELETE", "/registries/delete/test", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Token: "+token)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestScanRegistry(t *testing.T) {
	clearTable()

	addRegistry()

	payload := []byte(``)
	req, _ := http.NewRequest("GET", "/registries/1/scan", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Token: "+token)
	response := executeRequest(req)

	checkResponseCode(t, 202, response.Code)

	req, _ = http.NewRequest("GET", "/images", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := []models.SimpleImage{}
	json.Unmarshal(response.Body.Bytes(), &m)

	testImage := models.SimpleImage{ID: 1, RegistryId: 1, Name: "my-job-0.1.0-seed:0.1.0",
		Registry: "docker.io", Org: "johnptobe", JobName: "my-job", Title: "My first job",
		Maintainer: "John Doe", Email: "jdoe@example.com", MaintOrg: "E-corp",
		Description: "Reads an HDF5 file and outputs two TIFF images, a CSV and manifest containing cell_count",
		JobVersion:  "0.1.0", PackageVersion: "0.1.0"}
	if fmt.Sprint(m[0]) != fmt.Sprint(testImage) {
		t.Errorf("Expected image to be %v. Got '%v'", testImage, m[0])
	}

	req, _ = http.NewRequest("GET", "/registries/test/scan", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Token: "+token)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestListRegistries(t *testing.T) {
	clearTable()

	addRegistry()

	payload := []byte(``)
	req, _ := http.NewRequest("GET", "/registries/scan", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Token: "+token)
	response := executeRequest(req)

	checkResponseCode(t, 202, response.Code)

	req, _ = http.NewRequest("GET", "/registries", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func clearTable() {
	db.Exec("DELETE FROM RegistryInfo")
	db.Exec("DELETE FROM Image")
	db.Exec("DELETE FROM sqlite_sequence")
	db.Exec("DELETE FROM Job")
	db.Exec("DELETE FROM JobVersion")
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func addRegistry() {
	payload := []byte(`{"name":"dockerhub", "url":"https://hub.docker.com", "org":"johnptobe", "username":"", "password": ""}`)

	req, _ := http.NewRequest("POST", "/registries/add", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Token: "+token)
	executeRequest(req)
}

func login(username, password string) (string, error) {
	payload := []byte(`{"username":"` + username + `", "password": "` + password + `"}`)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	response := executeRequest(req)

	if response.Code != 200 {
		return "", errors.New("Login error")
	}

	m := map[string]string{}
	json.Unmarshal(response.Body.Bytes(), &m)

	return m["token"], nil
}
