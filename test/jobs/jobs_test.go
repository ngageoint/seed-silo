package handlers_jobs_test

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
	"github.com/ngageoint/seed-silo/database"
	"github.com/ngageoint/seed-silo/models"
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

	if get_images() == false {
		os.Remove("./silo-test.db")
		os.Exit(-1)
	}

	code := m.Run()

	os.Remove("./silo-test.db")

	os.Exit(code)
}

func TestSearchJobs(t *testing.T) {
	payload := []byte(``)
	req, _ := http.NewRequest("GET", "/jobs/search/my-job-0.1.0", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := make(map[int]models.Job)
	json.Unmarshal(response.Body.Bytes(), &m)

	searchResult := m[1]
	searchResult.JobVersions = nil

	testJob := models.Job{ID: 1, Name: "my-job", LatestJobVersion: "1.0.0", LatestPackageVersion: "0.1.0",
		Title: "My first job", Maintainer: "John Doe", Email: "jdoe@example.com", MaintOrg: "E-corp",
		Description: "Reads an HDF5 file and outputs two TIFF images, a CSV and manifest containing cell_count", ImageIDs: []int{1}}

	if fmt.Sprint(searchResult) != fmt.Sprint(testJob) {
		t.Errorf("Expected image to be %v. Got '%v'", testJob, searchResult)
	}

	req, _ = http.NewRequest("GET", "/jobs/search/asdfasdf", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m2 := make(map[int]models.Job)
	json.Unmarshal(response.Body.Bytes(), &m2)

	if len(m2) != 0 {
		t.Errorf("Expected emtpy job list. Got %d results.", len(m2))
	}
}

func TestJob(t *testing.T) {
	payload := []byte(``)

	req, _ := http.NewRequest("GET", "/jobs/1", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := models.Job{}
	json.Unmarshal(response.Body.Bytes(), &m)

	m.JobVersions = nil

	testJob := models.Job{ID: 1, Name: "my-job", LatestJobVersion: "1.0.0", LatestPackageVersion: "0.1.0",
		Title: "My first job", Maintainer: "John Doe", Email: "jdoe@example.com", MaintOrg: "E-corp",
		Description: "Reads an HDF5 file and outputs two TIFF images, a CSV and manifest containing cell_count", ImageIDs: []int{1, 3, 5}}

	mStr := fmt.Sprintf("%v", m)
	testStr := fmt.Sprintf("%v", testJob)
	if mStr != testStr {
		t.Errorf("Expected manifest to be %v. Got '%v'", testJob, m)
	}
}

func TestListJobs(t *testing.T) {
	payload := []byte(``)

	req, _ := http.NewRequest("GET", "/jobs", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := []models.Job{}
	json.Unmarshal(response.Body.Bytes(), &m)

	m[0].JobVersions = nil

	testJob := models.Job{ID: 1, Name: "my-job", LatestJobVersion: "1.0.0", LatestPackageVersion: "0.1.0",
		Title: "My first job", Maintainer: "John Doe", Email: "jdoe@example.com", MaintOrg: "E-corp",
		Description: "Reads an HDF5 file and outputs two TIFF images, a CSV and manifest containing cell_count", ImageIDs: []int{1, 3, 5}}

	mStr := fmt.Sprintf("%v", m[0])
	testStr := fmt.Sprintf("%v", testJob)
	if mStr != testStr {
		t.Errorf("Expected manifest to be %v. Got '%v'", testJob, m[0])
	}
}

func TestJobVersion(t *testing.T) {
	payload := []byte(``)

	req, _ := http.NewRequest("GET", "/job-versions/1", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := models.JobVersion{}
	json.Unmarshal(response.Body.Bytes(), &m)

	testImage := models.SimpleImage{ID: 1, RegistryId: 1, Name: "my-job-0.1.0-seed:0.1.0",
		Registry: "docker.io", Org: "johnptobe", JobName: "my-job", Title: "My first job",
		Maintainer: "John Doe", Email: "jdoe@example.com", MaintOrg: "E-corp",
		Description: "Reads an HDF5 file and outputs two TIFF images, a CSV and manifest containing cell_count",
		JobVersion:  "0.1.0", PackageVersion: "0.1.0"}

	testJobVersion := models.JobVersion{ID: 1, JobId: 1, JobName: "my-job", JobVersion: "0.1.0", LatestPackageVersion: "0.1.0"}
	testJobVersion.Images = append(testJobVersion.Images, testImage)

	mStr := fmt.Sprintf("%v", m)
	testStr := fmt.Sprintf("%v", testJobVersion)
	if mStr != testStr {
		t.Errorf("Expected manifest to be %v. Got '%v'", testJobVersion, m)
	}
}

func TestJobVersions(t *testing.T) {
	payload := []byte(``)

	req, _ := http.NewRequest("GET", "/jobs/1/job-versions", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := []models.JobVersion{}
	json.Unmarshal(response.Body.Bytes(), &m)

	testImage := models.SimpleImage{ID: 1, RegistryId: 1, Name: "my-job-0.1.0-seed:0.1.0",
		Registry: "docker.io", Org: "johnptobe", JobName: "my-job", Title: "My first job",
		Maintainer: "John Doe", Email: "jdoe@example.com", MaintOrg: "E-corp",
		Description: "Reads an HDF5 file and outputs two TIFF images, a CSV and manifest containing cell_count",
		JobVersion:  "0.1.0", PackageVersion: "0.1.0"}

	testJobVersion := models.JobVersion{ID: 1, JobId: 1, JobName: "my-job", JobVersion: "0.1.0", LatestPackageVersion: "0.1.0"}
	testJobVersion.Images = append(testJobVersion.Images, testImage)

	mStr := fmt.Sprintf("%v", m[0])
	testStr := fmt.Sprintf("%v", testJobVersion)
	if mStr != testStr {
		t.Errorf("Expected manifest to be %v. Got '%v'", testJobVersion, m[0])
	}
}

func TestListJobVersions(t *testing.T) {
	payload := []byte(``)

	req, _ := http.NewRequest("GET", "/job-versions", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := []models.JobVersion{}
	json.Unmarshal(response.Body.Bytes(), &m)

	testImage := models.SimpleImage{ID: 1, RegistryId: 1, Name: "my-job-0.1.0-seed:0.1.0",
		Registry: "docker.io", Org: "johnptobe", JobName: "my-job", Title: "My first job",
		Maintainer: "John Doe", Email: "jdoe@example.com", MaintOrg: "E-corp",
		Description: "Reads an HDF5 file and outputs two TIFF images, a CSV and manifest containing cell_count",
		JobVersion:  "0.1.0", PackageVersion: "0.1.0"}

	testJobVersion := models.JobVersion{ID: 1, JobId: 1, JobName: "my-job", JobVersion: "0.1.0", LatestPackageVersion: "0.1.0"}
	testJobVersion.Images = append(testJobVersion.Images, testImage)

	mStr := fmt.Sprintf("%v", m[0])
	testStr := fmt.Sprintf("%v", testJobVersion)
	if mStr != testStr {
		t.Errorf("Expected manifest to be %v. Got '%v'", testJobVersion, m[0])
	}
}

func get_images() bool {
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
