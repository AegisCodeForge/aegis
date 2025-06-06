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
	if r.Err() != nil { return nil, r.Err() }
	var desc, owner, acl string
	var status int
	err = r.Scan(&desc, &owner, &acl, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, db.NewAegisDatabaseError(db.ENTITY_NOT_FOUND, fmt.Sprintf("Repository %s not found in %s", repoName, nsName))
		}
		return nil, err
	}
	p := path.Join(dbif.config.GitRoot, nsName, repoName)
	res, err := model.NewRepository(nsName, repoName, gitlib.NewLocalGitRepository(nsName, repoName, p))
	res.Owner = owner
	res.Status = model.AegisRepositoryStatus(status)
	aclobj, err := model.ParseACL(acl)
	if err != nil { return nil, err }
	res.AccessControlList = aclobj
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

func (dbif *SqliteAegisDatabaseInterface) GetAllVisibleNamespace(username string) (map[string]*model.Namespace, error) {
	pfx := dbif.config.DatabaseTablePrefix
	privateSelectClause := ""
	if len(username) > 0 {
		privateSelectClause = fmt.Sprintf("OR (ns_owner = ? OR ns_acl LIKE ? ESCAPE ?)", )
	}
	stmt1, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_status, ns_acl
FROM %snamespace
WHERE ns_status = 1 %s
`, pfx, privateSelectClause))
	if err != nil { return nil, err }
	defer stmt1.Close()
	var rs *sql.Rows
	if len(username) > 0 {
		pattern := strings.ReplaceAll(username, "\\", "\\\\")
		pattern = strings.ReplaceAll(pattern, "%", "\\%")
		pattern = strings.ReplaceAll(pattern, "_", "\\_")
		pattern = "%" + pattern + "%"
		rs, err = stmt1.Query(username, pattern, "\\")
	} else {
		rs, err = stmt1.Query()
	}
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

func (dbif *SqliteAegisDatabaseInterface) GetAllVisibleRepositoryFromNamespace(username string, ns string) ([]*model.Repository, error) {
	pfx := dbif.config.DatabaseTablePrefix
	privateSelectClause := ""
	if len(username) > 0 {
		privateSelectClause = fmt.Sprintf("OR (repo_owner = ? OR repo_acl LIKE ? ESCAPE ?)", )
	}
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT repo_name, repo_description, repo_owner, repo_acl, repo_status
FROM %srepository
WHERE repo_namespace = ? AND (repo_status = 1 OR repo_status = 4 %s)
`, pfx, privateSelectClause))
	if err != nil { return nil, err }
	var rs *sql.Rows
	if len(username) > 0 {
		pattern := strings.ReplaceAll(username, "\\", "\\\\")
		pattern = strings.ReplaceAll(pattern, "%", "\\%")
		pattern = strings.ReplaceAll(pattern, "_", "\\_")
		pattern = "%" + pattern + "%"
		rs, err = stmt.Query(username, pattern, "\\")
	} else {
		rs, err = stmt.Query(ns, username)
	}
	if err != nil { return nil, err }
	defer rs.Close()
	res := make([]*model.Repository, 0)
	for rs.Next() {
		var name, desc, owner, acl string
		var status int64
		err = rs.Scan(&name, &desc, &owner, &acl, &status)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, ns, name)
		res = append(res, &model.Repository{
			Namespace: ns,
			Name: name,
			Owner: owner,
			Description: desc,
			AccessControlList: a,
			Status: model.AegisRepositoryStatus(status),
			Repository: gitlib.NewLocalGitRepository(ns, name, p),
		})
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllRepositoryFromNamespace(ns string) (map[string]*model.Repository, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT repo_name, repo_description, repo_owner, repo_acl, repo_status
FROM %srepository
WHERE repo_namespace = ?
`, pfx))
	if err != nil { return nil, err }
	rs, err := stmt.Query(ns)
	if err != nil { return nil, err }
	defer rs.Close()
	res := make(map[string]*model.Repository, 0)
	for rs.Next() {
		var name, desc, acl, owner string
		var status int64
		err = rs.Scan(&name, &desc, &owner, &acl, &status)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, ns, name)
		res[name] = &model.Repository{
			Namespace: ns,
			Name: name,
			Owner: owner,
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
SET ns_title = ?, ns_description = ?, ns_email = ?, ns_owner = ?, ns_status = ?
WHERE ns_name = ?
`, pfx))
	if err != nil { tx.Rollback(); return err }
	_, err = stmt.Exec(nsobj.Title, nsobj.Description, nsobj.Email, nsobj.Owner, nsobj.Status, name)
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

func (dbif *SqliteAegisDatabaseInterface) CreateRepository(ns string, name string, owner string) (*model.Repository, error) {
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
	_, err = stmt1.Exec(fullName, ns, name, new(string), new(string), model.REPO_NORMAL_PUBLIC, owner)
	if err != nil { return nil, err }
	p := path.Join(dbif.config.GitRoot, ns, name)
	err = os.RemoveAll(p)
	if err != nil { return nil, err }
	if err = os.MkdirAll(p, os.ModeDir|0775); err != nil {
		return nil, err
	}
	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = p
	if err = cmd.Run(); err != nil { return nil, err }
	if err = tx.Commit(); err != nil { return nil, err }
	r, err := model.NewRepository(ns, name, gitlib.NewLocalGitRepository(ns, name, p))
	r.Owner = owner
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
	defer tx.Rollback()
	stmt2, err := tx.Prepare(fmt.Sprintf(`
UPDATE %srepository
SET repo_description = ?, repo_owner = ?, repo_status = ?
WHERE rowid = ?
`, pfx))
	if err != nil { return err }
	_, err = stmt2.Exec(robj.Description, robj.Owner, robj.Status, rowid)
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	robj.Repository.Description = robj.Description
	// we don't deal with error here because it's not critical.
	err = robj.Repository.SyncLocalDescription()
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

func (dbif *SqliteAegisDatabaseInterface) GetAllNamespaces(pageNum int, pageSize int) (map[string]*model.Namespace, error) {
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
	res := make(map[string]*model.Namespace, 0)
	var name, title, desc, email, owner, acl string
	var reg_date int64
	var status int
	for r.Next() {
		err = r.Scan(&name, &title, &desc, &email, &owner, &reg_date, &status, &acl)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res[name] = &model.Namespace{
			Name: name,
			Title: title,
			Description: desc,
			Email: email,
			Owner: owner,
			RegisterTime: reg_date,
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		}
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllRepositories(pageNum int, pageSize int) ([]*model.Repository, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT repo_namespace, repo_name, repo_description, repo_acl, repo_owner, repo_status
FROM %srepository
ORDER BY rowid ASC LIMIT ? OFFSET ?
`, pfx))
	if err != nil { return nil, err }
	defer stmt.Close()
	r, err := stmt.Query(pageSize, pageNum * pageSize)
	if err != nil { return nil, err }
	defer r.Close()
	res := make([]*model.Repository, 0)
	var ns, name, desc, acl, owner string
	var status int
	for r.Next() {
		err = r.Scan(&ns, &name, &desc, &acl, &owner, &status)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, ns, name)
		res = append(res, &model.Repository{
			Namespace: ns,
			Name: name,
			Owner: owner,
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

func (dbif *SqliteAegisDatabaseInterface) CountAllRepositoriesSearchResult(q string) (int64, error) {
	pfx := dbif.config.DatabaseTablePrefix
	searchPattern := strings.ReplaceAll(q, "\\", "\\\\")
	searchPattern = strings.ReplaceAll(searchPattern, "%", "\\%")
	searchPattern = strings.ReplaceAll(searchPattern, "_", "\\_")
	searchPattern = "%" + searchPattern + "%"
	stmt, err := dbif.connection.Prepare(
		fmt.Sprintf(`
SELECT COUNT(*) FROM %srepository
WHERE (repo_name LIKE ? ESCAPE ? OR repo_namespace LIKE ? ESCAPE ?)`, pfx),
	)
	if err != nil { return 0, err }
	defer stmt.Close()
	r := stmt.QueryRow(searchPattern, "\\", searchPattern, "\\")
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

func (dbif *SqliteAegisDatabaseInterface) SearchForNamespace(k string, pageNum int, pageSize int) (map[string]*model.Namespace, error) {
	pfx := dbif.config.DatabaseTablePrefix
	pattern := strings.ReplaceAll(k, "\\", "\\\\")
	pattern = strings.ReplaceAll(pattern, "%", "\\%")
	pattern = strings.ReplaceAll(pattern, "_", "\\_")
	pattern = "%" + pattern + "%"
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
	res := make(map[string]*model.Namespace, 0)
	var name, title, desc, email, owner, acl string
	var reg_date int64
	var status int
	for r.Next() {
		err = r.Scan(&name, &title, &desc, &email, &owner, &reg_date, &status, &acl)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res[name] = &model.Namespace{
			Name: name,
			Title: title,
			Description: desc,
			Email: email,
			Owner: owner,
			RegisterTime: reg_date,
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		}
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
SELECT repo_namespace, repo_name, repo_description, repo_acl, repo_owner, repo_status
FROM %srepository
WHERE (repo_namespace LIKE ? ESCAPE ? OR repo_name LIKE ? ESCAPE ?)
ORDER BY rowid ASC LIMIT ? OFFSET ?
`, pfx))
	if err != nil { return nil, err }
	defer stmt.Close()
	r, err := stmt.Query(pattern, "\\", pattern, "\\", pageSize, pageNum * pageSize)
	if err != nil { return nil, err }
	defer r.Close()
	res := make([]*model.Repository, 0)
	var ns, name, desc, acl, owner string
	var status int
	for r.Next() {
		err = r.Scan(&ns, &name, &desc, &acl, &owner, &status)
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

func (dbif *SqliteAegisDatabaseInterface) SetNamespaceACL(nsName string, targetUserName string, aclt *model.ACLTuple) error {
	pfx := dbif.config.DatabaseTablePrefix
	stmt1, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT ns_acl FROM %snamespace WHERE ns_name = ?
`, pfx))
	if err != nil { return err }
	defer stmt1.Close()
	r := stmt1.QueryRow(nsName)
	if r.Err() != nil { stmt1.Close(); return r.Err() }
	var aclStr string
	err = r.Scan(&aclStr)
	if err != nil { stmt1.Close(); return err }
	acl, err := model.ParseACL(aclStr)
	if err != nil { return err }
	if acl == nil {
		acl = &model.ACL{
			Version: "0",
			ACL: make(map[string]*model.ACLTuple, 0),
		}
	}
	acl.ACL[targetUserName] = aclt
	aclStr, err = acl.SerializeACL()
	if err != nil { return err }
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	defer tx.Rollback()
	stmt2, err := tx.Prepare(fmt.Sprintf(`
UPDATE %snamespace SET ns_acl = ? WHERE ns_name = ?
`, pfx))
	if err != nil { return err }
	_, err = stmt2.Exec(aclStr, nsName)
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) SetRepositoryACL(nsName string, repoName string, targetUserName string, aclt *model.ACLTuple) error {
	pfx := dbif.config.DatabaseTablePrefix
	stmt1, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT repo_acl FROM %srepository WHERE repo_namespace = ? AND repo_name = ?
`, pfx))
	if err != nil { return err }
	defer stmt1.Close()
	r := stmt1.QueryRow(nsName, repoName)
	if r.Err() != nil { stmt1.Close(); return r.Err() }
	var aclStr string
	err = r.Scan(&aclStr)
	if err != nil { stmt1.Close(); return err }
	acl, err := model.ParseACL(aclStr)
	if err != nil { return err }
	if acl == nil {
		acl = &model.ACL{
			Version: "0",
			ACL: make(map[string]*model.ACLTuple, 0),
		}
	}
	acl.ACL[targetUserName] = aclt
	aclStr, err = acl.SerializeACL()
	if err != nil { return err }
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	defer tx.Rollback()
	stmt2, err := tx.Prepare(fmt.Sprintf(`
UPDATE %srepository SET repo_acl = ? WHERE repo_namespace = ? AND repo_name = ?
`, pfx))
	if err != nil { return err }
	_, err = stmt2.Exec(aclStr, nsName, repoName)
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllComprisingNamespace(username string) (map[string]*model.Namespace, error) {
	pfx := dbif.config.DatabaseTablePrefix
	pattern := strings.ReplaceAll(username, "\\", "\\\\")
	pattern = strings.ReplaceAll(pattern, "%", "\\%")
	pattern = strings.ReplaceAll(pattern, "_", "\\_")
	pattern = "%" + pattern + "%"
	stmt1, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_status, ns_acl
FROM %snamespace
WHERE ns_owner = ? OR ns_acl LIKE ? ESCAPE ?
`, pfx))
	if err != nil { return nil, err }
	defer stmt1.Close()
	r, err := stmt1.Query(username, pattern, "\\")
	if err != nil { return nil, err }
	defer r.Close()
	res := make(map[string]*model.Namespace, 0)
	for r.Next() {
		var name, title, desc, email, owner, acl string
		var regtime int64
		var status int64
		err = r.Scan(&name, &title, &desc, &email, &owner, &regtime, &status, &acl)
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

func (dbif *SqliteAegisDatabaseInterface) GetAllVisibleNamespacePaginated(username string, pageNum int, pageSize int) (map[string]*model.Namespace, error) {
	pfx := dbif.config.DatabaseTablePrefix
	privateSelectClause := ""
	if len(username) > 0 {
		privateSelectClause = fmt.Sprintf("OR (ns_owner = ? OR ns_acl LIKE ? ESCAPE ?)", )
	}
	stmt1, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_status, ns_acl
FROM %snamespace
WHERE ns_status = 1 %s
ORDER BY rowid ASC LIMIT ? OFFSET ?
`, pfx, privateSelectClause))
	if err != nil { return nil, err }
	defer stmt1.Close()
	var rs *sql.Rows
	if len(username) > 0 {
		pattern := strings.ReplaceAll(username, "\\", "\\\\")
		pattern = strings.ReplaceAll(pattern, "%", "\\%")
		pattern = strings.ReplaceAll(pattern, "_", "\\_")
		pattern = "%" + pattern + "%"
		rs, err = stmt1.Query(username, pattern, "\\", pageSize, pageNum*pageSize)
	} else {
		rs, err = stmt1.Query(pageSize, pageNum*pageSize)
	}
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

func (dbif *SqliteAegisDatabaseInterface) GetAllVisibleRepositoryPaginated(username string, pageNum int, pageSize int) ([]*model.Repository, error) {
	// private ns, not ns member, not repo member --> invisible.
	// private ns, ns member --> all repo visible.
	// private ns, not ns member, repo member --> only repo is visible.
	// public ns --> all repo visible.
	// ------>
	// select repo from all public ns
	// select repo from all ns member
	// select repo from all repo member
	pfx := dbif.config.DatabaseTablePrefix
	nsPrivateClause := ""
	repoPrivateClause := ""
	if len(username) > 0 {
		nsPrivateClause = "OR (ns_owner = ? OR ns_acl LIKE ? ESCAPE ?)"
		repoPrivateClause = "OR (repo_owner = ? OR repo_acl LIKE ? ESCAPE ?)"
	}
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status
FROM %srepository
INNER JOIN (SELECT ns_name FROM %snamespace WHERE ns_status = 1 %s) a
ON %srepository.repo_namespace = a.ns_name
WHERE repo_status = 1 OR repo_status = 4 %s
ORDER BY rowid ASC LIMIT ? OFFSET ?
`, pfx, pfx, nsPrivateClause, pfx, repoPrivateClause))
	defer stmt.Close()
	if err != nil { return nil, err }
	var rs *sql.Rows
	if len(username) > 0 {
		pattern := ToSqlSearchPattern(username)
		rs, err = stmt.Query(username, pattern, "\\", username, pattern, "\\", pageSize, pageNum*pageSize)
	} else {
		rs, err = stmt.Query(pageSize, pageNum*pageSize)
	}
	if err != nil { return nil, err }
	defer rs.Close()
	res := make([]*model.Repository, 0)
	for rs.Next() {
		var ns, name, desc, owner, acl string
		var status int64
		err = rs.Scan(&ns, &name, &desc, &owner, &acl, &status)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, ns, name)
		res = append(res, &model.Repository{
			Namespace: ns,
			Name: name,
			Owner: owner,
			Description: desc,
			AccessControlList: a,
			Status: model.AegisRepositoryStatus(status),
			Repository: gitlib.NewLocalGitRepository(ns, name, p),
		})
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) CountAllVisibleNamespace(username string) (int64, error) {
	pfx := dbif.config.DatabaseTablePrefix
	privateSelectClause := ""
	if len(username) > 0 {
		privateSelectClause = fmt.Sprintf("OR (ns_owner = ? OR ns_acl LIKE ? ESCAPE ?)", )
	}
	stmt1, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT COUNT(*) FROM %snamespace WHERE ns_status = 1 %s
`, pfx, privateSelectClause))
	if err != nil { return 0, err }
	defer stmt1.Close()
	var r *sql.Row
	if len(username) > 0 {
		pattern := strings.ReplaceAll(username, "\\", "\\\\")
		pattern = strings.ReplaceAll(pattern, "%", "\\%")
		pattern = strings.ReplaceAll(pattern, "_", "\\_")
		pattern = "%" + pattern + "%"
		r = stmt1.QueryRow(username, pattern, "\\")
	} else {
		r = stmt1.QueryRow()
	}
	var res int64
	err = r.Scan(&res)
	if err != nil { return 0, r.Err() }
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) CountAllVisibleRepositories(username string) (int64, error) {
	pfx := dbif.config.DatabaseTablePrefix
	nsPrivateClause := ""
	repoPrivateClause := ""
	if len(username) > 0 {
		nsPrivateClause = "OR (ns_owner = ? OR ns_acl LIKE ? ESCAPE ?)"
		repoPrivateClause = "OR (repo_owner = ? OR repo_acl LIKE ? ESCAPE ?)"
	}
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT COUNT(*)
FROM %srepository
INNER JOIN (SELECT ns_name FROM %snamespace WHERE ns_status = 1 %s) a
ON %srepository.repo_namespace = a.ns_name
WHERE repo_status = 1 OR repo_status = 4 %s
`, pfx, pfx, nsPrivateClause, pfx, repoPrivateClause))
	if err != nil { return 0, err }
	var r *sql.Row
	if len(username) > 0 {
		pattern := ToSqlSearchPattern(username)
		r = stmt.QueryRow(
			username, pattern, "\\",
			username, pattern, "\\",
		)
	} else {
		r = stmt.QueryRow()
	}
	var res int64
	err = r.Scan(&res)
	if err != nil { return 0, r.Err() }
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) SearchAllVisibleNamespacePaginated(username string, query string, pageNum int, pageSize int) (map[string]*model.Namespace, error) {
	pfx := dbif.config.DatabaseTablePrefix
	queryPattern := ToSqlSearchPattern(query)
	privateSelectClause := ""
	if len(username) > 0 {
		privateSelectClause = fmt.Sprintf("OR (ns_owner = ? OR ns_acl LIKE ? ESCAPE ?)", )
	}
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_status, ns_acl
FROM %snamespace
WHERE
    (ns_name LIKE ? ESCAPE ? OR ns_title LIKE ? ESCAPE ?)
    AND (ns_status = 1 %s)
ORDER BY rowid ASC LIMIT ? OFFSET ?
`, pfx, privateSelectClause))
	if err != nil { return nil, err }
	defer stmt.Close()
	var r *sql.Rows
	if len(username) > 0 {
		usernamePattern := ToSqlSearchPattern(username)
		r, err = stmt.Query(queryPattern, "\\", queryPattern, "\\", username, usernamePattern, "\\", pageSize, pageNum * pageSize)
	} else {
		r, err = stmt.Query(queryPattern, "\\", queryPattern, "\\", pageSize, pageNum * pageSize)
	}
	if err != nil { return nil, err }
	defer r.Close()
	res := make(map[string]*model.Namespace, 0)
	var name, title, desc, email, owner, acl string
	var reg_date int64
	var status int
	for r.Next() {
		err = r.Scan(&name, &title, &desc, &email, &owner, &reg_date, &status, &acl)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res[name] = &model.Namespace{
			Name: name,
			Title: title,
			Description: desc,
			Email: email,
			Owner: owner,
			RegisterTime: reg_date,
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		}
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) SearchAllVisibleRepositoryPaginated(username string, query string, pageNum int, pageSize int) ([]*model.Repository, error) {
	pfx := dbif.config.DatabaseTablePrefix
	queryPattern := ToSqlSearchPattern(query)
	nsPrivateClause := ""
	repoPrivateClause := ""
	if len(username) > 0 {
		nsPrivateClause = "OR (ns_owner = ? OR ns_acl LIKE ? ESCAPE ?)"
		repoPrivateClause = "OR (repo_owner = ? OR repo_acl LIKE ? ESCAPE ?)"
	}
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT repo_namespace, repo_name, repo_description, repo_acl, repo_status
FROM %srepository
INNER JOIN (SELECT ns_name FROM %snamespace WHERE ns_status = 1 %s) a
ON %srepository.repo_namespace = a.ns_name
WHERE
    (repo_name LIKE ? ESCAPE ? OR repo_namespace LIKE ? ESCAPE ?)
    AND (repo_status = 1 OR repo_status = 4 %s)
ORDER BY rowid ASC LIMIT ? OFFSET ?
`, pfx, pfx, nsPrivateClause, pfx, repoPrivateClause))
	if err != nil { return nil, err }
	defer stmt.Close()
	var r *sql.Rows
	if len(username) > 0 {
		usernamePattern := ToSqlSearchPattern(username)
		r, err = stmt.Query(
			username, usernamePattern, "\\",
			queryPattern, "\\",
			queryPattern, "\\",
			username, usernamePattern, "\\",
			pageSize, pageNum * pageSize,
		)
	} else {
		r, err = stmt.Query(
			queryPattern, "\\",
			queryPattern, "\\",
			pageSize, pageNum * pageSize,
		)
	}
	if err != nil { return nil, err }
	defer r.Close()
	res := make([]*model.Repository, 0)
	for r.Next() {
		var ns, name, desc, acl string
		var status int64
		err = r.Scan(&ns, &name, &desc, &acl, &status)
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


func (dbif *SqliteAegisDatabaseInterface) CountAllVisibleNamespaceSearchResult(username string, pattern string) (int64, error) {
	pfx := dbif.config.DatabaseTablePrefix
	searchPattern := ToSqlSearchPattern(pattern)
	privateSelectClause := ""
	if len(username) > 0 {
		privateSelectClause = fmt.Sprintf("AND (ns_owner = ? OR ns_acl LIKE ? ESCAPE ?)", )
	}
	stmt, err := dbif.connection.Prepare(
		fmt.Sprintf(`
SELECT COUNT(*) FROM %snamespace
WHERE (ns_name LIKE ? ESCAPE ? OR ns_title LIKE ? ESCAPE ?)
%s
`, pfx, privateSelectClause),
	)
	if err != nil { return 0, err }
	defer stmt.Close()
	var r *sql.Row
	if len(username) > 0 {
		usernamePattern := ToSqlSearchPattern(username)
		r = stmt.QueryRow(searchPattern, "\\", searchPattern, "\\", username, usernamePattern, "\\")
	} else {
		r = stmt.QueryRow(searchPattern, "\\", searchPattern, "\\")
	}
	if r.Err() != nil { return 0, r.Err() }
	var res int64
	err = r.Scan(&res)
	if err != nil { return 0, r.Err() }
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) CountAllVisibleRepositoriesSearchResult(username string, pattern string) (int64, error) {
	pfx := dbif.config.DatabaseTablePrefix
	searchPattern := ToSqlSearchPattern(pattern)
	nsPrivateClause := ""
	repoPrivateClause := ""
	if len(username) > 0 {
		nsPrivateClause = "OR (ns_owner = ? OR ns_acl LIKE ? ESCAPE ?)"
		repoPrivateClause = "OR (repo_owner = ? OR repo_acl LIKE ? ESCAPE ?)"
	}
	stmt, err := dbif.connection.Prepare(
		fmt.Sprintf(`
SELECT COUNT(*)
FROM %srepository
INNER JOIN (SELECT ns_name FROM %snamespace WHERE ns_status = 1 %s) a
ON %srepository.repo_namespace = a.ns_name
WHERE
    (repo_name LIKE ? ESCAPE ? OR repo_namespace LIKE ? ESCAPE ?)
    AND (repo_status = 1 OR repo_status = 4 %s)
`, pfx, pfx, nsPrivateClause, pfx, repoPrivateClause),
	)
	if err != nil { return 0, err }
	defer stmt.Close()
	var r *sql.Row
	
	if len(username) > 0 {
		usernamePattern := ToSqlSearchPattern(username)
		r = stmt.QueryRow(
			username, usernamePattern, "\\",
			searchPattern, "\\",
			searchPattern, "\\",
			username, usernamePattern, "\\",
		)
	} else {
		r = stmt.QueryRow(searchPattern, "\\", searchPattern, "\\")
	}
	if r.Err() != nil { return 0, r.Err() }
	var res int64
	err = r.Scan(&res)
	if err != nil { return 0, r.Err() }
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllRepositoryIssue(ns string, name string) ([]*model.Issue, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT rowid, issue_id, issue_author, issue_title, issue_content
FROM %sissue
WHERE repo_namespace = ? AND repo_name = ?
`, pfx))
	if err != nil { return nil, err }
	defer stmt.Close()
	r, err := stmt.Query(ns, name)
	if err != nil { return nil, err }
	defer r.Close()
	res := make([]*model.Issue, 0)
	for r.Next() {
		var issueAbsId int64
		var issueId int
		var issueAuthor, issueTitle, issueContent string
		err = r.Scan(&issueAbsId, &issueId, &issueAuthor, &issueTitle, &issueContent)
		if err != nil { return nil, err }
		res = append(res, &model.Issue{
			IssueAbsId: issueAbsId,
			RepoNamespace: ns,
			RepoName: name,
			IssueId: issueId,
			IssueTitle: issueTitle,
			IssueAuthor: issueAuthor,
			IssueContent: issueContent,
		})
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) GetRepositoryIssue(ns string, name string, iid int) (*model.Issue, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT rowid, issue_timestamp, issue_author, issue_title, issue_content, issue_status
FROM %sissue
WHERE repo_namespace = ? AND repo_name = ? AND issue_id = ?
`, pfx))
	if err != nil { return nil, err }
	defer stmt.Close()
	r := stmt.QueryRow(ns, name, iid)
	if r.Err() != nil { return nil, r.Err() }
	var absid, timestamp int64
	var status int
	var author, title, content string
	err = r.Scan(&absid, &timestamp, &author, &title, &content, &status)
	if err != nil { return nil, err }
	return &model.Issue{
		IssueAbsId: absid,
		RepoNamespace: ns,
		RepoName: name,
		IssueId: iid,
		IssueAuthor: author,
		IssueTitle: title,
		IssueTime: timestamp,
		IssueContent: content,
		IssueStatus: status,
	}, nil
}

func (dbif *SqliteAegisDatabaseInterface) CountAllRepositoryIssue(ns string, name string) (int, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT COUNT(*)
FROM %sissue
WHERE repo_namespace = ? AND repo_name = ?
`, pfx))
	if err != nil { return 0, err }
	defer stmt.Close()
	r := stmt.QueryRow(ns, name)
	if r.Err() != nil { return 0, r.Err() }
	var res int
	err = r.Scan(&res)
	if err != nil { return 0, err }
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) NewRepositoryIssue(ns string, name string, author string, title string, content string) (int64, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT COUNT(*) FROM %sissue WHERE repo_namespace = ? AND repo_name = ?
`, pfx))
	if err != nil { return 0, err }
	defer stmt.Close()
	r := stmt.QueryRow(ns, name)
	if r.Err() != nil { return 0, err }
	var res int64
	err = r.Scan(&res)
	if err != nil { return 0, err }
	res += 1
	tx, err := dbif.connection.Begin()
	if err != nil { return 0, err }
	defer tx.Rollback()
	stmt2, err := tx.Prepare(fmt.Sprintf(`
INSERT INTO %sissue(repo_namespace, repo_name, issue_id, issue_timestamp, issue_author, issue_title, issue_content, issue_status)
VALUES (?,?,?,?,?,?,?)
`, pfx))
	_, err = stmt2.Exec(ns, name, res, time.Now().Unix(), author, title, content, model.ISSUE_OPENED)
	if err != nil { return 0, err }
	err = tx.Commit()
	if err != nil { return 0, err }
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) HardDeleteRepositoryIssue(ns string, name string, issueId int) error {
	pfx := dbif.config.DatabaseTablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	defer tx.Rollback()
	stmt, err := tx.Prepare(fmt.Sprintf(`
DELETE FROM %sissue WHERE repo_namespace = ? AND repo_name = ? AND issue_id = ?
`, pfx))
	if err != nil { return err }
	defer stmt.Close()
	_, err = stmt.Exec(ns, name, issueId)
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllIssueEvent(ns string, name string, issueId int) ([]*model.IssueEvent, error) {
	pfx := dbif.config.DatabaseTablePrefix
	stmt1, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT rowid FROM %sissue WHERE repo_namespace = ? AND repo_name = ?
`, pfx))
	if err != nil { return nil, err }
	defer stmt1.Close()
	r := stmt1.QueryRow(ns, name)
	if r.Err() != nil { return nil, r.Err() }
	var absId int
	err = r.Scan(&absId)
	if err != nil { return nil, err }
	stmt2, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT rowid, issue_event_type, issue_event_time, issue_event_author, issue_event_content
FROM %sissue_event
WHERE issue_abs_id = ?
`, pfx))
	if err != nil { return nil, err }
	defer stmt2.Close()
	rs, err := stmt2.Query(absId)
	if err != nil { return nil, err }
	defer rs.Close()
	res := make([]*model.IssueEvent, 0)
	for rs.Next() {
		var author, content string
		var eventType int
		var eventAbsId, timestamp int64
		err = rs.Scan(&eventAbsId, &eventType, &timestamp, &author, &content)
		if err != nil { return nil, err }
		res = append(res, &model.IssueEvent{
			EventAbsId: eventAbsId,
			EventType: eventType,
			EventTimestamp: timestamp,
			EventAuthor: author,
			EventContent: content,
		})
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) NewRepositoryIssueEvent(ns string, name string, issueId int, eType int, author string, content string) error {
	pfx := dbif.config.DatabaseTablePrefix
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT rowid, issue_status FROM %sissue WHERE repo_namespace = ? AND repo_name = ? AND issue_id = ?
`, pfx))
	if err != nil { return err }
	defer stmt.Close()
	r := stmt.QueryRow(ns, name, issueId)
	if r.Err() != nil { return r.Err() }
	var issueAbsId int64
	var issueStatus int
	err = r.Scan(&issueAbsId, &issueStatus)
	if err != nil { return err }
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	defer tx.Rollback()
	stmt2, err := tx.Prepare(fmt.Sprintf(`
INSERT INTO %sissue_event(issue_abs_id, issue_event_type, issue_event_time, issue_event_author, issue_event_content) VALUES (?,?,?,?,?)
`, pfx))
	if err != nil { return err }
	defer stmt2.Close()
	_, err = stmt2.Exec(issueAbsId, eType, time.Now().Unix(), author, content)
	if err != nil { return err }
	newIssueStatus := issueStatus
	if eType == model.EVENT_CLOSED_AS_SOLVED {
		newIssueStatus = model.ISSUE_CLOSED_AS_SOLVED
	} else if eType == model.EVENT_CLOSED_AS_DISCARDED {
		newIssueStatus = model.ISSUE_CLOSED_AS_SOLVED
	} else if eType == model.EVENT_REOPENED {
		newIssueStatus = model.ISSUE_OPENED
	}
	if newIssueStatus != issueStatus {
		stmt3, err := tx.Prepare(fmt.Sprintf(`
UPDATE %sissue SET issue_status = ? WHERE rowid = ?
`, pfx))
		if err != nil { return err }
		defer stmt3.Close()
		_, err = stmt3.Exec(newIssueStatus, issueAbsId)
		if err != nil { return err }
	}
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) HardDeleteRepositoryIssueEvent(eventAbsId int64) error {
	pfx := dbif.config.DatabaseTablePrefix
	tx, err := dbif.connection.Begin()
	if err != nil { return err }
	defer tx.Rollback()
	stmt, err := tx.Prepare(fmt.Sprintf(`
DELETE FROM %sissue_event WHERE rowid = ?
`, pfx))
	if err != nil { return err }
	defer stmt.Close()
	_, err = stmt.Exec(eventAbsId)
	if err != nil { return err }
	err = tx.Commit()
	if err != nil { return err }
	return nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllBelongingNamespace(viewingUser string, userName string) ([]*model.Namespace, error) {
	pfx := dbif.config.DatabaseTablePrefix
	nsStatusClause := "ns_status = 1"
	if len(viewingUser) > 0 {
		if viewingUser == userName {
			nsStatusClause = "1"
		} else {
			nsStatusClause = fmt.Sprintf("ns_acl LIKE ? ESCAPE ?")
		}
	}
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_status, ns_acl
FROM %snamespace
WHERE (%s) AND (ns_owner = ? OR ns_acl LIKE ? ESCAPE ?)
`, pfx, nsStatusClause))
	if err != nil { return nil, err }
	defer stmt.Close()
	userSearchPattern := ToSqlSearchPattern(userName)
	var rs *sql.Rows
	if len(viewingUser) > 0 {
		if viewingUser == userName {
			rs, err = stmt.Query(userName, userSearchPattern, "\\")
		} else {
			viewingPattern := ToSqlSearchPattern(viewingUser)
			rs, err = stmt.Query(viewingPattern, "\\", userName, userSearchPattern, "\\")
		}
	} else {
		rs, err = stmt.Query(userName, userSearchPattern, "\\")
	}
	if err != nil { return nil, err }
	res := make([]*model.Namespace, 0)
	for rs.Next() {
		var name, title, desc, email, owner, acl string
		var regtime int64
		var status int64
		err = rs.Scan(&name, &title, &desc, &email, &owner, &regtime, &status, &acl)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res = append(res, &model.Namespace{
			Name: name,
			Title: title,
			Description: desc,
			Email: email,
			Owner: owner,
			RegisterTime: regtime,
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		})
	}
	return res, nil
}

func (dbif *SqliteAegisDatabaseInterface) GetAllBelongingRepository(viewingUser string, user string, pageNum int, pageSize int) ([]*model.Repository, error) {
	pfx := dbif.config.DatabaseTablePrefix
	nsStatusClause := "ns_status = 1"
	repoStatusClause := "repo_status = 1 OR repo_status = 4"
	if len(viewingUser) > 0 {
		if viewingUser == user {
			nsStatusClause = "1"
			repoStatusClause = "1"
		} else {
			nsStatusClause = "ns_acl LIKE ? ESCAPE ?"
			repoStatusClause = "repo_status = 1 OR repo_status = 4 OR (repo_acl LIKE ? ESCAPE ?)"
		}
	}
	stmt, err := dbif.connection.Prepare(fmt.Sprintf(`
SELECT repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status
FROM %srepository
    INNER JOIN (SELECT ns_name FROM %snamespace WHERE (%s)) a
    ON %srepository.repo_namespace = a.ns_name
WHERE (%s) AND (repo_owner = ?)
`, pfx, pfx, nsStatusClause, pfx, repoStatusClause))
	if err != nil { return nil, err }
	defer stmt.Close()
	var r *sql.Rows
	if len(viewingUser) > 0 {
		if viewingUser == user {
			r, err = stmt.Query(user)
		} else {
			viewingPattern := ToSqlSearchPattern(viewingUser)
			r, err = stmt.Query(viewingPattern, "\\", viewingPattern, "\\", user)
		}
	} else {
		r, err = stmt.Query(user)
	}
	if err != nil { return nil, err }
	res := make([]*model.Repository, 0)
	for r.Next() {
		var ns, name, desc, acl, owner string
		var status int64
		err := r.Scan(&ns, &name, &desc, &owner, &acl, &status)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, ns, name)
		res = append(res, &model.Repository{
			Namespace: ns,
			Name: name,
			Description: desc,
			Owner: owner,
			AccessControlList: a,
			Status:model.AegisRepositoryStatus(status),
			Repository: gitlib.NewLocalGitRepository(ns, name, p),
		})
	}
	return res, nil
}

