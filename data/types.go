package data

type User struct {
	Id       uint32 `json:"Id"`
	Email    string `json:"Email"`
	Username string `json:"Username"`
	Password string `json:"Password"`
}
