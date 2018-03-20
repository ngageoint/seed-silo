package models

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/ngageoint/seed-common/objects"
	"github.com/ngageoint/seed-common/util"
)

type Image struct {
	ID             int    `db:"id"`
	RegistryId     int    `db:"registry_id"`
	JobId          int    `db:"job_id"`
	JobVersionId   int    `db:"job_version_id"`
	FullName       string `db:"full_name"`  //full name from registry (may include org et. al.)
	ShortName      string `db:"short_name"` //job name from seed manifest
	Title          string `db:"title"`
	Maintainer     string `db:"maintainer"`
	Email          string `db:"email"`
	MaintOrg       string `db:"maint_org"`
	JobVersion     string `db:"job_version"`
	PackageVersion string `db:"package_version"`
	Description    string `db:"description"`
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
	Maintainer     string
	Email          string
	MaintOrg       string
	Description    string
	JobVersion     string
	PackageVersion string
}

func SimplifyImage(img Image) SimpleImage {
	simple := SimpleImage{}
	simple.ID = img.ID
	simple.RegistryId = img.RegistryId
	simple.Name = img.FullName
	simple.Registry = img.Registry
	simple.Org = img.Org
	simple.JobName = img.ShortName
	simple.Title = img.Title
	simple.Maintainer = img.Maintainer
	simple.Email = img.Email
	simple.MaintOrg = img.MaintOrg
	simple.Description = img.Description
	simple.JobVersion = img.JobVersion
	simple.PackageVersion = img.PackageVersion

	return simple
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
		title TEXT,
		maintainer TEXT,
		email TEXT,
		maint_org TEXT,
		job_version TEXT,
		package_version TEXT,
		description TEXT,
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

func StoreImages(db *sql.DB, images []Image) {
	sql_addimg := `
	INSERT INTO Image(
	    registry_id,
	    job_id,
	    job_version_id,
	    full_name,
		short_name,
		title,
		maintainer,
		email,
		maint_org,
		job_version,
		package_version,
		description,
		registry,
		org,
		manifest
	) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sql_addimg)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	for _, img := range images {
		_, err2 := stmt.Exec(img.RegistryId, img.JobId, img.JobVersionId, img.FullName,
			img.ShortName, img.Title, img.Maintainer, img.Email, img.MaintOrg,
			img.JobVersion, img.PackageVersion, img.Description, img.Registry,
			img.Org, img.Manifest)
		if err2 != nil {
			panic(err2)
		}
	}
}

func StoreOrUpdateImages(db *sql.DB, images []Image) {
	sql_add_img := `
	INSERT INTO Image(
	    registry_id,
	    job_id,
	    job_version_id,
	    full_name,
		short_name,
		title,
		maintainer,
		email,
		maint_org,
		job_version,
		package_version,
		description,
		registry,
		org,
		manifest
	) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	addStatement, err := db.Prepare(sql_add_img)
	if err != nil {
		panic(err)
	}
	defer addStatement.Close()

	sql_update_img := `
	UPDATE Image SET
	    registry_id=?,
	    job_id=?,
	    job_version_id=?,
	    full_name=?,
		short_name=?,
		title=?,
		maintainer=?,
		email=?,
		maint_org=?,
		job_version=?,
		package_version=?,
		description=?,
		registry=?,
		org=?,
		manifest=?
	WHERE id=?
	`

	updateStatement, err := db.Prepare(sql_update_img)
	if err != nil {
		panic(err)
	}
	defer updateStatement.Close()

	for _, img := range images {
		if img.ID != 0 {
			_, err2 := updateStatement.Exec(img.RegistryId, img.JobId, img.JobVersionId, img.FullName,
				img.ShortName, img.Title, img.Maintainer, img.Email, img.MaintOrg, img.JobVersion,
				img.PackageVersion, img.Description, img.Registry, img.Org, img.Manifest, img.ID)
			if err2 != nil {
				panic(err2)
			}
		} else {
			_, err2 := addStatement.Exec(img.RegistryId, img.JobId, img.JobVersionId, img.FullName,
				img.ShortName, img.Title, img.Maintainer, img.Email, img.MaintOrg, img.JobVersion,
				img.PackageVersion, img.Description, img.Registry, img.Org, img.Manifest)
			if err2 != nil {
				panic(err2)
			}
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
			&item.ShortName, &item.Title, &item.Maintainer, &item.Email, &item.MaintOrg, &item.JobVersion,
			&item.PackageVersion, &item.Description, &item.Registry, &item.Org, &item.Manifest)
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
		img := Image{}
		var manifest string
		err2 := rows.Scan(&item.ID, &item.RegistryId, &img.JobId, &img.JobVersionId, &item.Name,
			&item.JobName, &item.Title, &item.Maintainer, &item.Email, &item.MaintOrg,
			&item.JobVersion, &item.PackageVersion, &item.Description,
			&item.Registry, &item.Org, &manifest)
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

func ReadImage(db *sql.DB, id int) (Image, error) {
	row := db.QueryRow("SELECT * FROM Image WHERE id=?", id)

	var result Image
	err := row.Scan(&result.ID, &result.RegistryId, &result.JobId, &result.JobVersionId,
		&result.FullName, &result.ShortName, &result.Title, &result.Maintainer, &result.Email,
		&result.MaintOrg, &result.JobVersion, &result.PackageVersion, &result.Description,
		&result.Registry, &result.Org, &result.Manifest)

	if err != nil {
		util.PrintUtil("ERROR scanning in read image: %v", err.Error())
	}

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

func GetJobImageIds(db *sql.DB, jobid int) []int {
	sql_readall := `SELECT ID FROM Image WHERE job_id=?`

	rows, err := db.Query(sql_readall, jobid)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []int
	for rows.Next() {
		var id int
		err2 := rows.Scan(&id)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, id)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return result
}

func GetJobImages(db *sql.DB, jobid int) []SimpleImage {
	sql_readall := `SELECT * FROM Image WHERE job_id=?`

	rows, err := db.Query(sql_readall, jobid)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []SimpleImage
	for rows.Next() {
		item := SimpleImage{}
		img := Image{}
		var manifest string
		err2 := rows.Scan(&item.ID, &item.RegistryId, &img.JobId, &img.JobVersionId, &item.Name,
			&item.JobName, &item.Title, &item.Maintainer, &item.Email, &item.MaintOrg,
			&item.JobVersion, &item.PackageVersion, &item.Description, &item.Registry, &item.Org, &manifest)
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

func GetJobVersionImageIds(db *sql.DB, jobversionid int) []int {
	sql_readall := `SELECT ID FROM Image WHERE job_version_id=?`

	rows, err := db.Query(sql_readall, jobversionid)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []int
	for rows.Next() {
		var id int
		err2 := rows.Scan(&id)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, id)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return result
}

func GetJobVersionImages(db *sql.DB, jobversionid int) []Image {
	sql_readall := `SELECT * FROM Image WHERE job_version_id=?`

	rows, err := db.Query(sql_readall, jobversionid)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []Image
	for rows.Next() {
		item := Image{}
		err2 := rows.Scan(&item.ID, &item.RegistryId, &item.JobId, &item.JobVersionId, &item.FullName,
			&item.ShortName, &item.Title, &item.Maintainer, &item.Email, &item.MaintOrg,
			&item.JobVersion, &item.PackageVersion, &item.Description, &item.Registry, &item.Org, &item.Manifest)
		if err2 != nil {
			panic(err2)
		}

		err = json.Unmarshal([]byte(item.Manifest), &item.Seed)
		if err != nil {
			log.Printf("Error unmarshalling seed manifest for %s: %s \n", item.FullName, err.Error())
			err = nil
		}

		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return result
}
