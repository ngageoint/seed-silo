package models

import (
	"database/sql"
	"log"
	"strings"
)

//TODO: find better way to store credentials for low side registries
type RegistryInfo struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	Url      string `db:"url"`
	Org      string `db:"org"`
	Username string `db:"username"`
	Password string `db:"password"`
}

type DisplayRegistry struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	Url      string `db:"url"`
	Org      string `db:"org"`
}

func CreateRegistryTable(db *sql.DB, type string) {
	// create table if it does not exist
	sql_table := `
	CREATE TABLE IF NOT EXISTS RegistryInfo(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		url TEXT,
		org TEXT,
		username TEXT,
		password TEXT
	);
	`

	if type == "postgres" {
	    strings.replace(sql_table, "id INTEGER PRIMARY KEY AUTOINCREMENT", "id SERIAL PRIMARY KEY", 1)
	}

	_, err := db.Exec(sql_table)
	if err != nil {
		panic(err)
	}
}

func AddRegistry(db *sql.DB, r RegistryInfo) (int, error) {
	sql_addreg := `
	INSERT INTO RegistryInfo(
		name,
		url,
	    org,
		username,
		password
	) values(?, ?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sql_addreg)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(r.Name, r.Url, r.Org, r.Username, r.Password)

	id := -1
	var id64 int64
	if err == nil {
		id64, err = result.LastInsertId()
		id = int(id64)
	}

	return id, err
}

func DeleteRegistry(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM RegistryInfo WHERE id=$1", id)

	return err
}

//Get list of registries without username/password for display
func DisplayRegistries(db *sql.DB) ([]DisplayRegistry, error) {
	sql_readall := `
	SELECT id, name, url, org FROM RegistryInfo
	ORDER BY id ASC
	`

	rows, err := db.Query(sql_readall)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DisplayRegistry
	for rows.Next() {
		item := DisplayRegistry{}
		err2 := rows.Scan(&item.ID, &item.Name, &item.Url, &item.Org)
		if err2 != nil {
			return nil, err
		}
		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return result, err
}

func GetRegistry(db *sql.DB, id int) (RegistryInfo, error) {
	row := db.QueryRow("SELECT * FROM RegistryInfo WHERE id=?", id)

	var result RegistryInfo
	err := row.Scan(&result.ID, &result.Name, &result.Url, &result.Org, &result.Username, &result.Password)

	return result, err
}

func GetRegistries(db *sql.DB) ([]RegistryInfo, error){
	rows, err := db.Query("SELECT * FROM RegistryInfo")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []RegistryInfo
	for rows.Next() {
		item := RegistryInfo{}
		err2 := rows.Scan(&item.ID, &item.Name, &item.Url, &item.Org, &item.Username, &item.Password)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, item)
	}
	return result, err
}
