package data

import (
	"database/sql"
	"fmt"

	"github.com/dustin-ward/CYH2021-Backend/util"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var err error

func Setup() {
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@%s/%s?parseTime=true", username, password, host, dbname))
	util.ErrHandle(err)
}

func Takedown() {
	db.Close()
}
