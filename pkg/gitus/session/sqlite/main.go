package sqlite

import (
	"database/sql"
	"time"

	"github.com/bctnry/gitus/pkg/gitus"
	_ "github.com/mattn/go-sqlite3"
)

type GitusSqliteSessionStore struct {
	config *gitus.GitusConfig
	connection *sql.DB
}

func NewGitusSqliteSessionStore(cfg *gitus.GitusConfig) (*GitusSqliteSessionStore, error) {
	db, err := sql.Open("sqlite3", cfg.SessionPath)
	if err != nil { return nil, err }
	return &GitusSqliteSessionStore{
		config: cfg,
		connection: db,
	}, nil
}

func (ss *GitusSqliteSessionStore) IsSessionStoreUsable() (bool, error) {
	var x string
 	err := ss.connection.QueryRow("SELECT 1 FROM sqlite_schema WHERE type = 'table' AND name = 'session'").Scan(&x)
	if err == sql.ErrNoRows { return false, nil }
	if err != nil { return false, err }
	if len(x) <= 0 { return false, nil }
	return true, nil
}

func (ss *GitusSqliteSessionStore) RegisterSession(name string, session string) error {
	tx, err := ss.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare("INSERT INTO session(user_name, value, reg_timestamp) VALUES (?,?,?)")
	if err != nil { tx.Rollback(); return err }
	_, err = stmt.Exec(name, session, time.Now().UTC())
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit();
	if err != nil { tx.Rollback(); return err }
	return nil
}

func (ss *GitusSqliteSessionStore) RetrieveSession(name string) (string, error) {
	stmt, err := ss.connection.Prepare("SELECT value FROM session WHERE user_name = ?")
	if err != nil { return "", err }
	s := ""
	err = stmt.QueryRow(name).Scan(&s)
	if err != nil { return "", err }
	return s, nil
}

func (ss *GitusSqliteSessionStore) VerifySession(name string, target string) (bool, error) {
	stmt, err := ss.connection.Prepare("SELECT 1 FROM session WHERE user_name = ? AND value = ?")
	if err != nil { return false, err }
	s := ""
	err = stmt.QueryRow(name, target).Scan(&s)
	if err == sql.ErrNoRows { return false, nil }
	if err != nil { return false, err }
	return (len(s) > 0), nil
}

func (ss *GitusSqliteSessionStore) RevokeSession(target string) error {
	tx, err := ss.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare("DELETE FROM session WHERE value = ?")
	if err != nil { tx.Rollback(); return err }
	_, err = stmt.Exec(target)
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit()
	if err != nil { tx.Rollback(); return err }
	return nil
}

