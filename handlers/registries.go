package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ngageoint/seed-common/objects"
	"github.com/ngageoint/seed-common/registry"
	"github.com/ngageoint/seed-silo/models"
	"github.com/ngageoint/seed-silo/database"
)

// ScanLock is safe to use concurrently.
type ScanLock struct {
	ScanInProcess bool
	mux           sync.Mutex
}

// IsScanning checks whether the registries are being scanned
func (sl *ScanLock) IsScanning() bool {
	sl.mux.Lock()
	defer sl.mux.Unlock()
	return sl.ScanInProcess
}

// StartScan
func (sl *ScanLock) StartScan() {
	sl.mux.Lock()
	defer sl.mux.Unlock()
	sl.ScanInProcess = true
}

// EndScan
func (sl *ScanLock) EndScan() {
	sl.mux.Lock()
	defer sl.mux.Unlock()
	sl.ScanInProcess = false
}

var sl = ScanLock{ScanInProcess: false}

func Registry(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	reg, err := models.GetRegistry(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "No registry found with that ID")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondWithJSON(w, http.StatusOK, reg)
}

func AddRegistry(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	dbType := database.GetDbType()
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	if err := r.Body.Close(); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	var reginfo models.RegistryInfo
	if err := json.Unmarshal(body, &reginfo); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error())
	}
	url := reginfo.Url
	org := reginfo.Org
	username := reginfo.Username
	password := reginfo.Password

	registry, err := registry.CreateRegistry(url, org, username, password)
	if registry == nil || err != nil {
		humanError := checkError(err, url, username, password)
		respondWithError(w, http.StatusBadRequest, humanError)
		log.Print(humanError)
		log.Print(err)
	} else {
		var id int
		var err2 error
		if dbType == "postgres" {
			id, err2 = models.AddRegistryPg(db, reginfo)
		} else {
			id, err2 = models.AddRegistryLite(db, reginfo)
		}
		if err2 != nil {
			errStr := err2.Error()
			if strings.Contains(strings.ToLower(errStr), "unique") {
				respondWithJSON(w, http.StatusBadRequest, "Registry already exists with name " + reginfo.Name)
			} else {
				respondWithError(w, http.StatusInternalServerError, err2.Error())
			}
		}
		reginfo.ID = id
		respondWithJSON(w, http.StatusCreated, reginfo)
	}
}

func DeleteRegistry(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := models.DeleteRegistry(db, id); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func ScanRegistry(w http.ResponseWriter, r *http.Request) {
	//prevent multiple requests to scan registries
	if sl.IsScanning() {
		respondWithJSON(w, http.StatusAccepted, map[string]string{"message": "Scanning Registries"})
		return
	}
	sl.StartScan()
	defer sl.EndScan()

	db := database.GetDB()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	registry, err := models.GetRegistry(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "No registry found with that ID")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	w.Header().Set("Content-Type", "application/text; charset=UTF-8")
	w.WriteHeader(http.StatusAccepted)

	list := []models.RegistryInfo{}
	list = append(list, registry)
	dbImages, err := Scan(w, r, list)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//clear out image table before scanning
	err = models.DeleteRegistryImages(db, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//get existing images from other registries, if any
	oldImages := models.ReadImages(db)
	allImages := oldImages
	allImages = append(allImages, dbImages...)

	//clear out job table
	dbType := database.GetDbType()
	err = models.ResetJobTable(db, dbType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//clear out job version table
	err = models.ResetJobVersionTable(db, dbType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.BuildJobsList(db, &allImages, dbType)
	models.StoreOrUpdateImages(db, allImages, dbType)
}

func ScanRegistries(w http.ResponseWriter, r *http.Request) {
	//prevent multiple requests to scan registries
	if sl.IsScanning() {
		respondWithJSON(w, http.StatusAccepted, map[string]string{"message": "Scanning Registries"})
		return
	}
	sl.StartScan()
	defer sl.EndScan()

	db := database.GetDB()
	registries, err := models.GetRegistries(db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/txt; charset=UTF-8")
	w.WriteHeader(http.StatusAccepted)
	log.Print("Scanning registries...")

	dbImages, err := Scan(w, r, registries)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//clear out image table before scanning
	dbType := database.GetDbType()
	err = models.ResetImageTable(db, dbType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//clear out job table
	err = models.ResetJobTable(db, dbType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//clear out job version table
	err = models.ResetJobVersionTable(db, dbType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.BuildJobsList(db, &dbImages, dbType)
	models.StoreImages(db, dbImages, dbType)
}

func Scan(w http.ResponseWriter, req *http.Request, registries []models.RegistryInfo) ([]models.Image, error) {
	dbImages := []models.Image{}
	var err error
	for _, r := range registries {
		log.Printf("Scanning registry %s... \n url: %s \n org: %s \n", r.Name, r.Url, r.Org)
		registry, err := registry.CreateRegistry(r.Url, r.Org, r.Username, r.Password)
		if err != nil {
			humanError := checkError(err, r.Url, r.Username, r.Password)
			respondWithError(w, http.StatusInternalServerError, humanError)
			return nil, err
		}

		if registry == nil {
			respondWithError(w, http.StatusInternalServerError, "Error creating registry.")
			return nil, errors.New("ERROR: Unknown error creating registry.")
		}

		var images []objects.Image
		images, err = registry.ImagesWithManifests()

		for _, img := range images {
			image := models.Image{FullName: img.Name, Registry: img.Registry, Org: img.Org, Manifest: img.Manifest, RegistryId: r.ID}
			err1 := json.Unmarshal([]byte(img.Manifest), &image.Seed)
			if err1 != nil {
				log.Printf("Error unmarshalling seed manifest for %s: %s \n", img.Name, err1.Error())
			}
			image.ShortName = image.Seed.Job.Name
			image.Title = image.Seed.Job.Title
			image.Maintainer = image.Seed.Job.Maintainer.Name
			image.Email = image.Seed.Job.Maintainer.Email
			image.MaintOrg = image.Seed.Job.Maintainer.Organization
			image.JobVersion = image.Seed.Job.JobVersion
			image.PackageVersion = image.Seed.Job.PackageVersion
			image.Description = image.Seed.Job.Description
			dbImages = append(dbImages, image)
		}
	}

	return dbImages, err
}

func ListRegistries(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	registries, err := models.DisplayRegistries(db)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := []models.DisplayRegistry{}
	list = append(list, registries...)
	respondWithJSON(w, http.StatusOK, list)
}