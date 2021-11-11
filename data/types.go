package data

import "time"

type User struct {
	ID       uint32 `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Calendar struct {
	ID      uint32 `json:"id"`
	User_ID uint32 `json:"user_id"`
}

type Day struct {
	ID            uint32    `json:"id"`
	Calendar_ID   uint32    `json:"calendar_id"`
	Calendar_Date time.Time `json:"calendar_date"`
	Value         float64   `json:"value"`
}

type Task struct {
	ID          uint32    `json:"id"`
	Day_ID      uint32    `json:"day_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Task_Time   time.Time `json:"task_time"`
}

type Mood struct {
	ID     uint32  `json:"id"`
	Day_ID uint32  `json:"day_id"`
	Mood   string  `json:"mood"`
	Value  float64 `json:"value"`
}
