package models

import (
	"database/sql"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
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

func CreateUser(db *sql.DB) {
	// create table if it does not exist
	sql_table := `
	CREATE TABLE IF NOT EXISTS User(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT,
		role TEXT
	);
	`

	_, err := db.Exec(sql_table)
	if err != nil {
		panic(err)
	}

	users, _ := GetUsers(db)
	if len(users) == 0 {
		//add default admin
		var admin= User{Username: "admin", Password: "spicy-pickles17!", Role: AdminRole}
		_, err = AddUser(db, admin)

		if err != nil {
			panic(err)
		}
	}
}

func AddUser(db *sql.DB, r User) (int, error) {
	sql_addreg := `
	INSERT OR REPLACE INTO User(
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

func DeleteUser(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM User WHERE id=$1", id)

	return err
}

//Get list of users without username for display
func DisplayUsers(db *sql.DB) ([]DisplayUser, error) {
	sql_readall := `
	SELECT id, username, role FROM User
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
	row := db.QueryRow("SELECT id, username, role FROM User WHERE id=?", id)

	var item DisplayUser
	err := row.Scan(&item.ID, &item.Username, &item.Role)

	return item, err
}

func GetUserByName(db *sql.DB, username string) (DisplayUser, error) {
	row := db.QueryRow("SELECT id, username, role FROM User WHERE username=?", username)

	var item DisplayUser
	err := row.Scan(&item.ID, &item.Username, &item.Role)

	return item, err
}

func GetUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT * FROM User")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []User
	for rows.Next() {
		item := User{}
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
	row := db.QueryRow("SELECT * FROM User WHERE username=?", username)

	var item User
	err := row.Scan(&item.ID, &item.Username, &item.Password, &item.Role)

	if err != nil {
		return false, err
	}

	return CheckPasswordHash(password, item.Password), nil
}
