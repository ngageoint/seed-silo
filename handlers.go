package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/ngageoint/seed-silo/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/JohnPTobe/seed-common/registry"
	"github.com/JohnPTobe/seed-common/util"
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
	username := reginfo.Username
	password := reginfo.Password

	registry, err := registry.CreateRegistry(url, username, password)
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

	//clear out image table before scanning
	err = models.DeleteRegistryImages(db, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/text; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusAccepted)

	list := []models.RegistryInfo{}
	list = append(list, registry)
	Scan(w, r, list)
}

func ScanRegistries(w http.ResponseWriter, r *http.Request) {
	registries, err := models.GetRegistries(db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//clear out image table before scanning
	_, err = db.Exec("DELETE FROM Image")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/txt; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Scanning registries...")

	Scan(w, r, registries)
}

func Scan(w http.ResponseWriter, r *http.Request, registries []models.RegistryInfo) {
	for _, r := range registries {
		dbImages := []models.Image{}
		fmt.Fprintf(w, "Scanning registry %s... \n url: %s \n org: %s \n", r.Name, r.Url, r.Org)
		registry, err := registry.CreateRegistry(r.Url, r.Username, r.Password)
		if err != nil {
			humanError := checkError(err, r.Url, r.Username, r.Password)
			fmt.Fprint(w, humanError, "\n")
		}

		images, err := registry.ImagesWithManifests(r.Org)

		for _, img := range images {
			image := models.Image{Name: img.Name, Registry: img.Registry, Org: img.Org, Manifest: img.Manifest, RegistryId: r.ID}
			dbImages = append(dbImages, image)
			_, err := json.Marshal(img)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}

		models.StoreImage(db, dbImages)
	}
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
			if strings.Contains(img.Name, term) {
				score += 10
			}
			if strings.Contains(img.Org, term) {
				score += 10
			}
			seed := img.Seed

			if strings.Contains(fmt.Sprintf("%s", img.Seed), term) {
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

	results := []models.Image{}
	for _, res := range rankedResults {
		results = append(results, res.Image)
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

func Login(w http.ResponseWriter, r *http.Request) {
	//get user provided login and validate it
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	if err := r.Body.Close(); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	var user models.User
	if err := json.Unmarshal(body, &user); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error())
	}

	valid, err := models.ValidateUser(db, user.Username, user.Password)
	if !valid || err != nil{
		respondWithError(w, http.StatusUnauthorized, "Invalid login")
	}

	//get the user object from db with the role attribute and wrap it in a token
	displayuser, _ := models.GetUserByName(db, user.Username)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": displayuser.Username,
		"role": displayuser.Role,
	})

	tokenString, error := token.SignedString([]byte(TokenSecret))
	if error != nil {
		util.PrintUtil("Error signing token: %s\n", error.Error())
		respondWithError(w, http.StatusInternalServerError, "Error creating token")
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
					json.NewEncoder(w).Encode(models.Exception{Message: error.Error()})
					return
				}
				if token.Valid {
					context.Set(req, "decoded", token.Claims)
					var user models.User
					mapstructure.Decode(token.Claims, &user)
					fmt.Printf("user: %s\n", user.Username)
					if util.ContainsString(roles, user.Role) {
						next(w, req)
					} else {
						respondWithError(w, http.StatusUnauthorized, "User does not have permission to perform this action")
					}
				} else {
					respondWithError(w, http.StatusUnauthorized, "Invalid authorization token")
				}
			}
		} else {
			json.NewEncoder(w).Encode(models.Exception{Message: "An authorization header is required"})
		}
	})
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	if err := r.Body.Close(); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	var user models.User
	if err := json.Unmarshal(body, &user); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, err.Error())
	}

	id, err := models.AddUser(db, user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
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
	}

	list := []models.DisplayUser{}
	list = append(list, users...)
	respondWithJSON(w, http.StatusOK, list)
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
}
