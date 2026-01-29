package model

import "time"


type RegistrationRequest struct {
	AbsId int64
	Username string
	Email string
	PasswordHash string
	Reason string
	Timestamp time.Time
}



