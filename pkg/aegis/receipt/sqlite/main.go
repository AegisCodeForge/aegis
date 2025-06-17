package sqlite

import (
	"database/sql"
	"strings"
	"time"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/receipt"
	_ "github.com/mattn/go-sqlite3"
)

type AegisSqliteReceiptSystemInterface struct {
	config *aegis.AegisConfig
	connection *sql.DB
}

var requiredTableList = []string{
	"receipt",
}

func NewSqliteReceiptSystemInterface(cfg *aegis.AegisConfig) (*AegisSqliteReceiptSystemInterface, error) {
	db, err := sql.Open("sqlite3", cfg.ProperReceiptSystemPath())
	if err != nil { return nil, err }
	return &AegisSqliteReceiptSystemInterface{
		config: cfg,
		connection: db,
	}, nil
}

func (rsif *AegisSqliteReceiptSystemInterface) Dispose() error {
	return rsif.connection.Close()
}

func (rsif *AegisSqliteReceiptSystemInterface) IsReceiptSystemUsable() (bool, error) {
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

func (rsif *AegisSqliteReceiptSystemInterface) Install() error {
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

func (rsif *AegisSqliteReceiptSystemInterface) RetrieveReceipt(rid string) (*receipt.Receipt, error) {
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
		Command: receipt.ParseReceiptCommand(cmd),
		IssueTime: issueTime,
		TimeoutMinute: timeoutMinute,
	}, nil
}

func (rsif *AegisSqliteReceiptSystemInterface) IssueReceipt(timeoutMinute int64, command []string) (string, error) {
	tx, err := rsif.connection.Begin()
	if err != nil { return "", err }
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
INSERT INTO receipt(id, command, issue_time, timeout_minute)
VALUES (?,?,?,?)
`)
	if err != nil { return "", err }
	rid := receipt.NewReceiptId()
	_, err = stmt.Exec(rid, receipt.SerializeReceiptCommand(command), time.Now().Unix(), timeoutMinute)
	if err != nil { return "", err }
	err = tx.Commit()
	if err != nil { return "", err }
	return rid, nil
}

func (rsif *AegisSqliteReceiptSystemInterface) CancelReceipt(rid string) error {
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

func (rsif *AegisSqliteReceiptSystemInterface) GetAllReceipt(pageNum int, pageSize int) ([]*receipt.Receipt, error) {
	stmt, err := rsif.connection.Prepare(`
SELECT id, command, issue_time, timeout_minute
FROM receipt
ORDER BY rowid ASC
LIMIT ? OFFSET ?`)
	if err != nil { return nil, err }
	defer stmt.Close()
	r, err := stmt.Query(pageSize, pageNum * pageSize)
	if err != nil { return nil, err }
	defer r.Close()
	res := make([]*receipt.Receipt, 0)
	var id, command string
	var issueTime, timeoutMinute int64
	for r.Next() {
		err = r.Scan(&id, &command, &issueTime, &timeoutMinute)
		if err != nil { return nil, err }
		res = append(res, &receipt.Receipt{
			Id: id,
			Command: receipt.ParseReceiptCommand(command),
			IssueTime: issueTime,
			TimeoutMinute: timeoutMinute,
		})
	}
	return res, nil
}

func (rsif *AegisSqliteReceiptSystemInterface) SearchReceipt(q string, pageNum int, pageSize int) ([]*receipt.Receipt, error) {
	pattern := strings.ReplaceAll(q, "\\", "\\\\")
	pattern = strings.ReplaceAll(pattern, "%", "\\%")
	pattern = strings.ReplaceAll(pattern, "_", "\\_")
	pattern = "%" + pattern + "%"
	stmt, err := rsif.connection.Prepare(`
SELECT id, command, issue_time, timeout_minute
FROM receipt
WHERE id LIKE ? ESCAPE ? OR command LIKE ? ESCAPE ?
ORDER BY rowid ASC LIMIT ? OFFSET ?`)
	if err != nil { return nil, nil }
	defer stmt.Close()
	r, err := stmt.Query(pattern, "\\", pattern, "\\", pageSize, pageNum * pageSize)
	if err != nil { return nil, err }
	res := make([]*receipt.Receipt, 0)
	for r.Next() {
		var id, command string
		var issueTime, timeoutMinute int64
		err = r.Scan(&id, &command, &issueTime, &timeoutMinute)
		if err != nil { return nil, err }
		res = append(res, &receipt.Receipt{
			Id: id,
			Command: receipt.ParseReceiptCommand(command),
			IssueTime: issueTime,
			TimeoutMinute: timeoutMinute,
		})
	}
	return res, nil
}

func (rsif *AegisSqliteReceiptSystemInterface) EditReceipt(id string, robj *receipt.Receipt) error {
	tx, err := rsif.connection.Begin()
	if err != nil { return err }
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
UPDATE receipt
SET command = ?, issue_time = ?, timeout_minute = ?
WHERE id = ?
`)
	if err != nil { return err }
	defer stmt.Close()
	_, err = stmt.Exec(receipt.SerializeReceiptCommand(robj.Command), robj.IssueTime, robj.TimeoutMinute, robj.Id)
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

