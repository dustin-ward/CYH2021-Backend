package data

import (
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
	var u User
	u.Email = r.FormValue("email")
	u.Username = r.FormValue("username")
	u.Password = r.FormValue("password")
	hash, err := util.HashPassword(u.Password)
	util.ErrHandle(err)
	u.Password = hash

	// SQL Query

	fmt.Println(u)
	util.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "Account successfully created!"})
}
