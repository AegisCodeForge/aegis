package sqlite

import (
	"database/sql"

	"github.com/bctnry/gitus/pkg/gitus"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteGitusDatabaseInterface struct {
	config *gitus.GitusConfig
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

func (dbif *SqliteGitusDatabaseInterface) IsDatabaseUsable() (bool, error) {
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

func (dbif *SqliteGitusDatabaseInterface) Close() {
	dbif.connection.Close()
}

func NewSqliteGitusDatabaseInterface(cfg *gitus.GitusConfig) (*SqliteGitusDatabaseInterface, error) {
	db, err := sql.Open("sqlite3", cfg.DatabasePath)
	if err != nil { return nil, err }
	return &SqliteGitusDatabaseInterface{
		config: cfg,
		connection: db,
	}, nil
}

