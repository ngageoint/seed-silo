package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ngageoint/seed-silo/database"
	"github.com/ngageoint/seed-silo/models"
	"github.com/ngageoint/seed-common/util"
)

func ListJobs(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	jobList := []models.Job{}
	jobs := models.ReadJobs(db)
	jobList = append(jobList, jobs...)

	respondWithJSON(w, http.StatusOK, jobList)
}

func Job(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	job, err := models.ReadJob(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "No job found with that ID")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	job.ImageIDs = models.GetJobImageIds(db, job.ID)
	job.JobVersions = models.GetJobVersions(db, job.ID)

	respondWithJSON(w, http.StatusOK, job)
}

func JobVersions(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	job, err := models.ReadJob(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "No job found with that ID")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	job.JobVersions = models.GetJobVersions(db, job.ID)

	respondWithJSON(w, http.StatusOK, job.JobVersions)
}

func ListJobVersions(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	jvList := []models.JobVersion{}
	jvs := models.ReadJobVersions(db)
	jvList = append(jvList, jvs...)

	respondWithJSON(w, http.StatusOK, jvList)
}

func JobVersion(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	jv, err := models.ReadJobVersion(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "No job version found with that ID")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondWithJSON(w, http.StatusOK, jv)
}

func SearchJobs(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	vars := mux.Vars(r)
	query := vars["query"]

	if query == "" || query == "+" {
		ListJobs(w, r)
		return
	}

	terms := strings.Split(query, "+")

	images := models.ReadImages(db)

	rankedResults := []RankedResult{}
	for _, img := range images {
		score := 0
		for _, term := range terms {
			if strings.Contains(img.FullName, term) {
				score += 10
			}
			if strings.Contains(img.Org, term) {
				score += 10
			}
			seed := img.Seed

			if strings.Contains(fmt.Sprintf("%v", img.Seed), term) {
				score += 1
			}

			if strings.Contains(seed.Job.Name, term) {
				score += 10
			}
			if strings.Contains(seed.Job.Title, term) {
				score += 5
			}

			if strings.Contains(seed.Job.Description, term) {
				score += 5
			}

			if strings.Contains(fmt.Sprintf("%s", seed.Job.Tags), term) {
				score += 10
			}

			if strings.Contains(fmt.Sprintf("%s", seed.Job.Maintainer), term) {
				score += 5
			}

		}
		if score > 0 {
			rankedResults = append(rankedResults, RankedResult{Score: score, Image: img})
		}
	}

	sort.Sort(ByScore(rankedResults))

	results := []models.Job{}
	jobMap := make(map[int]models.Job)
	for _, res := range rankedResults {
		job, _ := models.ReadJob(db, res.Image.JobId)
		jv, err := models.ReadJobVersion(db, res.Image.JobVersionId)
		job.JobVersions = []models.JobVersion{}
		if err == nil {
			job.JobVersions = append(job.JobVersions, jv)
		} else {
			util.PrintUtil("ERROR: Error getting job version %d, %v\n", res.Image.JobVersionId, err.Error())
		}
		job1, ok := jobMap[res.Image.JobId]
		if !ok {
			results = append(results, job)
			job.ImageIDs = append(job.ImageIDs, res.Image.ID)
			jobMap[res.Image.JobId] = job
		} else {
			job1.ImageIDs = append(job1.ImageIDs, res.Image.ID)
			job1.JobVersions = append(job1.JobVersions, job.JobVersions...)
			jobMap[res.Image.JobId] = job1
		}
	}

	respondWithJSON(w, http.StatusOK, jobMap)
}
