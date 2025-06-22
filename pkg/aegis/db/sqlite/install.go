package sqlite

import "fmt"

func (dbif *SqliteAegisDatabaseInterface) InstallTables() error {
	pfx := dbif.config.Database.TablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %suser (
    user_name TEXT UNIQUE,
    user_title TEXT,
    user_email TEXT,
    user_bio TEXT,
    user_website TEXT,
    user_reg_datetime TEXT,
    user_password_hash TEXT,
    user_status INTEGER
)`, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %suser_authkey (
    user_name TEXT,
    key_name TEXT,
    key_text TEXT,
    FOREIGN KEY (user_name) REFERENCES %suser(user_name)
)`, pfx, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %suser_signkey (
    user_name TEXT,
    key_name TEXT,
    key_text TEXT,
    FOREIGN KEY (user_name) REFERENCES %suser(user_name)
)`, pfx, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %snamespace (
      ns_name TEXT UNIQUE,
  	ns_title TEXT,
  	ns_description TEXT,
  	ns_email TEXT,
  	ns_owner TEXT,
  	ns_reg_datetime TEXT,
    ns_acl TEXT,
  	ns_status INTEGER,
      FOREIGN KEY (ns_owner) REFERENCES %suser(user_name)
)`, pfx, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %srepository (
      repo_fullname TEXT UNIQUE,
      repo_namespace TEXT,
      repo_name TEXT,
      repo_description TEXT,
      repo_owner TEXT,
  	  repo_acl TEXT,
  	  repo_status INTEGER,
  	repo_fork_origin_namespace TEXT,
  	repo_fork_origin_name TEXT,
      FOREIGN KEY (repo_namespace) REFERENCES %snamespace(ns_name)
)`, pfx, pfx))
	if err != nil { tx.Rollback(); return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %srepo_redirect (
    old_ns TEXT,
    old_name TEXT,
	new_ns TEXT,
	new_name TEXT,
	redirect_timestamp INTEGER
)`, pfx))
	if err != nil { tx.Rollback(); return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE UNIQUE INDEX IF NOT EXISTS idx_%suser_user_name
ON %suser (user_name)
`, pfx, pfx))
	if err != nil { tx.Rollback(); return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE INDEX IF NOT EXISTS idx_%suser_authkey_user_name
ON %suser_authkey (user_name);
`, pfx, pfx))
	if err != nil { tx.Rollback(); return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE INDEX IF NOT EXISTS idx_%suser_signkey_user_name
ON %suser_signkey (user_name);
`, pfx, pfx))
	if err != nil { tx.Rollback(); return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE INDEX IF NOT EXISTS idx_%snamespace_ns_name
ON %snamespace (ns_name);
`, pfx, pfx))
	if err != nil { tx.Rollback(); return err }
	
	tx.Commit()
	return nil
}

