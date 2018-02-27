package models

import "database/sql"

type ImageGroup struct {
	ID                   int    `db:"id"`
	Name                 string `db:"name"`
	JobName              string `db:"latest_image_version_id"`
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
		job_name TEXT,
		org TEXT,
		manifest TEXT,
		CONSTRAINT fk_inv_registry_id
		    FOREIGN KEY (registry_id)
		    REFERENCES RegistryInfo (id)
		    ON DELETE CASCADE
	);
	`

type ImageJobVersion struct {
	ID                   int    `db:"id"`
	Name                 string `db:"name"`
	LatestVersionId      int    `db:"latest_package_version"`
	LatestPackageVersion string `db:"latest_package_version"`
}
