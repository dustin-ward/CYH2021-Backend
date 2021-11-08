package auth

import "time"

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

type ActiveToken struct {
	Id      uint32
	Timeout time.Time
}

type AccessDetails struct {
	AccessUuid string
	UserId     uint32
}
