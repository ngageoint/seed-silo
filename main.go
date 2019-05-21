package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/ngageoint/seed-common/util"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/ngageoint/seed-silo/database"
	"github.com/ngageoint/seed-silo/route"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
    url := os.Getenv("DATABASE_URL")
    admin := getEnv("SILO_ADMIN", "admin")
    password := getEnv( "SILO_ADMIN_PASSWORD", "spicy-pickles17!")
    if url == ""{
    	lite := getEnv("SILO_LITE_PATH", "/usr/silo/seed-silo.db")
        db := database.InitSqliteDB(lite, admin, password)
        defer db.Close()
	} else {
		reset_url := os.Getenv("RESET_URL")
		reset_name := os.Getenv("RESET_NAME")
		if reset_url != "" {
			database.CreatePostgresDB(reset_url, reset_name)
		}
        db := database.InitPostgresDB(url, admin, password)
        defer db.Close()
	}

	router, err := route.NewRouter()
	util.InitPrinter(util.PrintLog)

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
