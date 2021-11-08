package auth

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/dustin-ward/CYH2021-Backend/data"
	"github.com/dustin-ward/CYH2021-Backend/util"
)

var ActiveTokens map[string]ActiveToken = make(map[string]ActiveToken)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: /register")

	// Check for email in database
	email := r.FormValue("email")
	row := data.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email=?", email)

	var count int
	row.Scan(&count)
	if count > 0 {
		util.RespondWithError(w, http.StatusConflict, "email already exists")
		return
	}

	// Gather account details and hash password
	username := r.FormValue("username")
	password := r.FormValue("password")
	hash, err := HashPassword(password)
	util.ErrHandle(err)
	password = hash

	// Insert new account into database
	_, err = data.DB.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", email, username, password)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "unable to create new user")
		return
	}

	fmt.Println("New user created:", email, username)
	util.RespondWithJSON(w, http.StatusCreated, map[string]string{"status": "account successfully created"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: /login")
	email := r.FormValue("email")
	row := data.DB.QueryRow("SELECT * FROM users WHERE email=?", email)

	// Check to see if user exists
	var u data.User
	err := row.Scan(&u.Id, &u.Email, &u.Username, &u.Password)
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
	if !CheckPasswordHash(password, u.Password) {
		util.RespondWithError(w, http.StatusForbidden, "incorrect password")
		return
	}

	// Respond
	tokens, err := CreateToken(u.Id)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "unable to create token")
		return
	}
	CreateAuth(u.Id, tokens)
	util.RespondWithJSON(w, http.StatusAccepted, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
	fmt.Println("User logged in:", u.Id, u.Email, u.Username)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: /logout")
	access, err := ExtractTokenMetadata(r)
	if err != nil {
		util.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	delID, err := DeleteAuth(access.AccessUuid)
	if err != nil || delID == 0 {
		util.RespondWithJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	util.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "successfully logged out"})
}
