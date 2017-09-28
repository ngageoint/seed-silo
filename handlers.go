package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome!\n")
}

func AddRegistry(w http.ResponseWriter, r *http.Request) {

}

func DeleteRegistry(w http.ResponseWriter, r *http.Request) {

}

func ScanRegistry(w http.ResponseWriter, r *http.Request) {

}