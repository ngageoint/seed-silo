package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/JohnPTobe/seed-discover/models"
	"log"
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

	reginfolist := []models.RegistryInfo{}
	reginfolist = append(reginfolist, reginfo)
	err = models.StoreRegistry(db, reginfolist)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err != nil {
		panic(err)
	}
}

func DeleteRegistry(w http.ResponseWriter, r *http.Request) {

}

func ScanRegistry(w http.ResponseWriter, r *http.Request) {

}