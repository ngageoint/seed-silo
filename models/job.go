package models

import (
	"database/sql"
	"log"
	"github.com/ngageoint/seed-common/util"
)

type Job struct {
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

func SetJobInfo(job *Job, img Image) {
	job.Name = img.ShortName
	job.LatestJobVersion = img.JobVersion
	job.LatestPackageVersion = img.PackageVersion
	job.LatestVersionId = img.ID
	job.Title = img.Seed.Job.Title
	job.Maintainer = img.Seed.Job.Maintainer.Name
	job.Email = img.Seed.Job.Maintainer.Email
	job.Description = img.Seed.Job.Description
}

func CreateJobTable(db *sql.DB) {
	// create table if it does not exist
	sql_table := `
	CREATE TABLE IF NOT EXISTS Job(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		latest_job_version TEXT,
		latest_package_version TEXT,
		latest_image_version_id INTEGER NOT NULL,
		title TEXT,
		maintainer TEXT,
		email TEXT,
		description TEXT,
		CONSTRAINT fk_inv_latest_image_id
		    FOREIGN KEY (latest_image_version_id)
		    REFERENCES Image (id)
		    ON DELETE CASCADE
	);
	`

	_, err := db.Exec(sql_table)
	if err != nil {
		panic(err)
	}
}

func ResetJobTable(db *sql.DB) error {
	// delete all jobs and reset the counter
	delete := `DELETE FROM Job;`

	_, err := db.Exec(delete)
	if err != nil {
		panic(err)
	}

	reset := `DELETE FROM sqlite_sequence WHERE NAME='Job';`

	_, err2 := db.Exec(reset)
	if err2 != nil {
		panic(err2)
	}

	return err2
}

func BuildJobsList(images []Image) []Job {
	jobs := []Job{}
	jobMap := make(map[string]Job)
	for _, img := range images {
		img.ShortName = img.Seed.Job.Name
		img.JobVersion = img.Seed.Job.JobVersion
		img.PackageVersion = img.Seed.Job.PackageVersion
			job, ok := jobMap[img.ShortName]
			if ok {
				jv := img.JobVersion
				pv := img.PackageVersion
				lj := job.LatestJobVersion
				lp := job.LatestPackageVersion
				if jv > lj || (jv == lj && pv > lp) {
					SetJobInfo(&job, img)
				}
			}
			if !ok {
				job = Job{}
				SetJobInfo(&job, img)
				jobMap[img.ShortName] = job
				jobs = append(jobs, job)
			}
	}

	return jobs
}

func StoreJob(db *sql.DB, jobs []Job) {
	sql_add := `
	INSERT OR REPLACE INTO Job(
		name,
		latest_image_version_id
	) values(?, ?)
	`

	stmt, err := db.Prepare(sql_add)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	for _, job := range jobs {
		_, err2 := stmt.Exec(job.Name, job.LatestVersionId)
		if err2 != nil {
			panic(err2)
		}
	}
}

func ReadJobs(db *sql.DB) []Job {
	sql_readall := `
	SELECT * FROM Job
	ORDER BY id ASC
	`

	rows, err := db.Query(sql_readall)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []Job
	for rows.Next() {
		item := Job{}
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

func ReadJob(db *sql.DB, id int) (Job, error) {
	row := db.QueryRow("SELECT * FROM Job WHERE id=?", id)

	var result Job
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
	JobId int `db:"image_group_id"`
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
		    REFERENCES Image (id)
		    ON DELETE CASCADE
	);
	`

	_, err := db.Exec(sql_table)
	if err != nil {
		panic(err)
	}
}

func ResetJobVersionTable(db *sql.DB) error {
	// delete all job versions and reset the counter
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
		_, err2 := stmt.Exec(j.GroupName, j.JobVersion, j.JobId, j.LatestPackageId)
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
		err2 := rows.Scan(&item.ID, &item.GroupName, &item.JobVersion, &item.JobId, &item.LatestPackageId)
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
	err := row.Scan(&result.ID, &result.GroupName, &result.JobVersion, &result.JobId, &result.LatestPackageId)
	if err != nil {
		panic(err)
	}

	img, err := ReadImage(db, result.LatestPackageId)

	result.LatestPackageVersion = img.PackageVersion

	return result, err
}