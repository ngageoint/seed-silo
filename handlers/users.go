package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/ngageoint/seed-common/util"
	"github.com/ngageoint/seed-silo/database"
	"github.com/ngageoint/seed-silo/models"
)

const TokenSecret = "Y4Y7)j)999>s(vDk"

func Login(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
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
	db := database.GetDB()
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
	db := database.GetDB()
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
	db := database.GetDB()
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
	db := database.GetDB()
	users, err := models.DisplayUsers(db)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	list := []models.DisplayUser{}
	list = append(list, users...)
	respondWithJSON(w, http.StatusOK, list)
}
