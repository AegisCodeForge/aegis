package db

import "fmt"

type GitusDatabaseErrorType int

const (
	ENTITY_NOT_FOUND GitusDatabaseErrorType = 1
	ENTITY_ALREADY_EXISTS GitusDatabaseErrorType = 2
	DATABASE_NOT_SUPPORTED GitusDatabaseErrorType = 3
)

func (gdet GitusDatabaseErrorType) String() string {
	switch gdet {
	case ENTITY_NOT_FOUND: return "ENTITY_NOT_FOUND"
	case ENTITY_ALREADY_EXISTS: return "ENTITY_ALREADY_EXISTS"
	}
	return "UNKNOWN_ERROR"
}

type GitusDatabaseError struct {
	ErrorType GitusDatabaseErrorType
	ErrorMsg string
}

func IsGitusDatabaseError(e error) bool {
	_, ok := e.(*GitusDatabaseError)
	return ok
}

func (gde GitusDatabaseError) Error() string {
	return fmt.Sprintf("%s: %s", gde.ErrorType, gde.ErrorMsg)
}

func NewGitusDatabaseError(t GitusDatabaseErrorType, msg string) *GitusDatabaseError {
	return &GitusDatabaseError{
		ErrorType: t,
		ErrorMsg: msg,
	}
}

