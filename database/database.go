package database

import (
	"database/sql"

	"github.com/ngageoint/seed-silo/models"
	_ "github.com/mattn/go-sqlite3"
    "github.com/lib/pq"
)

var data *sql.DB

var dbType string

func InitSqliteDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil { panic(err) }
	if db == nil { panic("db nil") }
	if err := db.Ping(); err != nil { panic(err) }

	data = db
	dbType = "sqlite"

	models.CreateImageTable(db, dbType)
	models.CreateRegistryTable(db, dbType)
	models.CreateUser(db, dbType)
	models.CreateJobTable(db, dbType)
	models.CreateJobVersionTable(db, dbType)

	return db
}

func InitPostgresDB(url string) *sql.DB {
    connection, _ := pq.ParseURL(url)
    db, err := sql.Open("postgres", connection)
    if err != nil { panic(err) }
    if db == nil { panic("db nil") }
    if err := db.Ping(); err != nil { panic(err) }

	data = db
	dbType = "postgres"

	models.CreateRegistryTable(db, dbType)
	models.CreateJobTable(db, dbType)
	models.CreateJobVersionTable(db, dbType)
	models.CreateImageTable(db, dbType)
	models.CreateUser(db, dbType)

	return db
}

func GetDB() *sql.DB {
	return data
}

func GetDbType() string {
	return dbType
}