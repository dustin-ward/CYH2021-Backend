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

func CreateCalendar(id uint32) (data.Calendar, error) {
	// Insert calendar associated with user into DB
	var c data.Calendar
	res, err := data.DB.Exec("INSERT INTO calendars (user_id) VALUES (?)", id)
	if err != nil {
		return c, err
	}

	// Return Calendar object
	c_id, err := res.LastInsertId()
	if err != nil {
		return c, err
	}
	c.ID = uint32(c_id)
	c.User_ID = id
	return c, nil
}

func CreateDay(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: /day POST")

	// Read request body
	var d data.Day
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "invalid fields")
		return
	}

	// Get token metadata
	ad, err := auth.ExtractTokenMetadata(r)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "unable to extract token metadata")
		return
	}

	// Get calendar for user
	var c data.Calendar
	row := data.DB.QueryRow("SELECT * FROM calendars WHERE user_id=?", ad.UserId)
	err = row.Scan(&c.ID, c.User_ID)
	if err != nil {
		if err == sql.ErrNoRows {
			util.RespondWithError(w, http.StatusBadRequest, "unable to find calendar")
			return
		} else {
			util.RespondWithError(w, http.StatusInternalServerError, "error finding calendar")
			return
		}
	}

	// Insert day into table (or update)
	res, err := data.DB.Exec("INSERT INTO days (calendar_id, calendar_date, value) VALUES (?,?,?) ON DUPLICATE KEY UPDATE value=?", c.ID, d.Calendar_Date.Format("2006-01-02"), d.Value, d.Value)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "invalid fields")
		return
	}
	last_id, _ := res.LastInsertId()
	d.ID = uint32(last_id)

	// Return success
	fmt.Println("Date modified:", d.ID, d.Calendar_ID, d.Calendar_Date, d.Value)
	dJson, _ := json.Marshal(d)
	util.RespondWithJSON(w, http.StatusAccepted, map[string]string{"day": string(dJson)})
}
