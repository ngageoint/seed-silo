package models

import (
	"database/sql"
	"log"
)

//TODO: find better way to store credentials for low side registries
type RegistryInfo struct {
	ID       int    `db:id`
	Name     string `db:name`
	Url      string `db:url`
	Org      string `db:org`
	Username string `db:username`
	Password string `db:password`
}

func CreateRegistryTable(db *sql.DB) {
	// create table if not exists
	sql_table := `
	CREATE TABLE IF NOT EXISTS RegistryInfo(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		url TEXT,
		org TEXT,
		username TEXT,
		password TEXT
	);
	`

	_, err := db.Exec(sql_table)
	if err != nil {
		panic(err)
	}
}

func StoreRegistry(db *sql.DB, registries []RegistryInfo) error {
	sql_addreg := `
	INSERT OR REPLACE INTO RegistryInfo(
		name,
		url,
	    org,
		username,
		password
	) values(?, ?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sql_addreg)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, reg := range registries {
		_, err2 := stmt.Exec(reg.Name, reg.Url, reg.Org, reg.Username, reg.Password)
		if err2 != nil {
			return err2
		}
	}

	return nil
}

func DeleteRegistry(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM RegistryInfo WHERE id=$1", id)

	return err
}

func ReadRegistries(db *sql.DB) []RegistryInfo {
	sql_readall := `
	SELECT id, name, url, org FROM RegistryInfo
	ORDER BY id ASC
	`

	rows, err := db.Query(sql_readall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []RegistryInfo
	for rows.Next() {
		item := RegistryInfo{}
		err2 := rows.Scan(&item.ID, &item.Name, &item.Url, &item.Org)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return result
}
