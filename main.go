package main

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/JohnPTobe/seed-common/util"
	"github.com/gorilla/mux"
	"gopkg.in/natefinch/lumberjack.v2"
)

var db *sql.DB
var router *mux.Router
var err error

// ScanLock is safe to use concurrently.
type ScanLock struct {
	ScanInProcess bool
	mux           sync.Mutex
}

// IsScanning checks whether the registries are being scanned
func (sl *ScanLock) IsScanning() bool {
	sl.mux.Lock()
	defer sl.mux.Unlock()
	return sl.ScanInProcess
}

// StartScan
func (sl *ScanLock) StartScan() {
	sl.mux.Lock()
	defer sl.mux.Unlock()
	sl.ScanInProcess = true
}

// EndScan
func (sl *ScanLock) EndScan() {
	sl.mux.Lock()
	defer sl.mux.Unlock()
	sl.ScanInProcess = false
}

var sl = ScanLock{ScanInProcess: false}

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

	logfile := &lumberjack.Logger{
		Filename:   "/usr/silo/silo.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,    //days
		Compress:   false, // disabled by default
	}

	mw := io.MultiWriter(os.Stdout, logfile)
	log.SetOutput(mw)

	log.Fatal(http.ListenAndServe(":9000", router))
}

func GetDb() *sql.DB {
	return db
}

func GetRouter() *mux.Router {
	return router
}
