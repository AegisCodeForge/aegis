package db

import "fmt"

type AegisDatabaseErrorType int

const (
	ENTITY_NOT_FOUND AegisDatabaseErrorType = 1
	ENTITY_ALREADY_EXISTS AegisDatabaseErrorType = 2
	DATABASE_NOT_SUPPORTED AegisDatabaseErrorType = 3
	NOT_ENOUGH_PERMISSION AegisDatabaseErrorType = 4
)

func (gdet AegisDatabaseErrorType) String() string {
	switch gdet {
	case ENTITY_NOT_FOUND: return "ENTITY_NOT_FOUND"
	case ENTITY_ALREADY_EXISTS: return "ENTITY_ALREADY_EXISTS"
	case NOT_ENOUGH_PERMISSION: return "NOT_ENOUGH_PERMISSION"
	}
	return "UNKNOWN_ERROR"
}

type AegisDatabaseError struct {
	ErrorType AegisDatabaseErrorType
	ErrorMsg string
}

func IsAegisDatabaseError(e error) bool {
	_, ok := e.(*AegisDatabaseError)
	return ok
}

func (gde AegisDatabaseError) Error() string {
	return fmt.Sprintf("%s: %s", gde.ErrorType, gde.ErrorMsg)
}

func NewAegisDatabaseError(t AegisDatabaseErrorType, msg string) *AegisDatabaseError {
	return &AegisDatabaseError{
		ErrorType: t,
		ErrorMsg: msg,
	}
}

var ErrNotEnoughPermission = NewAegisDatabaseError(NOT_ENOUGH_PERMISSION, "Not enough permission.")

