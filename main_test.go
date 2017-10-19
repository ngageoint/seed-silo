package main

import (
	"os"
	"testing"

	"net/http"
	"net/http/httptest"
	"encoding/json"
	"bytes"
	"fmt"
)

func TestMain(m *testing.M) {
	InitDB("./silo-test.db")

	code := m.Run()

	os.Remove("./silo-test.db")

	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	clearTable()

	cases := []struct {
		urlStr        string
	}{
		{"/registries"},
		{"/images"},
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
		urlStr        string
		code int
		errorMsg string
	}{
		{"/registry/1/scan", 404, "No registry found with that ID"},
		{"/images/1/manifest", 404, "No image found with that ID"},
	}

	for _, c := range cases {
		req, _ := http.NewRequest("GET", c.urlStr, nil)
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

	req, _ := http.NewRequest("POST", "/registry/add", bytes.NewBuffer(payload))
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
		t.Errorf("Expected product name to be 'dockerhub'. Got '%v'", m["Name"])
	}

	if m["Url"] != "https://hub.docker.com" {
		t.Errorf("Expected url to be 'https://hub.docker.com'. Got '%v'", m["Url"])
	}

	if m["Org"] != "johnptobe" {
		t.Errorf("Expected org to be 'johnptobe'. Got '%v'", m["Org"])
	}
}

func clearTable() {
	db := GetDb()
	db.Exec("DELETE FROM RegistryInfo")
	db.Exec("DELETE FROM Image")
	db.Exec("DELETE FROM sqlite_sequence")
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	GetRouter().ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}