package main

import (
	"log"
	"net/http"
	"database/sql"
	
	"github.com/JohnPTobe/seed-discover/models"
)

var db *sql.DB

func init() {
	db = InitDB("./seed-discovery.db")
}

func main() {
	defer db.Close()
	models.CreateImageTable(db)
	models.CreateRegistryTable(db)
	router := NewRouter()

	log.Fatal(http.ListenAndServe(":8080", router))
}