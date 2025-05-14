package sqlite

import (
	"database/sql"
	"strings"
	"time"

	"github.com/bctnry/gitus/pkg/gitus"
	"github.com/bctnry/gitus/pkg/gitus/receipt"
	_ "github.com/mattn/go-sqlite3"
)

type GitusSqliteReceiptSystemInterface struct {
	config *gitus.GitusConfig
	connection *sql.DB
}

var requiredTableList = []string{
	"receipt",
}

func NewSqliteReceiptSystemInterface(cfg *gitus.GitusConfig) (*GitusSqliteReceiptSystemInterface, error) {
	db, err := sql.Open("sqlite3", cfg.ReceiptSystem.Path)
	if err != nil { return nil, err }
	return &GitusSqliteReceiptSystemInterface{
		config: cfg,
		connection: db,
	}, nil
}

func (rsif *GitusSqliteReceiptSystemInterface) IsReceiptSystemUsable() (bool, error) {
	stmt, err := rsif.connection.Prepare("SELECT 1 FROM sqlite_schema WHERE type = 'table' AND name = ?")
	if err != nil { return false, err }
	for _, item := range requiredTableList {
		r := stmt.QueryRow(item)
		if r.Err() != nil { return false, r.Err() }
		var a string
		err := r.Scan(&a)
		if err == sql.ErrNoRows { return false, nil }
		if err != nil { return false, err }
		if len(a) <= 0 { return false, nil }
	}
	return true, nil
}

func (rsif *GitusSqliteReceiptSystemInterface) Install() error {
	tx, err := rsif.connection.Begin()
	if err != nil { return err }
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
CREATE TABLE IF NOT EXISTS receipt (
    id TEXT UNIQUE,
    command TEXT,
    issue_time INT,
    timeout_minute INT
)`)
	if err != nil { return err }
	_, err = stmt.Exec()
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

func (rsif *GitusSqliteReceiptSystemInterface) RetrieveReceipt(rid string) (*receipt.Receipt, error) {
	stmt, err := rsif.connection.Prepare(`
SELECT command, issue_time, timeout_minute
FROM receipt
WHERE id = ?
`)
	if err != nil { return nil, err }
	defer stmt.Close()
	r := stmt.QueryRow(rid)
	if r.Err() != nil { return nil, err }
	var cmd string
	var issueTime, timeoutMinute int64
	err = r.Scan(&cmd, &issueTime, &timeoutMinute)
	if err != nil { return nil, err }
	return &receipt.Receipt{
		Id: rid,
		Command: strings.Split(cmd, ","),
		IssueTime: issueTime,
		TimeoutMinute: timeoutMinute,
	}, nil
}

func (rsif *GitusSqliteReceiptSystemInterface) IssueReceipt(timeoutMinute int64, command []string) (string, error) {
	tx, err := rsif.connection.Begin()
	if err != nil { return "", err }
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
INSERT INTO receipt(id, command, issue_time, timeout_minute)
VALUES (?,?,?,?)
`)
	if err != nil { return "", err }
	rid := receipt.NewReceiptId()
	_, err = stmt.Exec(rid, strings.Join(command, ","), time.Now().Unix(), timeoutMinute)
	if err != nil { return "", err }
	err = tx.Commit()
	if err != nil { return "", err }
	return rid, nil
}

func (rsif *GitusSqliteReceiptSystemInterface) CancelReceipt(rid string) error {
	tx, err := rsif.connection.Begin()
	if err != nil { return err }
	defer tx.Rollback()
	stmt, err := tx.Prepare(`DELETE FROM receipt WHERE id = ?`)
	if err != nil { return err }
	_, err = stmt.Exec(rid)
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	return nil
}


