package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/JohnPTobe/seed-common/util"
	"github.com/gorilla/mux"
	"gopkg.in/natefinch/lumberjack.v2"
)

var db *sql.DB
var router *mux.Router
var err error

func main() {
	db = InitDB("/usr/silo/seed-silo.db")
	router, err = NewRouter()
	util.InitPrinter(util.PrintLog)
	defer db.Close()

	if err != nil {
		log.Fatalf("Error initializing router: %v\n", err.Error())
		log.Fatalln("quitting")
		return
	}

	log.SetOutput(&lumberjack.Logger{
		Filename:   "/usr/silo/silo.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,    //days
		Compress:   false, // disabled by default
	})

	log.Fatal(http.ListenAndServe(":9000", router))
}

func GetDb() *sql.DB {
	return db
}

func GetRouter() *mux.Router {
	return router
}
