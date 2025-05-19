package sqlite

func (ss *AegisSqliteSessionStore) Install() error {
	tx, err := ss.connection.Begin()
	if err != nil { return err }
	defer tx.Rollback()
	_, err = tx.Exec(`
CREATE TABLE IF NOT EXISTS session (
    user_name TEXT,
    value TEXT,
    reg_timestamp INTEGER
)`)
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

