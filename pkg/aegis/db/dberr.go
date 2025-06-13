package db

import (
	"errors"
)

var ErrEntityNotFound = errors.New("ENTITY_NOT_FOUND: The requested entity is not found")
var ErrEntityAlreadyExists = errors.New("ENTITY_ALREADY_EXISTS: The requested entity already exists")
var ErrDatabaseNotSupported = errors.New("DATABASE_NOT_SUPPORTED: The version of Aegis you have does not support the specified type of database.")
var ErrNotEnoughPermission = errors.New("NOT_ENOUGH_PERMISSION: Not enough permission.")

