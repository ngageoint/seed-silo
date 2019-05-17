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

	models.CreateImageTable(db)
	models.CreateRegistryTable(db)
	models.CreateUser(db)
	models.CreateJobTable(db)
	models.CreateJobVersionTable(db)

	data = db
	dbType = "sqlite"

	return db
}

func InitPostgresDB(url string) *sql.DB {
    connection, _ := pq.ParseURL(url)
    db, err := sql.Open("postgres", connection)
    if err != nil { panic(err) }
    if db == nil { panic("db nil") }
    if err := db.Ping(); err != nil { panic(err) }

	models.CreateImageTable(db)
	models.CreateRegistryTable(db)
	models.CreateUser(db)
	models.CreateJobTable(db)
	models.CreateJobVersionTable(db)

	data = db
	dbType = "postgres"

	return db
}

func GetDB() *sql.DB {
	return data, dbType
}
