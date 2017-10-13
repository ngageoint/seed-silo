package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/JohnPTobe/seed-discover/models"
	"github.com/gorilla/mux"
	"github.com/ngageoint/seed-cli/registry"
	"github.com/ngageoint/seed-cli/objects"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome!\n")
	var myMap = make(map[string]interface{})
	rows, err := db.Query("SELECT * FROM RegistryInfo")
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	colNames, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}
	for _, col := range colNames {
		fmt.Fprint(w, col, "\n")
	}
	cols := make([]interface{}, len(colNames))
	colPtrs := make([]interface{}, len(colNames))
	for i := 0; i < len(colNames); i++ {
		colPtrs[i] = &cols[i]
	}
	for rows.Next() {
		err = rows.Scan(colPtrs...)
		if err != nil {
			log.Fatal(err)
		}
		for i, col := range cols {
			myMap[colNames[i]] = col
		}
		// Do something with the map
		for key, val := range myMap {
			fmt.Fprint(w, "Key: ", key, " Value: ", val, "\n")
		}
	}
	rows.Close()

	rows, err = db.Query("SELECT * FROM Image")
	for rows.Next() {
		img := models.Image{}
		rows.Scan(&img.ID, &img.Name, &img.Registry, &img.Org, &img.Manifest)
		fmt.Fprintln(w, img)
	}
	rows.Close()
	fmt.Fprint(w, "Done!\n")
}

func AddRegistry(w http.ResponseWriter, r *http.Request) {
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
	username := reginfo.Username
	password := reginfo.Password

	registry, err := registry.CreateRegistry(url, username, password)
	if registry == nil || err != nil {
		humanError := checkError(err, url, username, password)
		respondWithError(w, http.StatusBadRequest, humanError)
		log.Print(humanError)
	} else {
		reginfolist := []models.RegistryInfo{}
		reginfolist = append(reginfolist, reginfo)
		err = models.StoreRegistry(db, reginfolist)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
	}
}

func DeleteRegistry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Registry ID")
		return
	}

	if err := models.DeleteRegistry(db, id); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func ScanRegistry(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM RegistryInfo")
	if err != nil {
		log.Fatal(err)
	}

	//clear out image table before scanning
	_, err = db.Exec("DELETE FROM Image")
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Scanning registries...")

	dbImages := []models.Image{}

	for rows.Next() {
		item := models.RegistryInfo{}
		err2 := rows.Scan(&item.ID, &item.Name, &item.Url, &item.Org, &item.Username, &item.Password)
		if err2 != nil {
			panic(err2)
		}
		fmt.Fprintf(w, "Scanning registry %s... \n url: %s \n org: %s \n", item.Name, item.Url, item.Org)
		registry, err := registry.CreateRegistry(item.Url, item.Username, item.Password)
		if err != nil {
			humanError := checkError(err, item.Url, item.Username, item.Password)
			fmt.Fprint(w, humanError, "\n")
		}

		images, err := registry.ImagesWithManifests(item.Org)

		for _, img := range images {
			image := models.Image{Name: img.Name, Registry: img.Registry, Org: img.Org, Manifest: img.Manifest, RegistryId: item.ID}
			dbImages = append(dbImages, image)
			_, err := json.Marshal(img)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	rows.Close()

	models.StoreImage(db, dbImages)
}

func ListImages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	images := models.ReadImages(db)

	json.NewEncoder(w).Encode(images)
}

func ListRegistries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	registries := models.ReadRegistries(db)

	json.NewEncoder(w).Encode(registries)
}

//TODO: Enhance search with multiple keywords, ranking results
func SearchImages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["query"]

	images := models.ReadImages(db)
	results := []models.Image{}
	for _, img := range images {
		if strings.Contains(img.Name, query) {
			results = append(results, img)
			continue
		}
		if strings.Contains(img.Org, query) {
			results = append(results, img)
			continue
		}
		if strings.Contains(img.Manifest, query) {
			results = append(results, img)
			continue
		}
	}

	respondWithJSON(w, http.StatusOK, results)
}

func ImageManifest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Registry ID")
		return
	}

	image, err := models.ReadImage(db, id)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusBadRequest, "No image found with that ID")
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
	}

	seed := &objects.Seed{}

	err = json.Unmarshal([]byte(image.Manifest), &seed)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, seed)
}

func checkError(err error, url, username, password string) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	humanError := ""

	if strings.Contains(errStr, "status=401") {
		if username == "" || password == "" {
			humanError = "The specified registry requires a login.  Please try again with a username (-u) and password (-p)."
		} else {
			humanError = "Incorrect username/password."
		}
	} else if strings.Contains(errStr, "status=404") {
		humanError = "Connected to registry but received a 404 error. Please check the url and try again."
	} else {
		humanError = "Could not connect to the specified registry. Please check the url and try again."
	}
	return humanError
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
