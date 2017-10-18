package main

import (
	"log"
	"net/http"
	"database/sql"

	"github.com/JohnPTobe/silo/models"
)

var db *sql.DB

func init() {
	db = InitDB("./seed-silo.db")
}

func main() {
	defer db.Close()
	models.CreateImageTable(db)
	models.CreateRegistryTable(db)
	router := NewRouter()

	log.Fatal(http.ListenAndServe(":8080", router))
}