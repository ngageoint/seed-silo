package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/JohnPTobe/seed-discover/models"
	"log"
	"github.com/ngageoint/seed-cli/registry"
	"strings"
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
			fmt.Fprint(w, "Key: ", key, "Value: ", val,  "\n")
		}
	}
	fmt.Fprint(w, "Done!\n")
}

func AddRegistry(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	var reginfo models.RegistryInfo
	if err := json.Unmarshal(body, &reginfo); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}
	url := reginfo.Url
	username := reginfo.Username
	password := reginfo.Password

	registry, err := registry.CreateRegistry(url, username, password)
	if registry != nil && err == nil {
		humanError := checkError(err, url, username, password)
		fmt.Fprint(w, humanError, "\n")
	} else {
		reginfolist := []models.RegistryInfo{}
		reginfolist = append(reginfolist, reginfo)
		err = models.StoreRegistry(db, reginfolist)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)
		if err != nil {
			panic(err)
		}
	}
}

func DeleteRegistry(w http.ResponseWriter, r *http.Request) {

}

func ScanRegistry(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM RegistryInfo")
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}

	//clear out image table before scanning
	_, err = db.Exec("DELETE FROM Image")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		item := models.RegistryInfo{}
		err2 := rows.Scan(&item.ID, &item.Name, &item.Url, &item.Org, &item.Username, &item.Password)
		if err2 != nil { panic(err2) }
		registry, err := registry.CreateRegistry(item.Url, item.Username, item.Password)
		if err != nil {
			humanError := checkError(err, item.Url, item.Username, item.Password)
			fmt.Fprint(w, humanError, "\n")
		}

		imgStrs, err := registry.Images(item.Org)

		images := []models.Image{}

		for _, img := range imgStrs {
			//TODO: get seed manifest
			image := models.Image{Name: img, Registry: item.Name, Org: item.Org, Manifest: ""}
			images = append(images, image)
			b, err := json.Marshal(img)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Fprint(w, string(b), "\n")
		}

		models.StoreImage(db, images)
	}
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