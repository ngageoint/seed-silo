package handlers_users_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/ngageoint/seed-common/util"
	"github.com/ngageoint/seed-silo/database"
	"github.com/ngageoint/seed-silo/route"
)

var token = ""
var db *sql.DB
var router *mux.Router

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

	// Run same tests with Postgres
	url := getEnv("DATABASE_URL", "postgres://scale:scale@localhost:55432/test_silo?sslmode=disable")
	base := strings.Replace(url, "test_silo", "", 1)
	full := strings.Replace(url, "test_silo", "test_silo", 1)
	fmt.Println(url)
	fmt.Println(base)
	fmt.Println(full)
	database.CreatePostgresDB(base, "test_silo")
	db = database.InitPostgresDB(full, "admin", "spicy-pickles17!")

	token, err = login("admin", "spicy-pickles17!")
	if err != nil {
		database.RemovePostgresDB(base, "test_silo")
		os.Exit(-1)
	}

	code += m.Run()

	os.Exit(code)
}

func TestAddUser(t *testing.T) {
	clearTablePG()
	clearTable()

	response := addUser(t)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["ID"] != 1.0 {
		t.Errorf("Expected user ID to be '1'. Got '%v'", m["ID"])
	}

	if m["username"] != "test" {
		t.Errorf("Expected username to be 'test'. Got '%v'", m["username"])
	}

	if m["password"] != "hunter17" {
		t.Errorf("Expected password to be 'hunter17'. Got '%v'", m["password"])
	}

	if m["role"] != "admin" {
		t.Errorf("Expected role to be 'admin'. Got '%v'", m["role"])
	}
}

func TestDeleteUser(t *testing.T) {
	clearTablePG()
	clearTable()

	addUser(t)

	payload := []byte(``)
	req, _ := http.NewRequest("DELETE", "/users/delete/1", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Token: "+token)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/users/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	errorStr := "No user found with that ID"
	if m["error"] != errorStr {
		t.Errorf("Expected error to be '%s'. Got '%v'", errorStr, m["error"])
	}

	req, _ = http.NewRequest("DELETE", "/users/delete/test", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Token: "+token)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestLogin(t *testing.T) {
	clearTablePG()
	clearTable()

	addUser(t)

	cases := []struct {
		username   string
		password   string
		success    bool
		errorMsg string
	}{
		{"test", "hunter17", true, ""},
		{"idiot", "12345", false, "Login error"},
	}

	for _, c := range cases {
		token, err := login(c.username, c.password)

		success := len(token) > 0

		if err == nil && c.success != success {
			t.Errorf("Login for user: %v password: %v failed unexpectedly\n", c.username, c.password)
		}
		if err != nil && !strings.Contains(err.Error(), c.errorMsg) {
			t.Errorf("Login returned an error: %v\n expected %v", err, c.errorMsg)
		}
		if err == nil && c.errorMsg != "" {
			t.Errorf("Login did not return an error when one was expected: %v", c.errorMsg)
		}
	}


}

func clearTable() {
	db.Exec("DELETE FROM RegistryInfo")
	db.Exec("DELETE FROM Image")
	db.Exec("DELETE FROM SiloUser")
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

func addUser(t *testing.T) *httptest.ResponseRecorder {
	payload := []byte(`{"username":"test", "password": "hunter17", "role": "admin"}`)

	req, _ := http.NewRequest("POST", "/users/add", bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Token: "+token)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	return response
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
