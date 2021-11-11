package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dustin-ward/CYH2021-Backend/data"
	"github.com/dustin-ward/CYH2021-Backend/util"
	"github.com/gorilla/mux"
)

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: /users")

	rows, err := data.DB.Query("SELECT * FROM users")
	util.ErrHandle(err)
	defer rows.Close()

	uList := make([]data.User, 0)
	for rows.Next() {
		var u data.User
		err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.Password)
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

	rows, err := data.DB.Query("SELECT * FROM users WHERE id=?", id)
	util.ErrHandle(err)
	defer rows.Close()

	var u data.User
	for rows.Next() {
		err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.Password)
		util.ErrHandle(err)
	}
	util.RespondWithJSON(w, http.StatusOK, u)
}
