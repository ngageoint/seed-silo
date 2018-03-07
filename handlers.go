package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/ngageoint/seed-common/registry"
	"github.com/ngageoint/seed-common/util"
	"github.com/ngageoint/seed-silo/models"
)

func Index(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, GetRoutes())
}

func Registry(w http.ResponseWriter, r *http.Request) {
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
	} else {
		id, err := models.AddRegistry(db, reginfo)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		reginfo.ID = id
		respondWithJSON(w, http.StatusCreated, reginfo)
	}
}

func DeleteRegistry(w http.ResponseWriter, r *http.Request) {
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

	//clear out job table
	err = models.ResetJobTable(db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.BuildJobsList(db, &dbImages)
	models.StoreImage(db, dbImages)
}

func ScanRegistries(w http.ResponseWriter, r *http.Request) {
	//prevent multiple requests to scan registries
	if sl.IsScanning() {
		respondWithJSON(w, http.StatusAccepted, map[string]string{"message": "Scanning Registries"})
		return
	}
	sl.StartScan()
	defer sl.EndScan()

	registries, err := models.GetRegistries(db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/txt; charset=UTF-8")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Scanning registries...")

	dbImages, err := Scan(w, r, registries)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//clear out image table before scanning
	err = models.ResetImageTable(db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//clear out job table
	err = models.ResetJobTable(db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.BuildJobsList(db, &dbImages)
	models.StoreImage(db, dbImages)
}

func Scan(w http.ResponseWriter, req *http.Request, registries []models.RegistryInfo) ([]models.Image, error) {
	dbImages := []models.Image{}
	for _, r := range registries {
		fmt.Fprintf(w, "Scanning registry %s... \n url: %s \n org: %s \n", r.Name, r.Url, r.Org)
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

		images, err := registry.ImagesWithManifests()

		for _, img := range images {
			image := models.Image{FullName: img.Name, Registry: img.Registry, Org: img.Org, Manifest: img.Manifest, RegistryId: r.ID}
			err := json.Unmarshal([]byte(img.Manifest), &image.Seed)
			if err != nil {
				log.Printf("Error unmarshalling seed manifest for %s: %s \n", img.Name, err.Error())
			}
			image.ShortName = image.Seed.Job.Name
			image.Title = image.Seed.Job.Title
			image.JobVersion = image.Seed.Job.JobVersion
			image.PackageVersion = image.Seed.Job.PackageVersion
			image.Description = image.Seed.Job.Description
			dbImages = append(dbImages, image)
		}
	}

	return dbImages, err
}

func ListImages(w http.ResponseWriter, r *http.Request) {
	imageList := []models.SimpleImage{}
	images := models.ReadSimpleImages(db)
	imageList = append(imageList, images...)

	respondWithJSON(w, http.StatusOK, imageList)
}

func ListRegistries(w http.ResponseWriter, r *http.Request) {
	registries, err := models.DisplayRegistries(db)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := []models.DisplayRegistry{}
	list = append(list, registries...)
	respondWithJSON(w, http.StatusOK, list)
}

type RankedResult struct {
	Score int
	Image models.Image
}

type ByScore []RankedResult

func (s ByScore) Len() int {
	return len(s)
}
func (s ByScore) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByScore) Less(i, j int) bool {
	return s[i].Score > s[j].Score
}

func SearchImages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["query"]

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

	results := []models.SimpleImage{}
	for _, res := range rankedResults {
		simple := models.SimplifyImage(res.Image)
		results = append(results, simple)
	}

	respondWithJSON(w, http.StatusOK, results)
}

func Image(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	img, err := models.ReadImage(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "No image found with that ID")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondWithJSON(w, http.StatusOK, img)
}

func ImageManifest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	image, err := models.ReadImage(db, id)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "No image found with that ID")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondWithJSON(w, http.StatusOK, image.Seed)
}

func ListJobs(w http.ResponseWriter, r *http.Request) {
	jobList := []models.Job{}
	jobs := models.ReadJobs(db)
	jobList = append(jobList, jobs...)

	respondWithJSON(w, http.StatusOK, jobList)
}

func Job(w http.ResponseWriter, r *http.Request) {
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

func Login(w http.ResponseWriter, r *http.Request) {
	//get user provided login and validate it
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := r.Body.Close(); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var user models.User
	if err := json.Unmarshal(body, &user); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	valid, err := models.ValidateUser(db, user.Username, user.Password)
	if !valid || err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid login")
		return
	}

	//get the user object from db with the role attribute and wrap it in a token
	displayuser, _ := models.GetUserByName(db, user.Username)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": displayuser.Username,
		"role":     displayuser.Role,
	})

	tokenString, error := token.SignedString([]byte(TokenSecret))
	if error != nil {
		util.PrintUtil("Error signing token: %s\n", error.Error())
		respondWithError(w, http.StatusInternalServerError, "Error creating token")
		return
	}

	respondWithJSON(w, http.StatusOK, models.JwtToken{Token: tokenString})
}

func Validate(roles []string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authorizationHeader := req.Header.Get("authorization")
		if authorizationHeader != "" {
			bearerToken := strings.Split(authorizationHeader, " ")
			if len(bearerToken) == 2 {
				token, error := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("There was an error")
					}
					return []byte(TokenSecret), nil
				})
				if error != nil {
					respondWithError(w, http.StatusInternalServerError, error.Error())
					return
				}
				if token.Valid {
					context.Set(req, "decoded", token.Claims)
					var user models.User
					mapstructure.Decode(token.Claims, &user)
					if util.ContainsString(roles, user.Role) {
						next(w, req)
					} else {
						respondWithError(w, http.StatusForbidden, "User does not have permission to perform this action")
					}
				} else {
					respondWithError(w, http.StatusUnauthorized, "Invalid authorization token")
				}
			} else {
				respondWithError(w, http.StatusUnauthorized, "Invalid authorization token format: expected 'token <token>'")
			}
		} else {
			respondWithError(w, http.StatusUnauthorized, "Missing authorization token")
		}
	})
}

func User(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	reg, err := models.GetUserById(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "No user found with that ID")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondWithJSON(w, http.StatusOK, reg)
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := r.Body.Close(); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var user models.User
	if err := json.Unmarshal(body, &user); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	id, err := models.AddUser(db, user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user.ID = id
	respondWithJSON(w, http.StatusCreated, user)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := models.DeleteUser(db, id); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := models.DisplayUsers(db)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := []models.DisplayUser{}
	list = append(list, users...)
	respondWithJSON(w, http.StatusOK, list)
}

func PreflightOptions(w http.ResponseWriter, r *http.Request) {
	// return 200 OK for all preflight CORS requests
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
	return
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

	util.PrintUtil("Response: %s\n", response)
}
