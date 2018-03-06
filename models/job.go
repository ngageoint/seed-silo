package models

import (
	"database/sql"

	"github.com/ngageoint/seed-common/util"
	"strings"
)

type Job struct {
	ID                   int    `db:"id"`
	Name                 string `db:"name"`
	LatestJobVersion     string `db:"latest_job_version"`
	LatestPackageVersion string `db:"latest_package_version"`
	Title                string `db:"title"`
	Maintainer           string `db:"maintainer"`
	Email                string `db:"email"`
	Description          string `db:"description"`
}

func SetJobInfo(job *Job, img Image) {
	job.Name = img.ShortName
	job.LatestJobVersion = img.JobVersion
	job.LatestPackageVersion = img.PackageVersion
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
		name TEXT NOT NULL UNIQUE,
		latest_job_version TEXT,
		latest_package_version TEXT,
		title TEXT,
		maintainer TEXT,
		email TEXT,
		description TEXT
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

func BuildJobsList(db *sql.DB, images *[]Image) []Job {
	jobs := []Job{}
	jobMap := make(map[string]Job)
	jobVersions := []JobVersion{}
	jvMap := make(map[string]JobVersion)
	(*images)[0].JobId = 1
	for i, _ := range *images {
		img := &(*images)[i]
		img.ShortName = img.Seed.Job.Name
		img.JobVersion = img.Seed.Job.JobVersion
		img.PackageVersion = img.Seed.Job.PackageVersion

		temp := strings.Split(img.JobVersion, ".")
		if len(temp) != 3 {
			util.PrintUtil("ERROR: Invalid version string for image %v: %v", img.FullName, img.JobVersion)
			continue
		}
		major := temp[0]
		versionName := img.ShortName + temp[0]

		job, ok := jobMap[img.ShortName]
		if ok {
			jv := img.JobVersion
			pv := img.PackageVersion
			lj := job.LatestJobVersion
			lp := job.LatestPackageVersion
			if jv > lj || (jv == lj && pv > lp) {
				SetJobInfo(&job, *img)
				UpdateJob(db, job)
			}
		}
		if !ok {
			job = Job{}
			SetJobInfo(&job, *img)

			id, err := AddJob(db, job)
			if err != nil {
				util.PrintUtil("ERROR: Error adding job in BuildJobsList: %v\n", err)
			}

			job.ID = id
			jobMap[img.ShortName] = job
			jobs = append(jobs, job)
		}

		img.JobId = job.ID

		jobVersion, ok := jvMap[versionName]
		if ok {
			jv := img.JobVersion
			pv := img.PackageVersion
			lj := jobVersion.LatestJobVersion
			lp := jobVersion.LatestPackageVersion
			if jv > lj || (jv == lj && pv > lp) {
				SetJobVersionInfo(&jobVersion, *img)
				UpdateJobVersion(db, jobVersion)
			}
		}
		if !ok {
			jobVersion = JobVersion{MajorVersion:major}
			SetJobVersionInfo(&jobVersion, *img)

			id, err := AddJobVersion(db, jobVersion)
			if err != nil {
				util.PrintUtil("ERROR: Error adding job version in BuildJobsList: %v\n", err)
			}

			jobVersion.ID = id
			jvMap[versionName] = jobVersion
			jobVersions = append(jobVersions, jobVersion)
		}

		img.JobVersionId = jobVersion.ID

	}

	return jobs
}

func AddJob(db *sql.DB, job Job) (int, error) {
	sql_add := `
	INSERT INTO Job(
		name,
		latest_job_version,
		latest_package_version,
		title,
		maintainer,
		email,
		description
	) values(?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sql_add)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(job.Name, job.LatestJobVersion, job.LatestPackageVersion,
		job.Title, job.Maintainer, job.Email, job.Description)
	if err != nil {
		return -1, err
	}

	id := -1
	var id64 int64
	if err == nil {
		id64, err = result.LastInsertId()
		id = int(id64)
	}

	return id, err
}

func UpdateJob(db *sql.DB, job Job) error {
	sql_update := `UPDATE Job SET 
		latest_job_version=?, 
		latest_package_version=?,		
		title=?,
		maintainer=?,
		email=?,
		description=?
		where id=?`

	stmt, err := db.Prepare(sql_update)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(job.LatestJobVersion, job.LatestPackageVersion,
		job.Title, job.Maintainer, job.Email, job.Description, job.ID)

	return err
}

func StoreJobs(db *sql.DB, jobs []Job) {
	sql_add := `
	INSERT INTO Job(
		name,
		latest_job_version,
		latest_package_version,
		title,
		maintainer,
		email,
		description
	) values(?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sql_add)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	for _, job := range jobs {
		_, err2 := stmt.Exec(job.Name, job.LatestJobVersion, job.LatestPackageVersion,
			job.Title, job.Maintainer, job.Email, job.Description)
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
		err2 := rows.Scan(&item.ID, &item.Name, &item.LatestJobVersion, &item.LatestPackageVersion,
			&item.Title, &item.Maintainer, &item.Email, &item.Description)
		if err2 != nil {
			panic(err2)
		}

		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		util.PrintUtil("ERROR: Error in ReadJobs: %v", err)
	}
	return result
}

func ReadJob(db *sql.DB, id int) (Job, error) {
	row := db.QueryRow("SELECT * FROM Job WHERE id=?", id)

	var result Job
	err := row.Scan(&result.ID, &result.Name, &result.LatestJobVersion, &result.LatestPackageVersion,
		&result.Title, &result.Maintainer, &result.Email, &result.Description)
	if err != nil {
		panic(err)
	}

	return result, err
}

type JobVersion struct {
	ID                   int    `db:"id"`
	JobName              string `db:"job_name"`
	MajorVersion         string `db:"major_version"`
	JobId                int    `db:"job_id"`
	LatestJobVersion     string `db:"latest_job_version"`
	LatestPackageVersion string `db:"latest_package_version"`
}

func SetJobVersionInfo(jv *JobVersion, img Image) {
	jv.JobName = img.ShortName
	jv.MajorVersion = img.JobVersion
	jv.JobId = img.JobId
	jv.LatestJobVersion = img.JobVersion
	jv.LatestPackageVersion = img.PackageVersion
}

func CreateJobVersionTable(db *sql.DB) {
	// create table if it does not exist
	sql_table := `
	CREATE TABLE IF NOT EXISTS JobVersion(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_name TEXT,
		major_version TEXT,
		job_id INTEGER NOT NULL,
		latest_job_version TEXT,
		latest_package_version TEXT,
		UNIQUE(job_name, major_version),
		CONSTRAINT fk_inv_job_id
		    FOREIGN KEY (job_id)
		    REFERENCES Job (id)
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

func AddJobVersion(db *sql.DB, jv JobVersion) (int, error) {
	sql_add := `
	INSERT INTO JobVersion(
		job_name,
		major_version,
		job_id,
		latest_job_version,
		latest_package_version
	) values(?, ?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sql_add)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(jv.JobName, jv.MajorVersion, jv.JobId, jv.LatestJobVersion, jv.LatestPackageVersion)
	if err != nil {
		return -1, err
	}

	id := -1
	var id64 int64
	if err == nil {
		id64, err = result.LastInsertId()
		id = int(id64)
	}

	return id, err
}

func UpdateJobVersion(db *sql.DB, jv JobVersion) error {
	sql_update := `UPDATE JobVersion SET 
		job_name=?,
		major_version=?,
		job_id=?,
		latest_job_version=?,
		latest_package_version=?
		where id=?`

	stmt, err := db.Prepare(sql_update)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(jv.JobName, jv.MajorVersion, jv.JobId, jv.LatestJobVersion, jv.LatestPackageVersion)

	return err
}

func ReadJobVersions(db *sql.DB, jobId int) []JobVersion {
	sql_read := `
	SELECT * FROM JobVersion WHERE job_id =?
	ORDER BY id ASC
	`

	rows, err := db.Query(sql_read, jobId)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var result []JobVersion
	for rows.Next() {
		item := JobVersion{}
		err2 := rows.Scan(&item.ID, &item.JobName, &item.MajorVersion, &item.JobId, &item.LatestJobVersion, &item.LatestPackageVersion)
		if err2 != nil {
			panic(err2)
		}

		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		util.PrintUtil("ERROR: Error in ReadJobVersions: %v", err)
	}
	return result
}

func ReadJobVersion(db *sql.DB, id int) (JobVersion, error) {
	row := db.QueryRow("SELECT * FROM JobVersion WHERE id=?", id)

	var result JobVersion
	err := row.Scan(&result.ID, &result.JobName, &result.MajorVersion, &result.JobId, &result.LatestJobVersion, &result.LatestPackageVersion)
	if err != nil {
		return result, err
	}

	return result, err
}
