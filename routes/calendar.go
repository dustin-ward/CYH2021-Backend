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

func GetDayHelper(id uint32) (data.DayResponse, error) {
	var dr data.DayResponse

	// Get data directly associated with day
	row := data.DB.QueryRow("SELECT * FROM days WHERE id=?", id)

	// var d data.Day
	if err := row.Scan(&dr.DayObj.ID, &dr.DayObj.Calendar_ID, &dr.DayObj.Calendar_Date, &dr.DayObj.Value); err != nil {
		return dr, err
	}

	// Get all moods associated with day
	rows, err := data.DB.Query("SELECT * FROM moods WHERE day_id=?", id)
	if err != nil {
		return dr, err
	}

	sum := 0.0
	for rows.Next() {
		var m data.Mood
		if err := rows.Scan(&m.ID, &m.Day_ID, &m.Mood, &m.Value); err != nil {
			return dr, err
		}

		sum += m.Value
		dr.Moods = append(dr.Moods, m)
	}
	dr.DayObj.Value = sum
	rows.Close()

	// Get all tasks associated with day
	rows, err = data.DB.Query("SELECT * FROM tasks WHERE day_id=?", id)
	if err != nil {
		return dr, err
	}

	for rows.Next() {
		var t data.Task
		if err := rows.Scan(&t.ID, &t.Day_ID, &t.Title, &t.Description, &t.Task_Time); err != nil {
			return dr, err
		}

		dr.Tasks = append(dr.Tasks, t)
	}
	rows.Close()

	return dr, nil
}

func UpdateMoods(new, old *[]data.Mood) (float64, error) {
	// Set of new elements
	s := make(map[string]struct{}, len(*new))
	for _, x := range *new {
		s[x.Mood] = struct{}{}
	}

	// Delete elements that no longer exist
	for _, x := range *old {
		if _, found := s[x.Mood]; !found {
			// Delete mood
			if _, err := data.DB.Exec("DELETE FROM moods WHERE id=?", x.ID); err != nil {
				return 0.0, err
			}
		}
	}

	// Update/Create new elements
	sum := 0.0
	for _, x := range *new {
		// Insert day into table (or update)
		res, err := data.DB.Exec("INSERT INTO moods (day_id, mood, value) VALUES (?,?,?) ON DUPLICATE KEY UPDATE mood=?, value=?", x.Day_ID, x.Mood, x.Value, x.Mood, x.Value)
		if err != nil {
			return 0.0, err
		}
		last_id, _ := res.LastInsertId()
		x.ID = uint32(last_id)
		sum += x.Value
	}

	return sum, nil
}

func UpdateTasks(new, old *[]data.Task) error {
	// Set of new elements
	s := make(map[string]struct{}, len(*new))
	for _, x := range *new {
		s[x.Title] = struct{}{}
	}

	// Delete elements that no longer exist
	for _, x := range *old {
		if _, found := s[x.Title]; !found {
			// Delete mood
			if _, err := data.DB.Exec("DELETE FROM tasks WHERE id=?", x.ID); err != nil {
				return err
			}
		}
	}

	// Update/Create new elements
	for _, x := range *new {
		// Insert day into table (or update)
		res, err := data.DB.Exec("INSERT INTO tasks (day_id, title, description, task_time) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE title=?, description=?, task_time=?", x.Day_ID, x.Title, x.Description, x.Task_Time, x.Title, x.Description, x.Task_Time)
		if err != nil {
			return err
		}
		last_id, _ := res.LastInsertId()
		x.ID = uint32(last_id)
	}

	return nil
}

func CreateDay(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: /days POST")

	// Read request body
	var d data.DayResponse
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
	err = row.Scan(&c.ID, &c.User_ID)
	if err != nil {
		if err == sql.ErrNoRows {
			util.RespondWithError(w, http.StatusInternalServerError, "unable to find calendar")
			return
		} else {
			util.RespondWithError(w, http.StatusInternalServerError, "error finding calendar")
			return
		}
	}

	var OrigDayRes data.DayResponse
	if d.DayObj.ID == 0 {
		// Insert day into table (or update)
		res, err := data.DB.Exec("INSERT INTO days (calendar_id, calendar_date, value) VALUES (?,?,?)", c.ID, d.DayObj.Calendar_Date, 0.0)
		if err != nil {
			util.RespondWithError(w, http.StatusBadRequest, "invalid fields")
			return
		}
		last_id, _ := res.LastInsertId()
		d.DayObj.ID = uint32(last_id)
	} else {
		// Get original day if exists
		OrigDayRes, err = GetDayHelper(d.DayObj.ID)
		if err != nil {
			util.RespondWithError(w, http.StatusBadRequest, "error finding day in database")
			return
		}
	}

	// Set day_id's
	for i := range d.Moods {
		d.Moods[i].Day_ID = d.DayObj.ID
	}
	for i := range d.Tasks {
		d.Tasks[i].Day_ID = d.DayObj.ID
	}

	// Compare Moods
	sum, err := UpdateMoods(&d.Moods, &OrigDayRes.Moods)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	d.DayObj.Value = sum

	// Compare Tasks
	if err := UpdateTasks(&d.Tasks, &OrigDayRes.Tasks); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "unable to update tasks")
		return
	}

	// Return success
	fmt.Println("Date modified:", d.DayObj.ID, d.DayObj.Calendar_ID, d.DayObj.Calendar_Date, d.DayObj.Value)
	util.RespondWithJSON(w, http.StatusAccepted, d)
}

func GetDays(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: /days GET")

	// Get token metadata
	ad, err := auth.ExtractTokenMetadata(r)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "unable to extract token metadata")
		return
	}

	// Get calendar for user
	var c data.Calendar
	row := data.DB.QueryRow("SELECT * FROM calendars WHERE user_id=?", ad.UserId)
	err = row.Scan(&c.ID, &c.User_ID)
	if err != nil {
		if err == sql.ErrNoRows {
			util.RespondWithError(w, http.StatusInternalServerError, "unable to find calendar")
			return
		} else {
			util.RespondWithError(w, http.StatusInternalServerError, "error finding calendar")
			return
		}
	}

	// Find all days associated with calendar
	rows, err := data.DB.Query("SELECT * FROM days WHERE calendar_id=?", c.ID)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "unable to fund calendar")
		return
	}

	// For each day...
	var Response data.DayResponseArr
	for rows.Next() {
		var dr data.DayResponse
		var d data.Day

		// Save day data
		if err := rows.Scan(&d.ID, &d.Calendar_ID, &d.Calendar_Date, &d.Value); err != nil {
			util.RespondWithError(w, http.StatusInternalServerError, "unable to scan day")
			return
		}

		// Get moods ans
		dr, err = GetDayHelper(d.ID)
		if err != nil {
			util.RespondWithError(w, http.StatusInternalServerError, "unable to retreive days")
			return
		}
		dr.DayObj.ID = d.ID
		dr.DayObj.Calendar_ID = d.Calendar_ID
		dr.DayObj.Calendar_Date = d.Calendar_Date

		Response.Days = append(Response.Days, dr)
	}

	util.RespondWithJSON(w, http.StatusOK, Response)
}
