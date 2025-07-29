package sqlite

import (
	"database/sql"
	"net/url"
	"strings"

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
	"issue",
	"issue_event",
	"pull_request",
	"pull_request_event",
}

func (dbif *SqliteAegisDatabaseInterface) IsDatabaseUsable() (bool, error) {
	stmt, err := dbif.connection.Prepare("SELECT 1 FROM sqlite_schema WHERE type = 'table' AND name = ?")
	for _, item := range requiredTableList {
		tableName := dbif.config.Database.TablePrefix + item
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

func ToSqlSearchPattern(s string) string {
	res := strings.ReplaceAll(s, "\\", "\\\\")
	res = strings.ReplaceAll(s, "%", "\\%")
	res = strings.ReplaceAll(s, "_", "\\_")
	res = "%" + res + "%"
	return res
}

