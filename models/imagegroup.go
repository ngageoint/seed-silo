package models

import (
	"database/sql"
	"log"
)

type ImageGroup struct {
	ID                   int    `db:"id"`
	Name                 string `db:"name"`
	LatestJobVersion     string `db:"latest_job_version"`
	LatestPackageVersion string `db:"latest_package_version"`
	LatestVersionId      int    `db:"latest_image_version_id"`
	Title                string `db:"title"`
	Maintainer           string `db:"maintainer"`
	Email                string `db:"email"`
	Description          string `db:"description"`
}

func CreateImageGroupTable(db *sql.DB) {
	// create table if not exists
	sql_table := `
	CREATE TABLE IF NOT EXISTS ImageGroup(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		latest_image_version_id INTEGER NOT NULL,
		CONSTRAINT fk_inv_latest_image_id
		    FOREIGN KEY (latest_image_version_id)
		    REFERENCES Images (id)
		    ON DELETE CASCADE
	);
	`

	_, err := db.Exec(sql_table)
	if err != nil {
		panic(err)
	}
}

func ResetImageGroupTable(db *sql.DB) error {
	// delete all images and reset the counter
	delete := `DELETE FROM ImageGroup;`

	_, err := db.Exec(delete)
	if err != nil {
		panic(err)
	}

	reset := `DELETE FROM sqlite_sequence WHERE NAME='ImageGroup';`

	_, err2 := db.Exec(reset)
	if err2 != nil {
		panic(err2)
	}

	return err2
}

func StoreImageGroup(db *sql.DB, imagegroups []ImageGroup) {
	sql_add := `
	INSERT OR REPLACE INTO ImageGroup(
		name,
		latest_image_version_id
	) values(?, ?)
	`

	stmt, err := db.Prepare(sql_add)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	for _, ig := range imagegroups {
		_, err2 := stmt.Exec(ig.Name, ig.LatestVersionId)
		if err2 != nil {
			panic(err2)
		}
	}
}

func ReadImageGroups(db *sql.DB) []ImageGroup {
	sql_readall := `
	SELECT * FROM ImageGroup
	ORDER BY id ASC
	`

	rows, err := db.Query(sql_readall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []ImageGroup
	for rows.Next() {
		item := ImageGroup{}
		err2 := rows.Scan(&item.ID, &item.Name, &item.LatestVersionId)
		if err2 != nil {
			panic(err2)
		}

		img, err2 := ReadImage(db, item.LatestVersionId)

		item.LatestJobVersion = img.JobVersion
		item.LatestPackageVersion = img.PackageVersion
		item.Title = img.Seed.Job.Title
		item.Description = img.Seed.Job.Description
		item.Maintainer = img.Seed.Job.Maintainer.Name
		item.Email = img.Seed.Job.Maintainer.Email

		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return result
}

func ReadImageGroup(db *sql.DB, id int) (ImageGroup, error) {
	row := db.QueryRow("SELECT * FROM ImageGroup WHERE id=?", id)

	var result ImageGroup
	err := row.Scan(&result.ID, &result.Name, &result.LatestVersionId)
	if err != nil {
		panic(err)
	}

	img, err := ReadImage(db, result.LatestVersionId)

	result.LatestJobVersion = img.JobVersion
	result.LatestPackageVersion = img.PackageVersion
	result.Title = img.Seed.Job.Title
	result.Description = img.Seed.Job.Description
	result.Maintainer = img.Seed.Job.Maintainer.Name
	result.Email = img.Seed.Job.Maintainer.Email

	return result, err
}

type JobVersion struct {
	ID                   int    `db:"id"`
	GroupName                 string `db:"group_name"`
	JobVersion string    `db:"job_version"`  //major job version
	ImageGroupId int `db:"image_group_id"`
	LatestPackageId      int    `db:"latest_package_id"`
	LatestPackageVersion string `db:"latest_package_version"`
}

func CreateJobVersionTable(db *sql.DB) {
	// create table if it does not exist
	sql_table := `
	CREATE TABLE IF NOT EXISTS JobVersion(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_name TEXT,
		job_version TEXT,
		image_group_id INTEGER NOT NULL,
		latest_package_id INTEGER NOT NULL,
		CONSTRAINT fk_inv_latest_package_id
		    FOREIGN KEY (latest_package_id)
		    REFERENCES Images (id)
		    ON DELETE CASCADE
	);
	`

	_, err := db.Exec(sql_table)
	if err != nil {
		panic(err)
	}
}

func ResetJobVersionTable(db *sql.DB) error {
	// delete all images and reset the counter
	delete := `DELETE FROM JobVersion;`

	_, err := db.Exec(delete)
	if err != nil {
		panic(err)
	}

	reset := `DELETE FROM sqlite_sequence WHERE NAME='JobVersion';`

	_, err2 := db.Exec(reset)
	if err2 != nil {
		panic(err2)
	}

	return err2
}

func StoreJobVersion(db *sql.DB, jobs []JobVersion) {
	sql_add := `
	INSERT OR REPLACE INTO JobVersion(
		group_name,
		job_version,
		image_group_id,
		latest_package_id
	) values(?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sql_add)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	for _, j := range jobs {
		_, err2 := stmt.Exec(j.GroupName, j.JobVersion, j.ImageGroupId, j.LatestPackageId)
		if err2 != nil {
			panic(err2)
		}
	}
}

func ReadJobVersions(db *sql.DB) []JobVersion {
	sql_readall := `
	SELECT * FROM JobVersion
	ORDER BY id ASC
	`

	rows, err := db.Query(sql_readall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []JobVersion
	for rows.Next() {
		item := JobVersion{}
		err2 := rows.Scan(&item.ID, &item.GroupName, &item.JobVersion, &item.ImageGroupId, &item.LatestPackageId)
		if err2 != nil {
			panic(err2)
		}

		img, err2 := ReadImage(db, item.LatestPackageId)

		item.LatestPackageVersion = img.PackageVersion

		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return result
}

func ReadJobVersion(db *sql.DB, id int) (JobVersion, error) {
	row := db.QueryRow("SELECT * FROM JobVersion WHERE id=?", id)

	var result JobVersion
	err := row.Scan(&result.ID, &result.GroupName, &result.JobVersion, &result.ImageGroupId, &result.LatestPackageId)
	if err != nil {
		panic(err)
	}

	img, err := ReadImage(db, result.LatestPackageId)

	result.LatestPackageVersion = img.PackageVersion

	return result, err
}