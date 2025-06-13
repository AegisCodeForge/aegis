package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bctnry/aegis/pkg/aegis"
	_ "github.com/mattn/go-sqlite3"
)

type AegisSqliteSessionStore struct {
	config *aegis.AegisConfig
	connection *sql.DB
}

func NewAegisSqliteSessionStore(cfg *aegis.AegisConfig) (*AegisSqliteSessionStore, error) {
	db, err := sql.Open("sqlite3", cfg.ProperSessionPath())
	if err != nil { return nil, err }
	return &AegisSqliteSessionStore{
		config: cfg,
		connection: db,
	}, nil
}

func (ss *AegisSqliteSessionStore) IsSessionStoreUsable() (bool, error) {
	tableName := fmt.Sprintf("%ssession", ss.config.Session.TablePrefix)
	stmt, err := ss.connection.Prepare("SELECT 1 FROM sqlite_schema WHERE type = 'table' AND name = ?")
	if err != nil { return false, err }
 	r := stmt.QueryRow(tableName)
	if r.Err() != nil { return false, r.Err() }
	var x string
	err = r.Scan(&x)
	if err == sql.ErrNoRows { return false, nil }
	if err != nil { return false, err }
	if len(x) <= 0 { return false, nil }
	return true, nil
}

func (ss *AegisSqliteSessionStore) RegisterSession(name string, session string) error {
	tx, err := ss.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %ssession(user_name, value, reg_timestamp) VALUES (?,?,?)", ss.config.Session.TablePrefix))
	if err != nil { tx.Rollback(); return err }
	_, err = stmt.Exec(name, session, time.Now().UTC())
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit();
	if err != nil { tx.Rollback(); return err }
	return nil
}

func (ss *AegisSqliteSessionStore) RetrieveSession(name string) (string, error) {
	stmt, err := ss.connection.Prepare(fmt.Sprintf("SELECT value FROM %ssession WHERE user_name = ?", ss.config.Session.TablePrefix))
	if err != nil { return "", err }
	s := ""
	err = stmt.QueryRow(name).Scan(&s)
	if err != nil { return "", err }
	return s, nil
}

func (ss *AegisSqliteSessionStore) VerifySession(name string, target string) (bool, error) {
	stmt, err := ss.connection.Prepare(fmt.Sprintf("SELECT 1 FROM %ssession WHERE user_name = ? AND value = ?", ss.config.Session.TablePrefix))
	if err != nil { return false, err }
	s := ""
	err = stmt.QueryRow(name, target).Scan(&s)
	if err == sql.ErrNoRows { return false, nil }
	if err != nil { return false, err }
	return (len(s) > 0), nil
}

func (ss *AegisSqliteSessionStore) RevokeSession(username string, target string) error {
	tx, err := ss.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare(fmt.Sprintf("DELETE FROM %ssession WHERE user_name = ? AND value = ?", ss.config.Session.TablePrefix))
	if err != nil { tx.Rollback(); return err }
	_, err = stmt.Exec(username, target)
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit()
	if err != nil { tx.Rollback(); return err }
	return nil
}

