package main

import (
	"log"
	"net/http"
	"database/sql"

	"github.com/JohnPTobe/silo/models"
	"github.com/gorilla/mux"
)

var db *sql.DB
var router *mux.Router

func init() {
	db = InitDB("./seed-silo.db")
	router = NewRouter()
}

func main() {
	defer db.Close()
	models.CreateImageTable(db)
	models.CreateRegistryTable(db)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func GetDb() *sql.DB {
	return db
}

func GetRouter() *mux.Router {
	return router
}