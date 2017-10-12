package models

import (
	"database/sql"
	"log"
)

type Image struct {
	ID         int    `db:id`
	RegistryId int    `db:registry_id`
	Name       string `db:name`
	Registry   string `db:registry`
	Org        string `db:org`
	Manifest   string `db:manifest`
}

func CreateImageTable(db *sql.DB) {
	// create table if not exists
	sql_table := `
	CREATE TABLE IF NOT EXISTS Image(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		registry_id INTEGER NOT NULL,
		name TEXT,
		registry TEXT,
		org TEXT,
		manifest TEXT,
		CONSTRAINT fk_inv_registry_id
		    FOREIGN KEY (registry_id)
		    REFERENCES RegistryInfo (id)
		    ON DELETE CASCADE
	);
	`

	_, err := db.Exec(sql_table)
	if err != nil {
		panic(err)
	}
}

func StoreImage(db *sql.DB, images []Image) {
	sql_addimg := `
	INSERT OR REPLACE INTO Image(
	    registry_id,
		name,
		registry,
		org,
		manifest
	) values(?, ?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sql_addimg)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	for _, img := range images {
		_, err2 := stmt.Exec(img.RegistryId, img.Name, img.Registry, img.Org, img.Manifest)
		if err2 != nil {
			panic(err2)
		}
	}
}

func ReadImages(db *sql.DB) []Image {
	sql_readall := `
	SELECT * FROM Image
	ORDER BY id ASC
	`

	rows, err := db.Query(sql_readall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []Image
	for rows.Next() {
		item := Image{}
		err2 := rows.Scan(&item.ID, &item.RegistryId, &item.Name, &item.Registry, &item.Org, &item.Manifest)
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
