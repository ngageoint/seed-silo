package models

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/ngageoint/seed-common/objects"
)

type Image struct {
	ID             int    `db:"id"`
	RegistryId     int    `db:"registry_id"`
	JobId          int    `db:"job_id"`
	JobVersionId   int    `db:"job_version_id"`
	FullName       string `db:"full_name"`  //full name from registry (may include org et. al.)
	ShortName      string `db:"short_name"` //job name from seed manifest
	JobVersion     string `db:"job_version"`
	PackageVersion string `db:"package_version"`
	Registry       string `db:"registry"`
	Org            string `db:"org"`
	Manifest       string `db:"manifest"`
	Seed           objects.Seed
}

type SimpleImage struct {
	ID             int
	RegistryId     int
	Name           string
	Registry       string
	Org            string
	JobName        string
	Title          string
	JobVersion     string
	PackageVersion string
	Description    string
}

func SimplifyImage(img Image) (SimpleImage, error) {
	simple := SimpleImage{}
	simple.ID = img.ID
	simple.RegistryId = img.RegistryId
	simple.Name = img.ShortName
	simple.Registry = img.Registry
	simple.Org = img.Org

	var seed objects.Seed
	err := json.Unmarshal([]byte(img.Manifest), &seed)
	if err != nil {
		log.Printf("Error unmarshalling seed manifest for %s: %s \n", img.ShortName, err.Error())
	}

	simple.JobName = seed.Job.Name
	simple.Title = seed.Job.Title
	simple.JobVersion = seed.Job.JobVersion
	simple.PackageVersion = seed.Job.PackageVersion
	simple.Description = seed.Job.Description

	return simple, err
}

func CreateImageTable(db *sql.DB) {
	// create table if it does not exist
	sql_table := `
	CREATE TABLE IF NOT EXISTS Image(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		registry_id INTEGER NOT NULL,
		job_id INTEGER,
		job_version_id INTEGER,
		full_name TEXT,
		short_name TEXT,
		job_version TEXT,
		package_version TEXT,
		registry TEXT,
		org TEXT,
		manifest TEXT,
		CONSTRAINT fk_inv_registry_id
		    FOREIGN KEY (registry_id)
		    REFERENCES RegistryInfo (id)
		    ON DELETE CASCADE,
		CONSTRAINT fk_inv_job_id
		    FOREIGN KEY (job_id)
		    REFERENCES Job (id)
		    ON DELETE SET NULL,
		CONSTRAINT fk_inv_job_version_id
		    FOREIGN KEY (job_version_id)
		    REFERENCES JobVersion (id)
		    ON DELETE SET NULL
	);
	`

	_, err := db.Exec(sql_table)
	if err != nil {
		panic(err)
	}
}

func ResetImageTable(db *sql.DB) error {
	// delete all images and reset the counter
	delete := `DELETE FROM Image;`

	_, err := db.Exec(delete)
	if err != nil {
		panic(err)
	}

	reset := `DELETE FROM sqlite_sequence WHERE NAME='Image';`

	_, err2 := db.Exec(reset)
	if err2 != nil {
		panic(err2)
	}

	return err2
}

func StoreImage(db *sql.DB, images []Image) {
	sql_addimg := `
	INSERT OR REPLACE INTO Image(
	    registry_id,
	    job_id,
	    job_version_id,
	    full_name,
		short_name,
		job_version,
		package_version,
		registry,
		org,
		manifest
	) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sql_addimg)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	for _, img := range images {
		_, err2 := stmt.Exec(img.RegistryId, img.JobId, img.JobVersionId, img.FullName,
			img.ShortName, img.JobVersion, img.PackageVersion, img.Registry, img.Org, img.Manifest)
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
		err2 := rows.Scan(&item.ID, &item.RegistryId, &item.JobId, &item.JobVersionId, &item.FullName,
			&item.ShortName, &item.JobVersion, &item.PackageVersion, &item.Registry, &item.Org, &item.Manifest)
		if err2 != nil {
			panic(err2)
		}

		err2 = json.Unmarshal([]byte(item.Manifest), &item.Seed)
		if err2 != nil {
			log.Printf("Error unmarshalling seed manifest for %s: %s \n", item.FullName, err2.Error())
		}

		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return result
}

func ReadSimpleImages(db *sql.DB) []SimpleImage {
	sql_readall := `
	SELECT * FROM Image
	ORDER BY id ASC
	`

	rows, err := db.Query(sql_readall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []SimpleImage
	for rows.Next() {
		item := SimpleImage{}
		var manifest string
		err2 := rows.Scan(&item.ID, &item.RegistryId, &item.Name, &item.Registry, &item.Org, &manifest)
		if err2 != nil {
			panic(err2)
		}

		var seed objects.Seed
		err2 = json.Unmarshal([]byte(manifest), &seed)
		if err2 != nil {
			log.Printf("Error unmarshalling seed manifest for %s: %s \n", item.Name, err2.Error())
		}

		item.JobName = seed.Job.Name
		item.Title = seed.Job.Title
		item.JobVersion = seed.Job.JobVersion
		item.PackageVersion = seed.Job.PackageVersion
		item.Description = seed.Job.Description

		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return result
}

func ReadImage(db *sql.DB, id int) (Image, error) {
	row := db.QueryRow("SELECT * FROM Image WHERE id=?", id)

	var result Image
	err := row.Scan(&result.ID, &result.RegistryId, &result.JobId, &result.JobVersionId, &result.FullName,
		&result.ShortName, &result.JobVersion, &result.PackageVersion, &result.Registry, &result.Org, &result.Manifest)

	if err == nil {
		err = json.Unmarshal([]byte(result.Manifest), &result.Seed)
		if err != nil {
			log.Printf("Error unmarshalling seed manifest for %s: %s \n", result.FullName, err.Error())
			err = nil
		}
	}

	return result, err
}

func DeleteRegistryImages(db *sql.DB, registryId int) error {
	_, err := db.Exec("DELETE FROM Image WHERE registry_id=$1", registryId)

	return err
}

func ImageExists(db *sql.DB, im Image) bool {
	row := db.QueryRow("SELECT 'id' FROM Image WHERE name=$1 AND registry_id=$2", im.FullName, im.RegistryId)

	var result Image
	err := row.Scan(&result.ID)

	return err == nil
}
