package data

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
	ID            uint32  `json:"id"`
	Calendar_ID   uint32  `json:"calendar_id"`
	Calendar_Date string  `json:"calendar_date"`
	Value         float64 `json:"value"`
}

type Task struct {
	ID          uint32 `json:"id"`
	Day_ID      uint32 `json:"day_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Task_Time   uint32 `json:"task_time"`
}

type Mood struct {
	ID     uint32  `json:"id"`
	Day_ID uint32  `json:"day_id"`
	Mood   string  `json:"mood"`
	Value  float64 `json:"value"`
}

/*
 * JSON Response Objects
 */

type DayResponse struct {
	DayObj Day    `json:"day"`
	Tasks  []Task `json:"tasks"`
	Moods  []Mood `json:"moods"`
}

type DayResponseArr struct {
	Days []DayResponse `json:"days"`
}

type RegisterResponse struct {
	Status      string   `json:"status"`
	UserObj     User     `json:"user"`
	CalendarObj Calendar `json:"calendar"`
}
