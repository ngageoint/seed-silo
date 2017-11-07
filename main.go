package main

import (
	"log"
	"net/http"
	"database/sql"

	"github.com/gorilla/mux"
	"gopkg.in/natefinch/lumberjack.v2"
)

var db *sql.DB
var router *mux.Router
var err error

func init() {
	db = InitDB("./seed-silo.db")
	router, err = NewRouter()
}

func main() {
	defer db.Close()

	if err != nil {
		log.Fatalf("Error initializing router: %v\n", err.Error())
		log.Fatalln("quitting")
		return
	}



	log.SetOutput(&lumberjack.Logger{
		Filename:   "/var/log/silo/silo.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Compress:   false, // disabled by default
	})

	log.Fatal(http.ListenAndServe(":8080", router))
}

func GetDb() *sql.DB {
	return db
}

func GetRouter() *mux.Router {
	return router
}