package models

import (
	"database/sql"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"fmt"
)

type SiloUser struct {
	ID       int    `db:id`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type DisplayUser struct {
	ID       int    `db:id`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type JwtToken struct {
	Token string `json:"token"`
}

type Exception struct {
	Message string `json:"message"`
}

const AdminRole = "admin"

func CreateUser(db *sql.DB, dbType, admin, password string) {
	// create table if it does not exist
	sql_table := `
	CREATE TABLE IF NOT EXISTS SiloUser(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT,
		role TEXT
	);
	`
	if dbType == "postgres" {
	    sql_table = strings.Replace(sql_table, "id INTEGER PRIMARY KEY AUTOINCREMENT", "id SERIAL PRIMARY KEY", 1)
	}

	_, err := db.Exec(sql_table)
	if err != nil {
		panic(err)
	}

	users, _ := GetUsers(db)
	fmt.Println(len(users))
	if len(users) == 0 {
		//add default admin
		var admin= SiloUser{Username: admin, Password: password, Role: AdminRole}
		if dbType == "postgres" {
			_, err = AddUserPg(db, admin)
		} else {
			_, err = AddUserLite(db, admin)
		}

		if err != nil {
			panic(err)
		}
	}
}

func AddUserLite(db *sql.DB, r SiloUser) (int, error) {
	sql_addreg := `
	INSERT INTO SiloUser(
		username,
		password,
	    role
	) values(?, ?, ?)
	`

	stmt, err := db.Prepare(sql_addreg)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	hash, err := HashPassword(r.Password)

	result, err := stmt.Exec(r.Username, hash, r.Role)

	id := -1
	var id64 int64
	if err == nil {
		id64, err = result.LastInsertId()
		id = int(id64)
	}

	return id, err
}

func AddUserPg(db *sql.DB, r SiloUser) (int, error) {
	hash, err := HashPassword(r.Password)
	query := `INSERT INTO SiloUser(username, password, role) 
			VALUES($1, $2, $3) RETURNING id;`


	var userid int
	err = db.QueryRow(query, r.Username, hash, r.Role).Scan(&userid)

	return userid, err
}

func DeleteUser(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM SiloUser WHERE id=$1", id)

	return err
}

//Get list of users without username for display
func DisplayUsers(db *sql.DB) ([]DisplayUser, error) {
	sql_readall := `
	SELECT id, username, role FROM SiloUser
	ORDER BY id ASC
	`

	rows, err := db.Query(sql_readall)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DisplayUser
	for rows.Next() {
		item := DisplayUser{}
		err2 := rows.Scan(&item.ID, &item.Username, &item.Role)
		if err2 != nil {
			return nil, err
		}
		result = append(result, item)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}
	return result, err
}

func GetUserById(db *sql.DB, id int) (DisplayUser, error) {
	row := db.QueryRow("SELECT id, username, role FROM SiloUser WHERE id=$1", id)

	var item DisplayUser
	err := row.Scan(&item.ID, &item.Username, &item.Role)

	return item, err
}

func GetUserByName(db *sql.DB, username string) (DisplayUser, error) {
	row := db.QueryRow("SELECT id, username, role FROM SiloUser WHERE username=$1", username)

	var item DisplayUser
	err := row.Scan(&item.ID, &item.Username, &item.Role)

	return item, err
}

func GetUsers(db *sql.DB) ([]SiloUser, error) {
	rows, err := db.Query("SELECT * FROM SiloUser")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []SiloUser
	for rows.Next() {
		item := SiloUser{}
		err2 := rows.Scan(&item.ID, &item.Username, &item.Password, &item.Role)
		if err2 != nil {
			panic(err2)
		}
		result = append(result, item)
	}
	return result, err
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidateUser(db *sql.DB, username, password string) (bool, error) {
	row := db.QueryRow("SELECT * FROM SiloUser WHERE username=$1", username)

	var item SiloUser
	err := row.Scan(&item.ID, &item.Username, &item.Password, &item.Role)

	if err != nil {
		return false, err
	}

	return CheckPasswordHash(password, item.Password), nil
}
