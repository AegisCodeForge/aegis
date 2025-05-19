package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/bctnry/aegis/pkg/aegis/db"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/gitlib"
	_ "github.com/mattn/go-sqlite3"
)

func (dbif *SqliteAegisDatabaseInterface) GetUserByName(name string) (*model.AegisUser, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT user_name, user_title, user_email, user_bio, user_website, user_status, user_password_hash
FROM %suser
WHERE user_name = ?
`, pfx))
	if err != nil { return nil, err }
	var username, title, email, bio, website, ph string
	var status int
	err = stmt.QueryRow(name).Scan(&username, &title, &email, &bio, &website, &status, &ph)
	if err == sql.ErrNoRows { return nil, db.NewAegisDatabaseError(db.ENTITY_NOT_FOUND, "") }
	if err != nil { return nil, err }
	return &model.AegisUser{
		Name: username,
		Title: title,
		Email: email,
		Bio: bio,
		Website: website,
		Status: model.AegisUserStatus(status),
		PasswordHash: ph,
	}, nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAuthKeyByName(userName string, keyName string) (*model.AegisAuthKey, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT key_text
FROM %suser_authkey
WHERE user_name = ? AND key_name = ?
`, pfx))
	if err != nil { return nil, err }
	r := stmt.QueryRow(userName, keyName)
	err = r.Err()
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, db.NewAegisDatabaseError(db.ENTITY_NOT_FOUND, "Cannot find key")
		}
		return nil, err
	}
	kt := ""
	err = r.Scan(&kt)
	if err != nil { return nil, err }
	return &model.AegisAuthKey{
		UserName: userName,
		KeyName: keyName,
		KeyText: kt,
	}, nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllAuthKeyByUsername(name string) ([]model.AegisAuthKey, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT key_name, key_text
FROM %suser_authkey
WHERE user_name = ?
`, pfx))
	if err != nil { return nil, err }
	r, err := stmt.Query(name)
	if err != nil { return nil, err }
	defer r.Close()
	res := make([]model.AegisAuthKey, 0)
	for r.Next() {
		var keyName, keyText string
		err = r.Scan(&keyName, &keyText)
		if err != nil { return nil, err }
		res = append(res, model.AegisAuthKey{
			UserName: name,
			KeyText: keyText,
			KeyName: keyName,
		})
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) RegisterAuthKey(username string, keyname string, keytext string) error {
	pfx := dbif.config.DatabaseTablePrefix
	stmt1, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT 1 FROM %suser_authkey WHERE user_name = ? AND key_name = ?
`, pfx))
	if err != nil { return err }
	r := stmt1.QueryRow(username, keyname)
	if r.Err() != nil { return r.Err() }
	var verdict string
	err = r.Scan(&verdict)
	if err != nil && err != sql.ErrNoRows { return err }
	if err == nil {
		return db.NewAegisDatabaseError(db.ENTITY_ALREADY_EXISTS, "Key already exists with the same name")
	}
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	stmt2, err := tx.Prepare(fmt.Sprintf(`
INSERT INTO %suser_authkey(user_name, key_name, key_text)
VALUES (?,?,?)
`, pfx))
	if err != nil { return err }
	_, err = stmt2.Exec(username, keyname, keytext)
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) RemoveAuthKey(username string, keyname string) error {
	pfx := dbif.config.DatabaseTablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare(fmt.Sprintf(`
DELETE FROM %suser_authkey
WHERE user_name = ? AND key_name = ?
`, pfx))
	if err != nil { return err }
	_, err = stmt.Exec(username, keyname)
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllSignKeyByUsername(name string) ([]model.AegisSigningKey, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT key_name, key_text
FROM %suser_signkey
WHERE user_name = ?
`, pfx))
	if err != nil { return nil, err }
	r, err := stmt.Query(name)
	if err != nil { return nil, err }
	defer r.Close()
	res := make([]model.AegisSigningKey, 0)
	for r.Next() {
		var keyName, keyText string
		err = r.Scan(&keyName, &keyText)
		if err != nil { return nil, err }
		res = append(res, model.AegisSigningKey{
			UserName: name,
			KeyText: keyText,
			KeyName: keyName,
		})
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) RegisterSignKey(username string, keyname string, keytext string) error {
	pfx := dbif.config.DatabaseTablePrefix
	stmt1, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT 1 FROM %suser_sign WHERE user_name = ? AND key_name = ?
`, pfx))
	if err != nil { return err }
	r := stmt1.QueryRow(username, keyname)
	if r.Err() != nil { return r.Err() }
	var verdict string
	err = r.Scan(&verdict)
	if err != nil && err != sql.ErrNoRows { return err }
	if err == nil {
		return db.NewAegisDatabaseError(db.ENTITY_ALREADY_EXISTS, "Key already exists with the same name")
	}
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	stmt2, err := tx.Prepare(fmt.Sprintf(`
INSERT INTO %suser_signkey(user_name, key_name, key_text)
VALUES (?,?,?)
`, pfx))
	if err != nil { return err }
	_, err = stmt2.Exec(username, keyname, keytext)
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) RemoveSignKey(username string, keyname string) error {
	pfx := dbif.config.DatabaseTablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare(fmt.Sprintf(`
DELETE FROM %suser_signkey
WHERE user_name = ? AND key_name = ?
`, pfx))
	if err != nil { return err }
	_, err = stmt.Exec(username, keyname)
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) RegisterUser(name string, email string, passwordHash string, status model.AegisUserStatus) (*model.AegisUser, error) {
	pfx := dbif.config.DatabaseTablePrefix
	t := time.Now().Unix()
	tx, err := dbif.connection.Begin()
	if err != nil { return nil, err }
	stmt, err := tx.Prepare(fmt.Sprintf(`
INSERT INTO %suser(user_name, user_title, user_email, user_bio, user_website, user_password_hash, user_reg_datetime, user_status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`, pfx))
	if err != nil { return nil, err }
	_, err = stmt.Exec(name, name, email, "", "", passwordHash, t, status)
	if err != nil { tx.Rollback(); return nil, err }
	err = tx.Commit()
	if err != nil { tx.Rollback(); return nil, err }
	// we delete whatever we have inside the old ns.
	// this should be ok since when a user exists as "deleted"
	// state the actions above should've violated unique constraint
	// and triggered an error already. of course we should remove
	// "deleted" user regularly from the dbto prevent possible sabotage.
	userNsPath := path.Join(dbif.config.GitRoot, name)
	err = os.RemoveAll(userNsPath)
	if err != nil && !os.IsNotExist(err) { tx.Rollback(); return nil, err }
	err = os.MkdirAll(userNsPath, os.ModeDir|0755)
	// TODO: chown.
	if err != nil { tx.Rollback(); return nil, err }
	return &model.AegisUser{
		Name: name,
		Title: name,
		Email: email,
		PasswordHash: passwordHash,
		Bio: "",
		Website: "",
		RegisterTime: t,
	}, nil
}

func (dbif *SqliteAegisDatabaseInterface) UpdateUserInfo(name string, uobj *model.AegisUser) error {
	pfx := dbif.config.DatabaseTablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare(fmt.Sprintf(`
UPDATE %suser
SET
    user_title = ?, user_email = ?, user_bio = ?,
    user_website = ?, user_status = ?
WHERE
    user_name = ?
`, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = stmt.Exec(uobj.Title, uobj.Email, uobj.Bio, uobj.Website, uobj.Status, name)
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit()
	if err != nil { tx.Rollback(); return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) UpdateUserPassword(name string, newPasswordHash string) error {
	pfx := dbif.config.DatabaseTablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare(fmt.Sprintf(`
UPDATE %suser
SET user_password_hash = ?
WHERE user_name = ?
`, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = stmt.Exec(newPasswordHash, name)
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit()
	if err != nil { tx.Rollback(); return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) UpdateUserStatus(name string, newStatus model.AegisUserStatus) error {
	pfx := dbif.config.DatabaseTablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare(fmt.Sprintf(`
UPDATE %suser
SET user_status = ?
WHERE user_name = ?
`, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = stmt.Exec(newStatus, name)
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit()
	if err != nil { tx.Rollback(); return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) HardDeleteUserByName(name string) error {
	pfx := dbif.config.DatabaseTablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare(fmt.Sprintf(`
DELETE FROM %suser WHERE user_name = ?
`, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = stmt.Exec(name)
	if err != nil { tx.Rollback(); return err }
	userNsPath := path.Join(dbif.config.GitRoot, name)
	err = os.RemoveAll(userNsPath)
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit()
	if err != nil { tx.Rollback(); return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) GetNamespaceByName(name string) (*model.Namespace, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_status, ns_acl
FROM %snamespace
WHERE ns_name = ?
`, pfx))
	if err != nil { return nil, err }
	r := stmt.QueryRow(name)
	if r.Err() != nil { return nil, r.Err() }
	var title, desc, email, owner, acl string
	var reg_date int64
	var status int
	err = r.Scan(&title, &desc, &email, &owner, &reg_date, &status, &acl)
	if err == sql.ErrNoRows {
		return nil, db.NewAegisDatabaseError(
			db.ENTITY_NOT_FOUND,
			fmt.Sprintf("Could not find namespace %s", name),
		)
	}
	if err != nil { return nil, err }
	a, err := model.ParseACL(acl)
	if err != nil { return nil, err }
	return &model.Namespace{
		Name: name,
		Title: title,
		Description: desc,
		Email: email,
		Owner: owner,
		RegisterTime: reg_date,
		ACL: a,
		Status: model.AegisNamespaceStatus(status),
	}, nil
}

func (dbif *SqliteAegisDatabaseInterface) GetRepositoryByName(nsName string, repoName string) (*model.Repository, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT repo_description, repo_owner, repo_acl, repo_status
FROM %srepository
WHERE repo_namespace = ? AND repo_name = ?
`, pfx))
	if err != nil { return nil, err }
	r := stmt.QueryRow(nsName, repoName)
	if r.Err() != nil {
		if r.Err() == sql.ErrNoRows {
			return nil, db.NewAegisDatabaseError(db.ENTITY_NOT_FOUND, fmt.Sprintf("Repository %s not found in %s", repoName, nsName))
		}
		return nil, r.Err()
	}
	var desc, owner, acl string
	var status int
	r = stmt.QueryRow(nsName, repoName)
	if r.Err() != nil { return nil, r.Err() }
	err = r.Scan(&desc, &owner, &acl, &status)
	if err != nil { return nil, err }
	p := path.Join(dbif.config.GitRoot, nsName, repoName)
	res, err := model.NewRepository(nsName, repoName, gitlib.NewLocalGitRepository(nsName, repoName, p))
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) RegisterNamespace(name string, ownerUsername string) (*model.Namespace, error) {
	pfx := dbif.config.DatabaseTablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return nil, err }
	stmt, err := tx.Prepare(fmt.Sprintf(`
INSERT INTO %snamespace(ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_status, ns_acl)
VALUES (?,?,?,?,?,?,?, ?)
`, pfx))
	if err != nil { tx.Rollback(); return nil, err }
	_, err = stmt.Exec(name, name, "", "", ownerUsername, time.Now().Unix(), model.NAMESPACE_NORMAL_PUBLIC, "")
	if err != nil { tx.Rollback(); return nil, err }
	nsPath := path.Join(dbif.config.GitRoot, name)
	err = os.RemoveAll(nsPath)
	if err != nil { tx.Rollback(); return nil, err }
	err = os.Mkdir(nsPath, os.ModeDir|0755)
	if err != nil { tx.Rollback(); return nil, err }
	// TODO: chown.
	err = tx.Commit()
	if err != nil { tx.Rollback(); return nil, err }
	return &model.Namespace{
		Name: name,
		Title: name,
		Description: "",
		Status: model.NAMESPACE_NORMAL_PUBLIC,
	}, nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllNamespace() (map[string]*model.Namespace, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_status, ns_acl
FROM %snamespace
WHERE ns_status != 3 AND ns_status != 4
`, pfx))
	if err != nil { return nil, err }
	rs, err := stmt.Query()
	if err != nil { return nil, err }
	defer rs.Close()
	res := make(map[string]*model.Namespace, 0)
	for rs.Next() {
		var name, title, desc, email, owner, acl string
		var regtime int64
		var status int64
		err = rs.Scan(&name, &title, &desc, &email, &owner, &regtime, &status, &acl)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res[name] = &model.Namespace{
			Name: name,
			Title: title,
			Description: desc,
			Email: email,
			Owner: owner,
			RegisterTime: regtime,
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		}
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllNamespaceByOwner(name string) (map[string]*model.Namespace, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_status, ns_acl
FROM %snamespace
WHERE ns_status != 3 AND ns_status != 4 AND ns_owner = ?
`, pfx))
	if err != nil { return nil, err }
	rs, err := stmt.Query(name)
	if err != nil { return nil, err }
	defer rs.Close()
	res := make(map[string]*model.Namespace, 0)
	for rs.Next() {
		var name, title, desc, email, owner, acl string
		var regtime int64
		var status int64
		err = rs.Scan(&name, &title, &desc, &email, &owner, &regtime, &status, &acl)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res[name] = &model.Namespace{
			Name: name,
			Title: title,
			Description: desc,
			Email: email,
			Owner: owner,
			RegisterTime: regtime,
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		}
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllRepositoryFromNamespace(ns string) (map[string]*model.Repository, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT repo_name, repo_description, repo_acl, repo_status
FROM %srepository
WHERE repo_namespace = ?
`, pfx))
	if err != nil { return nil, err }
	rs, err := stmt.Query(ns)
	if err != nil { return nil, err }
	defer rs.Close()
	res := make(map[string]*model.Repository, 0)
	for rs.Next() {
		var name, desc, acl string
		var status int64
		err = rs.Scan(&name, &desc, &acl, &status)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, ns, name)
		res[name] = &model.Repository{
			Namespace: ns,
			Name: name,
			Description: desc,
			AccessControlList: a,
			Status: model.AegisRepositoryStatus(status),
			Repository: gitlib.NewLocalGitRepository(ns, name, p),
		}
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) UpdateNamespaceInfo(name string, nsobj *model.Namespace) error {
	pfx := dbif.config.DatabaseTablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare(fmt.Sprintf(`
UPDATE %snamespace
SET ns_title = ?, ns_description = ?, ns_email = ?, ns_owner = ?
WHERE ns_name = ?
`, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = stmt.Exec(nsobj.Title, nsobj.Description, nsobj.Email, nsobj.Owner, name)
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit()
	if err != nil { tx.Rollback(); return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) UpdateNamespaceOwner(name string, newOwner string) error {
	pfx := dbif.config.DatabaseTablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare(fmt.Sprintf(`
UPDATE %snamespace
SET ns_owner = ?
WHERE ns_name = ?
`, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = stmt.Exec(newOwner, name)
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit()
	if err != nil { tx.Rollback(); return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) UpdateNamespaceStatus(name string, newStatus model.AegisNamespaceStatus) error {
	pfx := dbif.config.DatabaseTablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare(fmt.Sprintf(`
UPDATE %snamespace
SET ns_status = ?
WHERE ns_name = ?
`, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = stmt.Exec(newStatus, name)
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit()
	if err != nil { tx.Rollback(); return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) HardDeleteNamespaceByName(name string) error {
	pfx := dbif.config.DatabaseTablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare(fmt.Sprintf(`
DELETE FROM %snamespace WHERE ns_name = ?
`, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = stmt.Exec(name)
	if err != nil { tx.Rollback(); return err }
	nsPath := path.Join(dbif.config.GitRoot, name)
	err = os.RemoveAll(nsPath)
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit()
	if err != nil { tx.Rollback(); return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) CreateRepository(ns string, name string) (*model.Repository, error) {
	pfx := dbif.config.DatabaseTablePrefix
	fullName := ns + ":" + name
	tx, err := dbif.connection.Begin()
	if err != nil { return nil, err }
	defer tx.Rollback()
	stmt1, err := tx.Prepare(fmt.Sprintf(`
INSERT INTO %srepository(repo_fullname, repo_namespace, repo_name, repo_description, repo_acl, repo_status, repo_owner)
VALUES (?,?,?,?,?,?,?)
`, pfx))
	if err != nil { return nil, err }
	_, err = stmt1.Exec(fullName, ns, name, new(string), new(string), model.REPO_NORMAL_PUBLIC, new(string))
	if err != nil { return nil, err }
	p := path.Join(dbif.config.GitRoot, ns, name)
	err = os.RemoveAll(p)
	if err != nil { return nil, err }
	if err = os.MkdirAll(p, os.ModeDir|0755); err != nil {
		return nil, err
	}
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = p
	if err = cmd.Run(); err != nil { return nil, err }
	if err = tx.Commit(); err != nil {
		fmt.Println("what?", err)
		return nil, err
	}
	r, err := model.NewRepository(ns, name, gitlib.NewLocalGitRepository(ns, name, p))
	if err != nil { return nil, err }
	return r, nil
}

func (dbif *SqliteAegisDatabaseInterface) UpdateRepositoryInfo(ns string, name string, robj *model.Repository) error {
	pfx := dbif.config.DatabaseTablePrefix
	stmt1, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT rowid FROM %srepository WHERE repo_namespace = ? AND repo_name = ?
`, pfx))
	if err != nil { return err }
	v := stmt1.QueryRow(ns, name)
	if v.Err() != nil { return v.Err() }
	var rowid string
	err = v.Scan(&rowid)
	if err != nil { return err }
	if len(rowid) <= 0 { return db.NewAegisDatabaseError(db.ENTITY_NOT_FOUND, fmt.Sprintf("%s not found in %s", name, ns)) }
	tx, err := dbif.connection.Begin()
	if err != nil { tx.Rollback(); return err }
	stmt2, err := tx.Prepare(fmt.Sprintf(`
UPDATE %srepository
SET repo_description = ?, repo_owner = ?
WHERE rowid = ?
`, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = stmt2.Exec(robj.Description, robj.Owner, rowid)
	if err != nil { tx.Rollback(); return err }
	// TODO: maybe we should change the description stored as file in git dir
	// as well...
	err = tx.Commit()
	if err != nil { tx.Rollback(); return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) UpdateRepositoryStatus(ns string, name string, newStatus model.AegisRepositoryStatus) error {
	pfx := dbif.config.DatabaseTablePrefix
	stmt1, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT rowid FROM %srepository WHERE repo_namespace = ? AND repo_name = ?
`, pfx))
	if err != nil { return err }
	v := stmt1.QueryRow(ns, name)
	if v.Err() != nil { return v.Err() }
	var rowid string
	err = v.Scan(&rowid)
	if err != nil { return err }
	if len(rowid) <= 0 { return db.NewAegisDatabaseError(db.ENTITY_NOT_FOUND, fmt.Sprintf("%s not found in %s", name, ns)) }
	tx, err := dbif.connection.Begin()
	if err != nil { tx.Rollback(); return err }
	stmt2, err := tx.Prepare(fmt.Sprintf(`
UPDATE %srepository
SET repo_status = ?
WHERE rowid = ?
`, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = stmt2.Exec(newStatus, rowid)
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit()
	if err != nil { tx.Rollback(); return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) MoveRepository(oldNs string, oldName string, newNs string, newName string) error {
	// we first check if the new name is already taken
	pfx := dbif.config.DatabaseTablePrefix
	stmt1, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT 1 FROM %srepository
WHERE repo_namespace = ? AND repo_name = ?
`, pfx))
	if err != nil { return err }
	v := stmt1.QueryRow(newNs, newName)
	if v.Err() != nil { return v.Err() }
	var s string
	v.Scan(&s)
	if len(s) > 0 { return db.NewAegisDatabaseError(db.ENTITY_ALREADY_EXISTS, fmt.Sprintf("%s:%s already exists", newNs, newName)) }
	// this is sqlite thus we should be able to use rowid.
	// for other db engine we would need a PRIMARY KEY INT AUTO_INCREMENT.
	stmt2, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT rowid FROM %srepository
WHERE repo_namespace = ? AND repo_name = ?
`, pfx))
	if err != nil { return err }
	v = stmt2.QueryRow(oldNs, oldName)
	if v.Err() != nil { return v.Err() }
	v.Scan(&s)
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	stmt3, err := tx.Prepare(fmt.Sprintf(`
UPDATE %srepository
SET repo_namespace = ?, repo_name = ?
WHERE rowid = ?
`, pfx))
	if err != nil { return err }
	_, err = stmt3.Exec(newNs, newName, s)
	if err != nil { return err }
	oldPath := path.Join(dbif.config.GitRoot, oldNs, oldName)
	newPath := path.Join(dbif.config.GitRoot, newNs, newName)
	err = os.Rename(oldPath, newPath)
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit()
	if err != nil { tx.Rollback(); return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) HardDeleteRepository(ns string, name string) error {
	pfx := dbif.config.DatabaseTablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	stmt, err := tx.Prepare(fmt.Sprintf(`
DELETE FROM %srepository
WHERE repo_namespace = ? AND repo_name = ?
`, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = stmt.Exec(ns, name)
	if err != nil { tx.Rollback(); return err }
	p := path.Join(dbif.config.GitRoot, ns, name)
	err = os.RemoveAll(p)
	if err != nil { tx.Rollback(); return err }
	err = tx.Commit()
	if err != nil { tx.Rollback(); return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllUsers(pageNum int, pageSize int) ([]*model.AegisUser, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT user_name, user_title, user_email, user_bio, user_website, user_status, user_password_hash
FROM %suser
ORDER BY rowid ASC LIMIT ? OFFSET ?
`, pfx))
	if err != nil { return nil, err }
	defer stmt.Close()
	r, err := stmt.Query(pageSize, pageNum * pageSize)
	if err != nil { return nil, err }
	defer r.Close()
	res := make([]*model.AegisUser, 0)
	var username, title, email, bio, website, ph string
	var status int
	for r.Next() {
		err = r.Scan(&username, &title, &email, &bio, &website, &status, &ph)
		if err != nil { return nil, err }
		res = append(res, &model.AegisUser{
			Name: username,
			Title: title,
			Email: email,
			Bio: bio,
			Website: website,
			Status: model.AegisUserStatus(status),
			PasswordHash: ph,
		})
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllNamespaces(pageNum int, pageSize int) ([]*model.Namespace, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_status, ns_acl
FROM %snamespace
ORDER BY rowid ASC LIMIT ? OFFSET ?
`, pfx))
	if err != nil { return nil, err }
	defer stmt.Close()
	r, err := stmt.Query(pageSize, pageNum * pageSize)
	if err != nil { return nil, err }
	defer r.Close()
	res := make([]*model.Namespace, 0)
	var name, title, desc, email, owner, acl string
	var reg_date int64
	var status int
	for r.Next() {
		err = r.Scan(&name, &title, &desc, &email, &owner, &reg_date, &status, &acl)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res = append(res, &model.Namespace{
			Name: name,
			Title: title,
			Description: desc,
			Email: email,
			Owner: owner,
			RegisterTime: reg_date,
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		})
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllRepositories(pageNum int, pageSize int) ([]*model.Repository, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT repo_namespace, repo_name, repo_description, repo_acl, repo_status
FROM %srepository
ORDER BY rowid ASC LIMIT ? OFFSET ?
`, pfx))
	if err != nil { return nil, err }
	defer stmt.Close()
	r, err := stmt.Query(pageSize, pageNum * pageSize)
	if err != nil { return nil, err }
	defer r.Close()
	res := make([]*model.Repository, 0)
	var ns, name, title, desc, acl, owner string
	var status int
	for r.Next() {
		err = r.Scan(&ns, &name, &title, &desc, &acl, &owner, &status)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, ns, name)
		res = append(res, &model.Repository{
			Namespace: ns,
			Name: name,
			Description: desc,
			AccessControlList: a,
			Status: model.AegisRepositoryStatus(status),
			Repository: gitlib.NewLocalGitRepository(ns, name, p),
		})
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) CountAllUser() (int64, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(
		fmt.Sprintf(`SELECT COUNT(*) FROM %suser`, pfx),
	)
	if err != nil { return 0, err }
	defer stmt.Close()
	r := stmt.QueryRow()
	if r.Err() != nil { return 0, r.Err() }
	var res int64
	err = r.Scan(&res)
	if err != nil { return 0, r.Err() }
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) CountAllNamespace() (int64, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(
		fmt.Sprintf(`SELECT COUNT(*) FROM %snamespace`, pfx),
	)
	if err != nil { return 0, err }
	defer stmt.Close()
	r := stmt.QueryRow()
	if r.Err() != nil { return 0, r.Err() }
	var res int64
	err = r.Scan(&res)
	if err != nil { return 0, r.Err() }
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) CountAllRepositories() (int64, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(
		fmt.Sprintf(`SELECT COUNT(*) FROM %srepository`, pfx),
	)
	if err != nil { return 0, err }
	defer stmt.Close()
	r := stmt.QueryRow()
	if r.Err() != nil { return 0, r.Err() }
	var res int64
	err = r.Scan(&res)
	if err != nil { return 0, r.Err() }
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) SearchForUser(k string, pageNum int, pageSize int) ([]*model.AegisUser, error) {
	pfx := dbif.config.DatabaseTablePrefix
	pattern := strings.ReplaceAll(k, "\\", "\\\\")
	pattern = strings.ReplaceAll(pattern, "%", "\\%")
	pattern = strings.ReplaceAll(pattern, "_", "\\_")
	pattern = "%" + pattern + "%"
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT user_name, user_title, user_email, user_bio, user_website, user_status, user_password_hash
FROM %suser
WHERE user_name LIKE ? ESCAPE ? OR user_title LIKE ? ESCAPE ?
ORDER BY rowid ASC LIMIT ? OFFSET ?
`, pfx))
	if err != nil { return nil, err }
	defer stmt.Close()
	r, err := stmt.Query(pattern, "\\", pattern, "\\", pageSize, pageNum * pageSize)
	if err != nil { return nil, err }
	defer r.Close()
	res := make([]*model.AegisUser, 0)
	var username, title, email, bio, website, ph string
	var status int
	for r.Next() {
		err = r.Scan(&username, &title, &email, &bio, &website, &status, &ph)
		if err != nil { return nil, err }
		res = append(res, &model.AegisUser{
			Name: username,
			Title: title,
			Email: email,
			Bio: bio,
			Website: website,
			Status: model.AegisUserStatus(status),
			PasswordHash: ph,
		})
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) SearchForNamespace(k string, pageNum int, pageSize int) ([]*model.Namespace, error) {
	pfx := dbif.config.DatabaseTablePrefix
	pattern := strings.ReplaceAll(k, "\\", "\\\\")
	pattern = strings.ReplaceAll(pattern, "%", "\\%")
	pattern = strings.ReplaceAll(pattern, "_", "\\_")
	pattern = "%" + pattern + "%"
	fmt.Println("pattern", pattern)
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_status, ns_acl
FROM %snamespace
WHERE ns_name LIKE ? ESCAPE ? OR ns_title LIKE ? ESCAPE ?
ORDER BY rowid ASC LIMIT ? OFFSET ?
`, pfx))
	if err != nil { return nil, err }
	defer stmt.Close()
	r, err := stmt.Query(pattern, "\\", pattern, "\\", pageSize, pageNum * pageSize)
	if err != nil { return nil, err }
	defer r.Close()
	res := make([]*model.Namespace, 0)
	var name, title, desc, email, owner, acl string
	var reg_date int64
	var status int
	for r.Next() {
		err = r.Scan(&name, &title, &desc, &email, &owner, &reg_date, &status, &acl)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res = append(res, &model.Namespace{
			Name: name,
			Title: title,
			Description: desc,
			Email: email,
			Owner: owner,
			RegisterTime: reg_date,
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		})
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) SearchForRepository(k string, pageNum int, pageSize int) ([]*model.Repository, error) {
	pfx := dbif.config.DatabaseTablePrefix
	pattern := strings.ReplaceAll(k, "\\", "\\\\")
	pattern = strings.ReplaceAll(pattern, "%", "\\%")
	pattern = strings.ReplaceAll(pattern, "_", "\\_")
	pattern = "%" + pattern + "%"
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT repo_namespace, repo_name, repo_description, repo_acl, repo_status, repo_acl
FROM %srepository
WHERE repo_namespace LIKE ? ESCAPE ? OR repo_name LIKE ? ESCAPE ?
ORDER BY rowid ASC LIMIT ? OFFSET ?
`, pfx))
	if err != nil { return nil, err }
	defer stmt.Close()
	r, err := stmt.Query(pattern, "\\", pattern, "\\", pageSize, pageNum * pageSize)
	if err != nil { return nil, err }
	defer r.Close()
	res := make([]*model.Repository, 0)
	var ns, name, title, desc, acl, owner string
	var status int
	for r.Next() {
		err = r.Scan(&ns, &name, &title, &desc, &acl, &owner, &status, &acl)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, ns, name)
		res = append(res, &model.Repository{
			Namespace: ns,
			Name: name,
			Description: desc,
			AccessControlList: a,
			Status: model.AegisRepositoryStatus(status),
			Repository: gitlib.NewLocalGitRepository(ns, name, p),
		})
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) SetNamespaceACL(actionUserName string, nsName string, targetUserName string, aclt *model.ACLTuple) error {
	pfx := dbif.config.DatabaseTablePrefix
	stmt1, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT ns_owner, ns_acl FROM %snamespace WHERE ns_name = ?
`, pfx))
	if err != nil { return err }
	r := stmt1.QueryRow(nsName)
	if r.Err() != nil { stmt1.Close(); return r.Err() }
	var nsOwner, aclStr string
	err = r.Scan(&nsOwner, &aclStr)
	if err != nil { stmt1.Close(); return err }
	acl, err := model.ParseACL(aclStr)
	if err != nil { return err }
	v, ok := acl.ACL[actionUserName]
	if !ok && nsOwner != actionUserName { return db.ErrNotEnoughPermission }
	v2, ok := acl.ACL[targetUserName]
	if !ok {  // requires addMember
		if aclt == nil {
			// if the caller attemps to delete the target user but the
			// target user isn't even in the member list in the first
			// place, we don't report error.
			return nil
		}
		if !v.AddMember { return db.ErrNotEnoughPermission }
		if !v.DeleteMember && aclt.DeleteMember { return db.ErrNotEnoughPermission }
		if !v.EditMember && aclt.EditMember { return db.ErrNotEnoughPermission }
		if !v.AddRepository && aclt.AddRepository { return db.ErrNotEnoughPermission }
		if !v.EditRepository && aclt.EditRepository { return db.ErrNotEnoughPermission }
		if !v.PushToRepository && aclt.PushToRepository { return db.ErrNotEnoughPermission }
		if !v.ArchiveRepository && aclt.ArchiveRepository { return db.ErrNotEnoughPermission }
		if !v.DeleteRepository && aclt.DeleteRepository { return db.ErrNotEnoughPermission }
	} else {
		if aclt == nil {
			// calls for deleting target user - requires deleteMember
			if !v.DeleteMember { return db.ErrNotEnoughPermission }
		} else {
			// calls for editing target user - requires editMember.
			c1 := v2.AddMember != aclt.AddMember
			c2 := v2.DeleteMember != aclt.DeleteMember
			c3 := v2.EditMember != aclt.EditMember
			c4 := v2.AddRepository != aclt.AddRepository
			c5 := v2.EditRepository != aclt.EditRepository
			c6 := v2.PushToRepository != aclt.PushToRepository
			c7 := v2.ArchiveRepository != aclt.ArchiveRepository
			c8 := v2.DeleteRepository != aclt.DeleteRepository
			if c1 || c2 || c3 || c4 || c5 || c6 || c7 || c8 {
				if !v.EditMember { return db.ErrNotEnoughPermission }
			}
		}
	}
	acl.ACL[targetUserName] = aclt
	acls, err := acl.SerializeACL()
	if err != nil { return err }
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	defer tx.Rollback()
	stmt2, err := tx.Prepare(fmt.Sprintf(`
UPDATE %snamespace SET ns_acl = ? WHERE ns_name = ?
`, pfx))
	if err != nil { return err }
	_, err = stmt2.Exec(acls, nsName)
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) SetRepositoryACL(actionUserName string, nsName string, repoName string, targetUserName string, aclt *model.ACLTuple) error {
	pfx := dbif.config.DatabaseTablePrefix
	stmt1, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT repo_owner, repo_acl FROM %srepository WHERE repo_name = ? AND repo_namespace = ?
`, pfx))
	if err != nil { return err }
	r := stmt1.QueryRow(repoName, nsName)
	if r.Err() != nil { stmt1.Close(); return r.Err() }
	var owner, aclStr string
	err = r.Scan(&owner, &aclStr)
	if err != nil { stmt1.Close(); return err }
	acl, err := model.ParseACL(aclStr)
	if err != nil { return err }
	v, ok := acl.ACL[actionUserName]
	if !ok && owner != actionUserName { return db.ErrNotEnoughPermission }
	v2, ok := acl.ACL[targetUserName]
	if !ok {  // requires addMember
		if aclt == nil {
			// if the caller attemps to delete the target user but the
			// target user isn't even in the member list in the first
			// place, we don't report error.
			return nil
		}
		if !v.AddMember { return db.ErrNotEnoughPermission }
		if !v.DeleteMember && aclt.DeleteMember { return db.ErrNotEnoughPermission }
		if !v.EditMember && aclt.EditMember { return db.ErrNotEnoughPermission }
		if !v.AddRepository && aclt.AddRepository { return db.ErrNotEnoughPermission }
		if !v.EditRepository && aclt.EditRepository { return db.ErrNotEnoughPermission }
		if !v.PushToRepository && aclt.PushToRepository { return db.ErrNotEnoughPermission }
		if !v.ArchiveRepository && aclt.ArchiveRepository { return db.ErrNotEnoughPermission }
		if !v.DeleteRepository && aclt.DeleteRepository { return db.ErrNotEnoughPermission }
	} else {
		if aclt == nil {
			// calls for deleting target user - requires deleteMember
			if !v.DeleteMember { return db.ErrNotEnoughPermission }
		} else {
			// calls for editing target user - requires editMember.
			c1 := v2.AddMember != aclt.AddMember
			c2 := v2.DeleteMember != aclt.DeleteMember
			c3 := v2.EditMember != aclt.EditMember
			c4 := v2.AddRepository != aclt.AddRepository
			c5 := v2.EditRepository != aclt.EditRepository
			c6 := v2.PushToRepository != aclt.PushToRepository
			c7 := v2.ArchiveRepository != aclt.ArchiveRepository
			c8 := v2.DeleteRepository != aclt.DeleteRepository
			if c1 || c2 || c3 || c4 || c5 || c6 || c7 || c8 {
				if !v.EditMember { return db.ErrNotEnoughPermission }
			}
		}
	}
	acl.ACL[targetUserName] = aclt
	acls, err := acl.SerializeACL()
	if err != nil { return err }
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	defer tx.Rollback()
	stmt2, err := tx.Prepare(fmt.Sprintf(`
UPDATE %srepository SET repo_acl = ? WHERE repo_name = ? AND repo_namespace = ?
`, pfx))
	if err != nil { return err }
	_, err = stmt2.Exec(acls, repoName, nsName)
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	return nil
}



