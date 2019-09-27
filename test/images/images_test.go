package handlers_images_test

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
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/ngageoint/seed-common/objects"
	"github.com/ngageoint/seed-common/util"
	"github.com/ngageoint/seed-silo/models"
	"github.com/ngageoint/seed-silo/database"
	"github.com/ngageoint/seed-silo/route"
)

var token = ""
var db *sql.DB
var router *mux.Router
var imageID int

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func TestMain(m *testing.M) {
	var err error
	os.Remove("./silo-test.db")
	db = database.InitSqliteDB("./silo-test.db", "admin", "spicy-pickles17!")
	router, err = route.NewRouter()
	if err != nil {
		os.Remove("./silo-test.db")
		os.Exit(-1)
	}

	util.InitPrinter(util.Quiet, nil, nil)
	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)

	token, err = login("admin", "spicy-pickles17!")
	if err != nil {
		os.Remove("./silo-test.db")
		os.Exit(-1)
	}

	if get_images() == false {
		os.Remove("./silo-test.db")
		os.Exit(-1)
	}

	imageID = findTestImageID()

	code := m.Run()

	os.Remove("./silo-test.db")

	// Run same tests with Postgres
	url := getEnv("DATABASE_URL", "postgres://scale:scale@localhost:55432/test_silo?sslmode=disable")
	base := strings.Replace(url, "test_silo", "", 1)
	full := strings.Replace(url, "test_silo", "test_silo_image", 1)
	database.CreatePostgresDB(base, "test_silo_image")
	db = database.InitPostgresDB(full, "admin", "spicy-pickles17!")

	token, err = login("admin", "spicy-pickles17!")
	if err != nil {
		database.RemovePostgresDB(base, "test_silo_image")
		os.Exit(-1)
	}

	if get_images() == false {
		database.RemovePostgresDB(base, "test_silo_image")
		os.Exit(-1)
	}

	code += m.Run()

	database.RemovePostgresDB(base, "test_silo_image")

	os.Exit(code)
}

func TestSearchImages(t *testing.T) {
	payload := []byte(``)
	req, _ := http.NewRequest("GET", "/images/search/my-job-0.1.0", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := []models.SimpleImage{}
	json.Unmarshal(response.Body.Bytes(), &m)

	testImage := models.SimpleImage{ID: imageID, RegistryId: 1, Name: "my-job-0.1.0-seed:0.1.0",
		Registry: "docker.io", Org: "geointseed", JobName: "my-job", Title: "My first job",
		Maintainer: "John Doe", Email: "jdoe@example.com", MaintOrg: "E-corp",
		Description: "Reads an HDF5 file and outputs two TIFF images, a CSV and manifest containing cell_count",
		JobVersion:  "0.1.0", PackageVersion: "0.1.0"}
	if fmt.Sprint(m[0]) != fmt.Sprint(testImage) {
		t.Errorf("Expected image to be %v. Got '%v'", testImage, m[0])
	}

	req, _ = http.NewRequest("GET", "/images/search/asdfasdf", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
	json.Unmarshal(response.Body.Bytes(), &m)

	if len(m) != 0 {
		t.Errorf("Expected emtpy image list. Got %d results.", len(m))
	}
}

func TestImageManifest(t *testing.T) {
	payload := []byte(``)
	url := fmt.Sprintf("/images/%d/manifest", imageID)
	req, _ := http.NewRequest("GET", url, bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := objects.Seed{}
	json.Unmarshal(response.Body.Bytes(), &m)

	testManifest := objects.SeedFromManifestFile("../seed.manifest.json")

	mStr := fmt.Sprintf("%v", m)
	testStr := fmt.Sprintf("%v", testManifest)
	if mStr != testStr {
		t.Errorf("Expected manifest to be %v. Got '%v'", testManifest, m)
	}
}

func TestJITImageManifest(t *testing.T) {
	payload := []byte(``)
	url := "/images/manifest/docker.io/geointseed/my-job-0.1.0-seed:0.1.0"
	req, _ := http.NewRequest("GET", url, bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := objects.Seed{}
	json.Unmarshal(response.Body.Bytes(), &m)

	testManifest := objects.SeedFromManifestFile("../seed.manifest.json")

	mStr := fmt.Sprintf("%v", m)
	testStr := fmt.Sprintf("%v", testManifest)
	if mStr != testStr {
		t.Errorf("Expected manifest to be %v. Got '%v'", testManifest, m)
	}
}

func TestListImages(t *testing.T) {
	payload := []byte(``)
	req, _ := http.NewRequest("GET", "/images", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func get_images() bool {
	clearTablePG()
	clearTable()

	addRegistry()

	payload := []byte(``)
	req, _ := http.NewRequest("GET", "/registries/scan", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Token: "+token)
	response := executeRequest(req)

	return response.Code == 202
}

func clearTable() {
	db.Exec("DELETE FROM RegistryInfo")
	db.Exec("DELETE FROM Image")
	db.Exec("DELETE FROM sqlite_sequence")
	db.Exec("DELETE FROM Job")
	db.Exec("DELETE FROM JobVersion")
}

func clearTablePG() {
	db.Exec("TRUNCATE RegistryInfo RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE Image RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE SiloUser RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE Job RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE JobVersion RESTART IDENTITY CASCADE")
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
	payload := []byte(`{"name":"dockerhub", "url":"https://hub.docker.com", "org":"geointseed", "username":"", "password": ""}`)

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

func findTestImageID() int {
	payload := []byte(``)
	req, _ := http.NewRequest("GET", "/images", bytes.NewBuffer(payload))
	response := executeRequest(req)

	if response.Code != 200 {
		return -1
	}

	images := []models.SimpleImage{}
	err :=json.Unmarshal(response.Body.Bytes(), &images)

	if err != nil {
		return -1
	}

	var m models.SimpleImage
	for _, img := range images {
		if img.Name == "my-job-0.1.0-seed:0.1.0"{
			m = img
		}
	}

	return m.ID
}