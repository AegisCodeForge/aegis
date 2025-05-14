package sqlite

import (
	"database/sql"
	"time"

	"github.com/bctnry/aegis/pkg/aegis"
	_ "github.com/mattn/go-sqlite3"
)

type AegisSqliteSessionStore struct {
	config *aegis.AegisConfig
	connection *sql.DB
}

func NewAegisSqliteSessionStore(cfg *aegis.AegisConfig) (*AegisSqliteSessionStore, error) {
	db, err := sql.Open("sqlite3", cfg.SessionPath)
	if err != nil { return nil, err }
	return &AegisSqliteSessionStore{
		config: cfg,
		connection: db,
	}, nil
}

func (ss *AegisSqliteSessionStore) IsSessionStoreUsable() (bool, error) {
	var x string
 	err := ss.connection.QueryRow("SELECT 1 FROM sqlite_schema WHERE type = 'table' AND name = 'session'").Scan(&x)
	if err == sql.ErrNoRows { return false, nil }
	if err != nil { return false, err }
	if len(x) <= 0 { return false, nil }
	return true, nil
}

func (ss *AegisSqliteSessionStore) RegisterSession(name string, session string) error {
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

func (ss *AegisSqliteSessionStore) RetrieveSession(name string) (string, error) {
	stmt, err := ss.connection.Prepare("SELECT value FROM session WHERE user_name = ?")
	if err != nil { return "", err }
	s := ""
	err = stmt.QueryRow(name).Scan(&s)
	if err != nil { return "", err }
	return s, nil
}

func (ss *AegisSqliteSessionStore) VerifySession(name string, target string) (bool, error) {
	stmt, err := ss.connection.Prepare("SELECT 1 FROM session WHERE user_name = ? AND value = ?")
	if err != nil { return false, err }
	s := ""
	err = stmt.QueryRow(name, target).Scan(&s)
	if err == sql.ErrNoRows { return false, nil }
	if err != nil { return false, err }
	return (len(s) > 0), nil
}

func (ss *AegisSqliteSessionStore) RevokeSession(target string) error {
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

