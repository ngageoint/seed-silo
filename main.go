package main

import (
	"log"
	"net/http"
	"database/sql"
)

var db sql.DB

func init() {
	db := InitDB("./seed-discovery.db")
	_ = db
}

func main() {

	router := NewRouter()

	log.Fatal(http.ListenAndServe(":8080", router))
}