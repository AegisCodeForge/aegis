package sqlite

func (ss *AegisSqliteSessionStore) Install() error {
	tx, err := ss.connection.Begin()
	if err != nil { return err }
	_, err = tx.Exec(`
CREATE TABLE IF NOT EXISTS session (
    user_name TEXT,
    value TEXT,
    reg_timestamp INTEGER
)`)
	if err != nil { tx.Rollback(); return err }
	tx.Commit()
	return nil
}

