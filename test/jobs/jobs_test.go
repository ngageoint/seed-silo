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
	"strings"
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
var JobID int
var JVID int
var imageID int
var imageIDs []int

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

	JobID = findTestJobID()
	JVID = findTestJobVersionID()
	imageID = findTestImageID()

	code := m.Run()

	os.Remove("./silo-test.db")

	// Run same tests with Postgres
	url := getEnv("DATABASE_URL", "postgres://scale:scale@localhost:55432/test_silo?sslmode=disable")
	base := strings.Replace(url, "test_silo", "", 1)
	full := strings.Replace(url, "test_silo", "test_silo_job", 1)
	database.CreatePostgresDB( base, "test_silo_job")
	db = database.InitPostgresDB(full, "admin", "spicy-pickles17!")

	token, err = login("admin", "spicy-pickles17!")
	if err != nil {
		database.RemovePostgresDB(base, "test_silo_job")
		os.Exit(-1)
	}

	if get_images() == false {
		database.RemovePostgresDB(base, "test_silo_job")
		os.Exit(-1)
	}

	JobID = findTestJobID()
	JVID = findTestJobVersionID()
	imageID = findTestImageID()

	code += m.Run()

	database.RemovePostgresDB(base, "test_silo_job")

	os.Exit(code)
}

func TestSearchJobs(t *testing.T) {
	payload := []byte(``)
	req, _ := http.NewRequest("GET", "/jobs/search/my-job-0.1.0", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := make(map[int]models.Job)
	json.Unmarshal(response.Body.Bytes(), &m)

	searchResult := m[JobID]
	searchResult.JobVersions = nil
	searchResult.ImageIDs = nil

	testJob := models.Job{ID: JobID, Name: "my-job", LatestJobVersion: "1.0.0", LatestPackageVersion: "0.1.0",
		Title: "My first job", Maintainer: "John Doe", Email: "jdoe@example.com", MaintOrg: "E-corp",
		Description: "Reads an HDF5 file and outputs two TIFF images, a CSV and manifest containing cell_count", ImageIDs: nil}

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

	url := fmt.Sprintf("/jobs/%d", JobID)
	req, _ := http.NewRequest("GET", url, bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := models.Job{}
	json.Unmarshal(response.Body.Bytes(), &m)

	m.JobVersions = nil

	testJob := models.Job{ID: JobID, Name: "my-job", LatestJobVersion: "1.0.0", LatestPackageVersion: "0.1.0",
		Title: "My first job", Maintainer: "John Doe", Email: "jdoe@example.com", MaintOrg: "E-corp",
		Description: "Reads an HDF5 file and outputs two TIFF images, a CSV and manifest containing cell_count", ImageIDs: imageIDs}

	mStr := fmt.Sprintf("%v", m)
	testStr := fmt.Sprintf("%v", testJob)
	if mStr != testStr {
		t.Errorf("Expected job to be %v. Got '%v'", testJob, m)
	}
}

func TestListJobs(t *testing.T) {
	payload := []byte(``)

	req, _ := http.NewRequest("GET", "/jobs", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := []models.Job{}
	json.Unmarshal(response.Body.Bytes(), &m)

	m[JobID-1].JobVersions = nil

	testJob := models.Job{ID: JobID, Name: "my-job", LatestJobVersion: "1.0.0", LatestPackageVersion: "0.1.0",
		Title: "My first job", Maintainer: "John Doe", Email: "jdoe@example.com", MaintOrg: "E-corp",
		Description: "Reads an HDF5 file and outputs two TIFF images, a CSV and manifest containing cell_count", ImageIDs: imageIDs}

	mStr := fmt.Sprintf("%v", m[JobID-1])
	testStr := fmt.Sprintf("%v", testJob)
	if mStr != testStr {
		t.Errorf("Expected job #%d to be %v. Got '%v'", JobID, testJob, m[JobID-1])
	}
}

func TestJobVersion(t *testing.T) {
	payload := []byte(``)

	url := fmt.Sprintf("/job-versions/%d", JVID)
	req, _ := http.NewRequest("GET", url, bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := models.JobVersion{}
	json.Unmarshal(response.Body.Bytes(), &m)

	testImage := models.SimpleImage{ID: imageID, RegistryId: 1, Name: "my-job-0.1.0-seed:0.1.0",
		Registry: "docker.io", Org: "geointseed", JobName: "my-job", Title: "My first job",
		Maintainer: "John Doe", Email: "jdoe@example.com", MaintOrg: "E-corp",
		Description: "Reads an HDF5 file and outputs two TIFF images, a CSV and manifest containing cell_count",
		JobVersion:  "0.1.0", PackageVersion: "0.1.0"}

	testJobVersion := models.JobVersion{ID: JVID, JobId: JobID, JobName: "my-job", JobVersion: "0.1.0", LatestPackageVersion: "0.1.0"}
	testJobVersion.Images = append(testJobVersion.Images, testImage)

	mStr := fmt.Sprintf("%v", m)
	testStr := fmt.Sprintf("%v", testJobVersion)
	if mStr != testStr {
		t.Errorf("Expected job version to be %v. Got '%v'", testJobVersion, m)
	}
}

func TestJobVersions(t *testing.T) {
	payload := []byte(``)

	url := fmt.Sprintf("/jobs/%d/job-versions", JobID)
	req, _ := http.NewRequest("GET", url, bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := []models.JobVersion{}
	json.Unmarshal(response.Body.Bytes(), &m)

	testImage := models.SimpleImage{ID: imageID, RegistryId: 1, Name: "my-job-0.1.0-seed:0.1.0",
		Registry: "docker.io", Org: "geointseed", JobName: "my-job", Title: "My first job",
		Maintainer: "John Doe", Email: "jdoe@example.com", MaintOrg: "E-corp",
		Description: "Reads an HDF5 file and outputs two TIFF images, a CSV and manifest containing cell_count",
		JobVersion:  "0.1.0", PackageVersion: "0.1.0"}

	testJobVersion := models.JobVersion{ID: JVID, JobId: JobID, JobName: "my-job", JobVersion: "0.1.0", LatestPackageVersion: "0.1.0"}
	testJobVersion.Images = append(testJobVersion.Images, testImage)

	mStr := fmt.Sprintf("%v", m[0])
	testStr := fmt.Sprintf("%v", testJobVersion)
	if mStr != testStr {
		t.Errorf("Expected job version #%d to be %v. Got '%v'", JVID, testJobVersion, m[0])
	}
}

func TestListJobVersions(t *testing.T) {
	payload := []byte(``)

	req, _ := http.NewRequest("GET", "/job-versions", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	jvs := []models.JobVersion{}
	json.Unmarshal(response.Body.Bytes(), &jvs)

	m := jvs[JVID-1]

	testImage := models.SimpleImage{ID: imageID, RegistryId: 1, Name: "my-job-0.1.0-seed:0.1.0",
		Registry: "docker.io", Org: "geointseed", JobName: "my-job", Title: "My first job",
		Maintainer: "John Doe", Email: "jdoe@example.com", MaintOrg: "E-corp",
		Description: "Reads an HDF5 file and outputs two TIFF images, a CSV and manifest containing cell_count",
		JobVersion:  "0.1.0", PackageVersion: "0.1.0"}

	testJobVersion := models.JobVersion{ID: JVID, JobId: JobID, JobName: "my-job", JobVersion: "0.1.0", LatestPackageVersion: "0.1.0"}
	testJobVersion.Images = append(testJobVersion.Images, testImage)

	mStr := fmt.Sprintf("%v", m)
	testStr := fmt.Sprintf("%v", testJobVersion)
	if mStr != testStr {
		t.Errorf("Expected job version #%d to be %v. Got '%v'", JVID, testJobVersion, m)
	}
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

func findTestJobID() int {
	payload := []byte(``)
	req, _ := http.NewRequest("GET", "/jobs", bytes.NewBuffer(payload))
	response := executeRequest(req)

	if response.Code != 200 {
		util.PrintUtil("JOBID ERROR: %d\n\n", response.Code)
		util.PrintUtil("response: %v\n\n", response.Body)
		return -1
	}

	jobs := []models.Job{}
	err :=json.Unmarshal(response.Body.Bytes(), &jobs)

	if err != nil {
		return -1
	}

	var m models.Job
	for _, job := range jobs {
		if job.Name == "my-job"{
			m = job
			imageIDs = job.ImageIDs
		}
	}

	return m.ID
}

func findTestJobVersionID() int {
	payload := []byte(``)
	req, _ := http.NewRequest("GET", "/job-versions", bytes.NewBuffer(payload))
	response := executeRequest(req)

	if response.Code != 200 {
		return -1
	}

	jvs := []models.JobVersion{}
	err :=json.Unmarshal(response.Body.Bytes(), &jvs)

	if err != nil {
		return -1
	}

	var m models.JobVersion
	for _, jv := range jvs {
		if jv.JobVersion == "0.1.0" && jv.JobName == "my-job"{
			m = jv
		}
	}

	return m.ID
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