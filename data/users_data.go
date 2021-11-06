package data

import (
	"fmt"

	"github.com/dustin-ward/CYH2021-Backend/util"
)

func GetAllUsers() []User {
	rows, err := db.Query("SELECT * FROM users")
	util.ErrHandle(err)
	defer rows.Close()

	uList := make([]User, 0)
	for rows.Next() {
		var u User
		err := rows.Scan(&u.id, &u.email, &u.username, &u.password)
		util.ErrHandle(err)
		uList = append(uList, u)
	}
	return uList
}

func GetUser(id int32) User {
	fmt.Println("id = ", id)
	rows, err := db.Query("SELECT * FROM users WHERE id=?", id)
	util.ErrHandle(err)
	defer rows.Close()

	var u User
	for rows.Next() {
		err := rows.Scan(&u.id, &u.email, &u.username, &u.password)
		util.ErrHandle(err)
	}
	return u
}
