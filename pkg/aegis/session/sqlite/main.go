package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/session"
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

func (ss *AegisSqliteSessionStore) Dispose() error {
	return ss.connection.Close()
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

func (ss *AegisSqliteSessionStore) RetrieveSession(name string) ([]*session.AegisSession, error) {
	stmt, err := ss.connection.Prepare(fmt.Sprintf("SELECT value, reg_timestamp FROM %ssession WHERE user_name = ?", ss.config.Session.TablePrefix))
	if err != nil { return nil, err }
	res := make([]*session.AegisSession, 0)
	if err != nil { return nil, err }
	r, err := stmt.Query(name)
	for r.Next() {
		var id string
		var timestamp int64
		err = r.Scan(&id, &timestamp)
		if err != nil { return nil, err }
		res = append(res, &session.AegisSession{
			Username: name,
			Id: id,
			Timestamp: timestamp,
		})
	}
	return res, nil
}

func (ss *AegisSqliteSessionStore) RetrieveSessionByKey(username string, key string) (*session.AegisSession, error) {
	stmt, err := ss.connection.Prepare(fmt.Sprintf("SELECT reg_timestamp FROM %ssession WHERE user_name = ? AND value = ?", ss.config.Session.TablePrefix))
	if err != nil { return nil, err }
	r := stmt.QueryRow(username, key)
	if r.Err() != nil { return nil, r.Err() }
	var timestamp int64
	err = r.Scan(&timestamp)
	if err != nil { return nil, err }
	return &session.AegisSession{
		Username: username,
		Id: key,
		Timestamp: timestamp,
	}, nil
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

