package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dustin-ward/CYH2021-Backend/auth"
	"github.com/dustin-ward/CYH2021-Backend/data"
	"github.com/dustin-ward/CYH2021-Backend/util"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: /register")

	var u data.User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "invalid fields")
		return
	}

	// Check for email in database
	row := data.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email=?", u.Email)

	var count int
	row.Scan(&count)
	if count > 0 {
		util.RespondWithError(w, http.StatusConflict, "email already exists")
		return
	}

	// Gather account details and hash password
	hash, err := auth.HashPassword(u.Password)
	util.ErrHandle(err)
	u.Password = hash

	// Insert new account into database
	res, err := data.DB.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", u.Email, u.Username, u.Password)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "unable to create new user")
		return
	}
	last_id, err := res.LastInsertId()
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "unable to get id associated with new user")
		return
	}
	u.ID = uint32(last_id)

	fmt.Println("New user created:", u.Email, u.Username)
	calendar, err := CreateCalendar(u.ID)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "unable to create calendar for user")
	}

	uJson, _ := json.Marshal(u)
	cJson, _ := json.Marshal(calendar)
	util.RespondWithJSON(w, http.StatusCreated, map[string]string{
		"status":   "account successfully created",
		"user":     string(uJson),
		"calendar": string(cJson),
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: /login")

	var u1 data.User
	err := json.NewDecoder(r.Body).Decode(&u1)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "invalid fields")
		return
	}

	row := data.DB.QueryRow("SELECT * FROM users WHERE email=?", u1.Email)

	// Check to see if user exists
	var u2 data.User
	err = row.Scan(&u2.ID, &u2.Email, &u2.Username, &u2.Password)
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
	if !auth.CheckPasswordHash(u1.Password, u2.Password) {
		util.RespondWithError(w, http.StatusForbidden, "incorrect password")
		return
	}

	// Respond
	tokens, err := auth.CreateToken(u2.ID)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "unable to create token")
		return
	}
	auth.CreateAuth(u2.ID, tokens)
	util.RespondWithJSON(w, http.StatusAccepted, map[string]string{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
	fmt.Println("User logged in:", u2.ID, u2.Email, u2.Username)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: /logout")
	access, err := auth.ExtractTokenMetadata(r)
	if err != nil {
		util.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	delID, err := auth.DeleteAuth(access.AccessUuid)
	if err != nil || delID == 0 {
		util.RespondWithJSON(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	util.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "successfully logged out"})
}
