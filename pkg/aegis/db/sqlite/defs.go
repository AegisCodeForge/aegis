package sqlite

import (
	"database/sql"

	"github.com/bctnry/aegis/pkg/aegis"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteAegisDatabaseInterface struct {
	config *aegis.AegisConfig
	connection *sql.DB
}

var requiredTableList = []string{
	"user_authkey",
	"user_signkey",
	"user",
	"namespace",
	"repository",
	"repo_redirect",
}

func (dbif *SqliteAegisDatabaseInterface) IsDatabaseUsable() (bool, error) {
	stmt, err := dbif.connection.Prepare("SELECT 1 FROM sqlite_schema WHERE type = 'table' AND name = ?")
	for _, item := range requiredTableList {
		tableName := dbif.config.DatabaseTablePrefix + item
		if err != nil { return false, err }
		r := stmt.QueryRow(tableName)
		if r.Err() != nil { return false, r.Err() }
		var a string
		err := r.Scan(&a)
		if err == sql.ErrNoRows { return false, nil }
		if err != nil { return false, err }
		if len(a) <= 0 { return false, nil }
	}
	return true, nil
}

func (dbif *SqliteAegisDatabaseInterface) Close() {
	dbif.connection.Close()
}

func NewSqliteAegisDatabaseInterface(cfg *aegis.AegisConfig) (*SqliteAegisDatabaseInterface, error) {
	db, err := sql.Open("sqlite3", cfg.DatabasePath)
	if err != nil { return nil, err }
	return &SqliteAegisDatabaseInterface{
		config: cfg,
		connection: db,
	}, nil
}

