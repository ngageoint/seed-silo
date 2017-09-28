package models


import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type RegistryInfo  struct{
	ID int `db:id`
	Name string `db:name`
	Url string `db:url`
	Username string `db:username`
	Password string `db:password`
}


func CreateRegistryTable(db *sql.DB) {
	// create table if not exists
	sql_table := `
	CREATE TABLE IF NOT EXISTS RegistryInfo(
		id INTEGER NOT NULL PRIMARY KEY,
		name TEXT,
		url TEXT,
		username TEXT,
		password TEXT
	);
	`

	_, err := db.Exec(sql_table)
	if err != nil { panic(err) }
}

func StoreRegistry(db *sql.DB, registries []RegistryInfo) {
	sql_addreg := `
	INSERT OR REPLACE INTO RegistryInfo(
		id,
		name,
		url,
		username,
		password
	) values(?, ?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sql_addreg)
	if err != nil { panic(err) }
	defer stmt.Close()

	for _, reg := range registries {
		_, err2 := stmt.Exec(reg.ID, reg.Name, reg.Url, reg.Username, reg.Password)
		if err2 != nil { panic(err2) }
	}
}

func ReadRegistries(db *sql.DB) []RegistryInfo {
	sql_readall := `
	SELECT * FROM RegistryInfo
	ORDER BY id ASC
	`

	rows, err := db.Query(sql_readall)
	if err != nil { panic(err) }
	defer rows.Close()

	var result []RegistryInfo
	for rows.Next() {
		item := RegistryInfo{}
		err2 := rows.Scan(&item.ID, &item.Name, &item.Url)
		if err2 != nil { panic(err2) }
		result = append(result, item)
	}
	return result
}