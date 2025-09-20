package sqlite

import (
	"database/sql"
	"net/url"

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
	"user_email",
	"user_reg_request",
	"user",
	"namespace",
	"repository",
	"issue",
	"issue_event",
	"pull_request",
	"pull_request_event",
	"snippet",
	"webhook_log",
}

func (dbif *SqliteAegisDatabaseInterface) IsDatabaseUsable() (bool, error) {
	stmt, err := dbif.connection.Prepare("SELECT 1 FROM sqlite_schema WHERE type = 'table' AND name = ?")
	if err != nil { return false, err }
	for _, item := range requiredTableList {
		tableName := dbif.config.Database.TablePrefix + "_" + item
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

func NewSqliteAegisDatabaseInterface(cfg *aegis.AegisConfig) (*SqliteAegisDatabaseInterface, error) {
	p := cfg.ProperDatabasePath()
	r, _ := url.Parse(p)
	q := r.Query()
	q.Set("cache", "shared")
	q.Set("mode", "rwc")
	q.Set("_journal_mode", "WAL")
	r.RawQuery = q.Encode()
	db, err := sql.Open("sqlite3", r.String())
	if err != nil { return nil, err }
	return &SqliteAegisDatabaseInterface{
		config: cfg,
		connection: db,
	}, nil
}
