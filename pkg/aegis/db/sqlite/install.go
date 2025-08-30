package sqlite

import "fmt"


func (dbif *SqliteAegisDatabaseInterface) InstallTables() error {
	pfx := dbif.config.Database.TablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	defer tx.Rollback()
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
	if err != nil { return err }
	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %suser_authkey (
    user_name TEXT,
    key_name TEXT,
    key_text TEXT,
    FOREIGN KEY (user_name) REFERENCES %suser(user_name)
)`, pfx, pfx))
	if err != nil { return err }
	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %suser_signkey (
    user_name TEXT,
    key_name TEXT,
    key_text TEXT,
    FOREIGN KEY (user_name) REFERENCES %suser(user_name)
)`, pfx, pfx))
	if err != nil { return err }
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
	if err != nil { return err }
	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %srepository (
    repo_type INTEGER,
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
	if err != nil { return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %srepo_redirect (
    old_ns TEXT,
    old_name TEXT,
	new_ns TEXT,
	new_name TEXT,
	redirect_timestamp INTEGER
)`, pfx))
	if err != nil { return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE UNIQUE INDEX IF NOT EXISTS idx_%suser_user_name
ON %suser (user_name)
`, pfx, pfx))
	if err != nil { return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE INDEX IF NOT EXISTS idx_%suser_authkey_user_name
ON %suser_authkey (user_name);
`, pfx, pfx))
	if err != nil { return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE INDEX IF NOT EXISTS idx_%suser_signkey_user_name
ON %suser_signkey (user_name);
`, pfx, pfx))
	if err != nil { return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE INDEX IF NOT EXISTS idx_%snamespace_ns_name
ON %snamespace (ns_name);
`, pfx, pfx))
	if err != nil { return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %sissue (
    repo_namespace TEXT,
	repo_name TEXT,
	issue_id INTEGER,
	issue_timestamp INTEGER,
	issue_author TEXT,
	issue_title TEXT,
	issue_content TEXT,
	-- 1 - opened.  2 - close as solved.  3 - close as discarded.
	issue_status INTEGER,
    issue_priority SMALLINT,
	FOREIGN KEY (repo_namespace, repo_name)
      REFERENCES %srepository(repo_namespace, repo_name),
	FOREIGN KEY (issue_author)
	  REFERENCES %suser(user_name)
);
`, pfx, pfx, pfx))
	if err != nil { return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %sissue_event (
    issue_abs_id INTEGER,
	-- 1 - comment.  2 - close as solved.  3 - close as discarded.
	-- 4 - reopened.
	issue_event_type INTEGER,
	issue_event_time INTEGER,
	issue_event_author TEXT,
	issue_event_content TEXT,
    FOREIGN KEY (issue_abs_id) REFERENCES %sissue(rowid)
);
`, pfx, pfx))
	if err != nil { return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %spull_request (
    username TEXT,
    pull_request_id INT,
    title TEXT,
    -- the receiving repo
    receiver_namespace TEXT,
    receiver_name TEXT,
    receiver_branch TEXT,
    -- the repo you're pulling from
    provider_namespace TEXT,
    provider_name TEXT,
    provider_branch TEXT,
    -- in json.
    merge_conflict_check_result TEXT,
    merge_conflict_check_timestamp INT,
    pull_request_status INT,
    pull_request_timestamp INT,
	FOREIGN KEY (username) REFERENCES %suser(user_name)
  )
`, pfx, pfx))
	if err != nil { return err }
	
	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %spull_request_event (
    pull_request_abs_id INTEGER,
	-- 1 - normal comment.
	-- 2 - comment on code.
	-- 3 - update on provider branch.
	-- 4 - merge conflict check.
	-- 5 - close as not merged.
	-- 6 - close (merged).
    -- 7 - reopen
	event_type INTEGER,
	event_timestamp INTEGER,
	event_author TEXT,
	event_content TEXT,
	FOREIGN KEY (pull_request_abs_id) REFERENCES %spull_request(rowid)
  )
`, pfx, pfx))
	if err != nil { return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE INDEX IF NOT EXISTS idx_%spull_request_event_pull_request_abs_id
ON %spull_request_event (pull_request_abs_id);
`, pfx, pfx))
	if err != nil { return err }

	_, err = tx.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %suser_email (
    username TEXT,
	email TEXT,
    verified INTEGER,
	FOREIGN KEY (username) REFERENCES %suser(user_name)
)`, pfx, pfx))
	
	tx.Commit()
	return nil
}

