package data

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dustin-ward/CYH2021-Backend/util"
	"github.com/gorilla/mux"
)

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: /users")

	rows, err := db.Query("SELECT * FROM users")
	util.ErrHandle(err)
	defer rows.Close()

	uList := make([]User, 0)
	for rows.Next() {
		var u User
		err := rows.Scan(&u.Id, &u.Email, &u.Username, &u.Password)
		util.ErrHandle(err)
		uList = append(uList, u)
	}

	util.RespondWithJSON(w, http.StatusOK, uList)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	util.ErrHandle(err)
	fmt.Println("Endpoint Hit: /users/" + vars["id"])

	rows, err := db.Query("SELECT * FROM users WHERE id=?", id)
	util.ErrHandle(err)
	defer rows.Close()

	var u User
	for rows.Next() {
		err := rows.Scan(&u.Id, &u.Email, &u.Username, &u.Password)
		util.ErrHandle(err)
	}
	util.RespondWithJSON(w, http.StatusOK, u)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: /register")

	// Check for email in database
	email := r.FormValue("email")
	row := db.QueryRow("SELECT COUNT(*) FROM users WHERE email=?", email)

	var count int
	err = row.Scan(&count)
	if count > 0 {
		util.RespondWithError(w, http.StatusConflict, "email already exists")
		return
	}

	// Gather account details and hash password
	username := r.FormValue("username")
	password := r.FormValue("password")
	hash, err := util.HashPassword(password)
	util.ErrHandle(err)
	password = hash

	// Insert new account into database
	_, err = db.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", email, username, password)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "unable to create new user")
	}

	fmt.Println("New user created:", email, username)
	util.RespondWithJSON(w, http.StatusCreated, map[string]string{"status": "Account successfully created!"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: /login")
	email := r.FormValue("email")
	row := db.QueryRow("SELECT * FROM users WHERE email=?", email)

	// Check to see if user exists
	var u User
	err = row.Scan(&u.Id, &u.Email, &u.Username, &u.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			util.RespondWithError(w, http.StatusNotFound, "no matching email exists")
			return
		} else {
			util.RespondWithError(w, http.StatusInternalServerError, "error finding account")
			return
		}
	}

	// Check Password
	password := r.FormValue("password")
	if !util.CheckPasswordHash(password, u.Password) {
		util.RespondWithError(w, http.StatusForbidden, "incorrect password")
		return
	}

	// Respond
	util.RespondWithJSON(w, http.StatusAccepted, map[string]string{"status": "login successful"})
	fmt.Println("User logged in:", u.Id, u.Email, u.Username)
}
