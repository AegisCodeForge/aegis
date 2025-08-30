package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/bctnry/aegis/pkg/aegis/db"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/gitlib"
	"github.com/jackc/pgx/v5"
)

func (dbif *PostgresAegisDatabaseInterface) GetUserByName(name string) (*model.AegisUser, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT user_title, user_email, user_bio, user_website, user_reg_datetime, user_password_hash, user_status
FROM %s_user
WHERE user_name = $1
`, pfx), name)
	var title, email, bio, website, password string
	var datetime time.Time
	var status int
	err := stmt.Scan(&title, &email, &bio, &website, &datetime, &password, &status)
	if errors.Is(err, pgx.ErrNoRows) { return nil, db.ErrEntityNotFound }
	if err != nil { return nil, err }
	return &model.AegisUser{
		Name: name,
		Title: title,
		Email: email,
		Bio: bio,
		Website: website,
		PasswordHash: password,
		RegisterTime: datetime.Unix(),
		Status: model.AegisUserStatus(status),
	}, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllAuthKeyByUsername(name string) ([]model.AegisAuthKey, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT key_name, key_text
FROM %s_user_authkey
WHERE user_name = $1
`, pfx), name)
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make([]model.AegisAuthKey, 0)
	for stmt.Next() {
		var kname, ktext string
		err := stmt.Scan(&kname, &ktext)
		if err != nil { return nil, err }
		res = append(res, model.AegisAuthKey{
			UserName: name,
			KeyName: kname,
			KeyText: ktext,
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAuthKeyByName(userName string, keyName string) (*model.AegisAuthKey, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT key_text FROM %s_user_authkey WHERE user_name = $1 AND key_name = $2
`, pfx), userName, keyName)
	var ktext string
	err := stmt.Scan(&ktext)
	if errors.Is(err, pgx.ErrNoRows) { return nil, db.ErrEntityNotFound }
	if err != nil { return nil, err }
	return &model.AegisAuthKey{
		UserName: userName,
		KeyName: keyName,
		KeyText: ktext,
	}, nil
}

func (dbif *PostgresAegisDatabaseInterface) RegisterAuthKey(username string, keyname string, keytext string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s_user_authkey(user_name, key_name, key_text)
VALUES ($1, $2, $3)
`, pfx), username, keyname, keytext)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) UpdateAuthKey(username string, keyname string, keytext string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_user_authkey
SET key_text = $1
WHERE user_name = $2 AND key_name = $3
`, pfx), keytext, username, keyname)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) RemoveAuthKey(username string, keyname string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
DELETE FROM %s_user_authkey
WHERE user_name = $1 AND key_name = $2
`, pfx), username, keyname)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllSignKeyByUsername(name string) ([]model.AegisSigningKey, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT key_name, key_text FROM %s_user_signkey WHERE user_name = $1
`, pfx), name)
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make([]model.AegisSigningKey, 0)
	for stmt.Next() {
		var kname, ktext string
		err := stmt.Scan(&kname, &ktext)
		if err != nil { return nil, err }
		res = append(res, model.AegisSigningKey{
			UserName: name,
			KeyName: kname,
			KeyText: ktext,
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetSignKeyByName(userName string, keyName string) (*model.AegisSigningKey, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT key_text FROM %s_user_signkey WHERE user_name = $1 AND key_name = $2
`, pfx), userName, keyName)
	var ktext string
	err := stmt.Scan(&ktext)
	if errors.Is(err, pgx.ErrNoRows) { return nil, db.ErrEntityNotFound }
	if err != nil { return nil, err }
	return &model.AegisSigningKey{
		UserName: userName,
		KeyName: keyName,
		KeyText: ktext,
	}, nil
}

func (dbif *PostgresAegisDatabaseInterface) UpdateSignKey(username string, keyname string, keytext string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_user_signkey
SET key_text = $1
WHERE user_name = $2 AND key_name = $3
`, pfx), keytext, username, keyname)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) RegisterSignKey(username string, keyname string, keytext string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s_user_signkey(user_name, key_name, key_text)
VALUES ($1, $2, $3)
`, pfx), username, keyname, keytext)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) RemoveSignKey(username string, keyname string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
DELETE FROM %s_user_signkey
WHERE user_name = $1 AND key_name = $2
`, pfx), username, keyname)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) GetNamespaceByName(name string) (*model.Namespace, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status
FROM %s_namespace
WHERE ns_name = $1
`, pfx), name)
	var title, description, email, owner, acl string
	var datetime time.Time
	var status int
	err := stmt.Scan(&title, &description, &email, &owner, &datetime, &acl, &status)
	if errors.Is(err, pgx.ErrNoRows) { return nil, db.ErrEntityNotFound }
	if err != nil { return nil, err }
	a, err := model.ParseACL(acl)
	if err != nil { return nil, err }
	return &model.Namespace{
		Name: name,
		Title: title,
		Description: description,
		Email: email,
		Owner: owner,
		RegisterTime: datetime.Unix(),
		Status: model.AegisNamespaceStatus(status),
		ACL: a,
	}, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetRepositoryByName(nsName string, repoName string) (*model.Repository, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT repo_type, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
WHERE repo_namespace = $1 AND repo_name = $2
`, pfx), nsName, repoName)
	var description, owner, acl, forkOriginNamespace, forkOriginName string
	var repoType, repoStatus int
	err := stmt.Scan(&repoType, &description, &owner, &acl, &repoStatus, &forkOriginNamespace, &forkOriginName)
	if errors.Is(err, pgx.ErrNoRows) { return nil, db.ErrEntityNotFound }
	if err != nil { return nil, err }
	p := path.Join(dbif.config.GitRoot, nsName, repoName)
	localRepo, err := model.CreateLocalRepository(uint8(repoType), nsName, repoName, p)
	if err != nil { return nil, err }
	res, err := model.NewRepository(nsName, repoName, localRepo)
	if err != nil { return nil, err }
	res.Type = uint8(repoType)
	res.Owner = owner
	res.Status = model.AegisRepositoryStatus(repoStatus)
	res.ForkOriginNamespace = forkOriginNamespace
	res.ForkOriginName = forkOriginName
	aclobj, err := model.ParseACL(acl)
	if err != nil { return nil, err }
	res.AccessControlList = aclobj
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllNamespace() (map[string]*model.Namespace, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status
FROM %s_namespace
`, pfx))
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make(map[string]*model.Namespace, 0)
	var name, title, description, email, owner, acl string
	var datetime time.Time
	var status int
	for stmt.Next() {
		err = stmt.Scan(&name, &title, &description, &email, &owner, &datetime, &acl, &status)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res[name] = &model.Namespace{
			Name: name,
			Title: title,
			Description: description,
			Email: email,
			Owner: owner,
			RegisterTime: datetime.Unix(),
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		}
	}
	return  res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllVisibleNamespace(username string) (map[string]*model.Namespace, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status
FROM %s_namespace
WHERE ns_owner = $1 OR ns_acl->'ACL' ? $2
`, pfx))
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make(map[string]*model.Namespace, 0)
	var name, title, description, email, owner, acl string
	var datetime time.Time
	var status int
	for stmt.Next() {
		err = stmt.Scan(&name, &title, &description, &email, &owner, &datetime, &acl, &status)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res[name] = &model.Namespace{
			Name: name,
			Title: title,
			Description: description,
			Email: email,
			Owner: owner,
			RegisterTime: datetime.Unix(),
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		}
	}
	return  res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllVisibleNamespacePaginated(username string, pageNum int, pageSize int) (map[string]*model.Namespace, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	var stmt pgx.Rows
	var err error
	if len(username) > 0 {
		stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status
FROM %s_namespace
WHERE ns_status = 1 OR ns_owner = $3 OR ns_acl->'ACL' ? $3
ORDER BY ns_absid ASC LIMIT $1 OFFSET $2
`, pfx), pageSize, pageNum*pageSize, username)
	} else {
		stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status
FROM %s_namespace
WHERE ns_status = 1
ORDER BY ns_absid ASC LIMIT $1 OFFSET $2
`, pfx), pageSize, pageNum*pageSize)
	}
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make(map[string]*model.Namespace, 0)
	var name, title, description, email, owner, acl string
	var datetime time.Time
	var status int
	for stmt.Next() {
		err = stmt.Scan(&name, &title, &description, &email, &owner, &datetime, &acl, &status)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res[name] = &model.Namespace{
			Name: name,
			Title: title,
			Description: description,
			Email: email,
			Owner: owner,
			RegisterTime: datetime.Unix(),
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		}
	}
	return  res, nil
}

func (dbif *PostgresAegisDatabaseInterface) SearchAllVisibleNamespacePaginated(username string, query string, pageNum int, pageSize int) (map[string]*model.Namespace, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	var stmt pgx.Rows
	var err error
	if len(username) > 0 {
		if len(query) > 0 {
			stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status
FROM %s_namespace
WHERE (ns_name LIKE $1 ESCAPE $2) AND (ns_status = 1 OR ns_owner = $3 OR ns_acl->'ACL' ? $3)
ORDER BY ns_absid ASC LIMIT $4 OFFSET $5
`, pfx), db.ToSqlSearchPattern(query), "%", username, pageSize, pageNum*pageSize)
		} else {
			stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status
FROM %s_namespace
WHERE (ns_status = 1 OR ns_owner = $1 OR ns_acl->'ACL' ? $1)
ORDER BY ns_absid ASC LIMIT $2 OFFSET $3
`, pfx), username, pageSize, pageNum*pageSize)
		}
	} else {
		if len(query) > 0 {
			stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status
FROM %s_namespace
WHERE (ns_name LIKE $1 ESCAPE $2) AND (ns_status = 1)
ORDER BY ns_absid ASC LIMIT $3 OFFSET $4
`, pfx), db.ToSqlSearchPattern(query), "%", pageSize, pageNum*pageSize)
		} else {
			stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status
FROM %s_namespace
WHERE (ns_status = 1)
ORDER BY ns_absid ASC LIMIT $1 OFFSET $2
`, pfx), pageSize, pageNum*pageSize)
		}
	}
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make(map[string]*model.Namespace, 0)
	var name, title, description, email, owner, acl string
	var datetime time.Time
	var status int
	for stmt.Next() {
		err = stmt.Scan(&name, &title, &description, &email, &owner, &datetime, &acl, &status)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res[name] = &model.Namespace{
			Name: name,
			Title: title,
			Description: description,
			Email: email,
			Owner: owner,
			RegisterTime: datetime.Unix(),
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		}
	}
	return  res, nil
	
}

func (dbif *PostgresAegisDatabaseInterface) GetAllVisibleRepositoryPaginated(username string, pageNum int, pageSize int) ([]*model.Repository, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	var stmt pgx.Rows
	var err error
	if len(username) > 0 {
		stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
WHERE repo_status = 1 OR repo_owner = $3 OR repo_acl->'ACL' ? $3
ORDER BY repo_absid ASC LIMIT $1 OFFSET $2
`, pfx), pageSize, pageNum*pageSize, username)
	} else {
		stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
WHERE repo_status = 1
ORDER BY repo_absid ASC LIMIT $1 OFFSET $2
`, pfx), pageSize, pageNum*pageSize)
	}
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make([]*model.Repository, 0)
	var ns, name, description, owner, acl, forkOriginNs, forkOriginName string
	var rType, rStatus int
	for stmt.Next() {
		err = stmt.Scan(&rType, &ns, &name, &description, &owner, &acl, &rStatus, &forkOriginNs, &forkOriginName)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, ns, name)
		m, err := model.CreateLocalRepository(uint8(rType), ns, name, p)
		if err != nil { return nil, err }
		res = append(res, &model.Repository{
			Namespace: ns,
			Name: name,
			Description: description,
			Owner: owner,
			AccessControlList: a,
			Status: model.AegisRepositoryStatus(rStatus),
			Type: uint8(rType),
			ForkOriginNamespace: forkOriginNs,
			ForkOriginName: forkOriginName,
			Repository: m,
		})
	}
	return  res, nil
}

func (dbif *PostgresAegisDatabaseInterface) SearchAllVisibleRepositoryPaginated(username string, query string, pageNum int, pageSize int) ([]*model.Repository, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	var stmt pgx.Rows
	var err error
	if len(query) > 0 {
		queryPattern := db.ToSqlSearchPattern(query)
		if len(username) > 0 {
			stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
WHERE (repo_status = 1 OR repo_owner = $3 OR repo_acl->'ACL' ? $3) AND ((repo_namespace LIKE $4 ESCAPE $5) OR (repo_name LIKE $6 ESCAPE $7))
ORDER BY repo_absid ASC LIMIT $1 OFFSET $2
`, pfx), pageSize, pageNum*pageSize, username, queryPattern, "%", queryPattern, "%")
		} else {
			stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
WHERE repo_status = 1
ORDER BY repo_absid ASC LIMIT $1 OFFSET $2
`, pfx), pageSize, pageNum*pageSize)
		}
	} else {
		if len(username) > 0 {
			stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
WHERE repo_status = 1 OR repo_owner = $3 OR repo_acl->'ACL' ? $3
ORDER BY repo_absid ASC LIMIT $1 OFFSET $2
`, pfx), pageSize, pageNum*pageSize, username)
		} else {
			stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
WHERE repo_status = 1
ORDER BY repo_absid ASC LIMIT $1 OFFSET $2
`, pfx), pageSize, pageNum*pageSize)
		} 
	}
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make([]*model.Repository, 0)
	var ns, name, description, owner, acl, forkOriginNs, forkOriginName string
	var rType, rStatus int
	for stmt.Next() {
		err = stmt.Scan(&rType, &ns, &name, &description, &owner, &acl, &rStatus, &forkOriginNs, &forkOriginName)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, ns, name)
		m, err := model.CreateLocalRepository(uint8(rType), ns, name, p)
		if err != nil { return nil, err }
		res = append(res, &model.Repository{
			Namespace: ns,
			Name: name,
			Description: description,
			Owner: owner,
			AccessControlList: a,
			Status: model.AegisRepositoryStatus(rStatus),
			Type: uint8(rType),
			ForkOriginNamespace: forkOriginNs,
			ForkOriginName: forkOriginName,
			Repository: m,
		})
	}
	return  res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllNamespaceByOwner(name string) (map[string]*model.Namespace, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_status, ns_acl
FROM %s_namespace
WHERE ns_owner = $1
`, pfx), name)
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make(map[string]*model.Namespace, 0)
	for stmt.Next() {
		var name, title, desc, email, owner, acl string
		var regtime time.Time
		var status int64
		err = stmt.Scan(&name, &title, &desc, &email, &owner, &regtime, &status, &acl)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res[name] = &model.Namespace{
			Name: name,
			Title: title,
			Description: desc,
			Email: email,
			Owner: owner,
			RegisterTime: regtime.Unix(),
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		}
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllRepositoryFromNamespace(name string) (map[string]*model.Repository, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
WHERE repo_namespace = $1
`, pfx), name)
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make(map[string]*model.Repository, 0)
	for stmt.Next() {
		var rType, rStatus int
		var repoName, desc, owner, acl, forkOriginNs, forkOriginName string
		err := stmt.Scan(&rType, &repoName, &desc, &owner, &acl, &rStatus, &forkOriginNs, &forkOriginName)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, name, repoName)
		m, err := model.CreateLocalRepository(uint8(rType), name, repoName, p)
		if err != nil { return nil, err }
		res[name] = &model.Repository{
			Namespace: name,
			Name: repoName,
			Description: desc,
			Owner: owner,
			AccessControlList: a,
			Status: model.AegisRepositoryStatus(rStatus),
			Type: uint8(rType),
			ForkOriginNamespace: forkOriginNs,
			ForkOriginName: forkOriginName,
			Repository: m,
		}
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllVisibleRepositoryFromNamespace(username string, ns string) ([]*model.Repository, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	var stmt pgx.Rows
	var err error
	if len(username) > 0 {
		stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
WHERE repo_namespace = $1 AND repo_status = 1
`, pfx), ns)
	} else {
		stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
WHERE repo_namespace = $1 AND (repo_status = 1 OR repo_owner = $2 OR repo_acl->'ACL' ? $2)
`, pfx), ns, username)
	}
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make([]*model.Repository, 0)
	for stmt.Next() {
		var rType, rStatus int
		var repoName, desc, owner, acl, forkOriginNs, forkOriginName string
		err := stmt.Scan(&rType, &repoName, &desc, &owner, &acl, &rStatus, &forkOriginNs, &forkOriginName)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, ns, repoName)
		m, err := model.CreateLocalRepository(uint8(rType), ns, repoName, p)
		if err != nil { return nil, err }
		res = append(res, &model.Repository{
			Namespace: ns,
			Name: repoName,
			Description: desc,
			Owner: owner,
			AccessControlList: a,
			Status: model.AegisRepositoryStatus(rStatus),
			Type: uint8(rType),
			ForkOriginNamespace: forkOriginNs,
			ForkOriginName: forkOriginName,
			Repository: m,
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) RegisterUser(name string, email string, passwordHash string, status model.AegisUserStatus) (*model.AegisUser, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return nil, err }
	defer tx.Rollback(ctx)
	t := time.Now()
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s_user(user_name, user_title, user_email, user_bio, user_website, user_reg_datetime, user_password_hash, user_status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, pfx), name, name, email, new(string), new(string), t, passwordHash, status)
	if err != nil { return nil, err }
	err = tx.Commit(ctx)
	if err != nil { return nil, err }
	return &model.AegisUser{
		Name: name,
		Title: name,
		Email: email,
		Bio: "",
		Website: "",
		PasswordHash: passwordHash,
		RegisterTime: t.Unix(),
		Status: model.AegisUserStatus(status),
	}, nil
}

func (dbif *PostgresAegisDatabaseInterface) UpdateUserInfo(name string, uobj *model.AegisUser) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_user
SET user_title = $1, user_email = $2, user_bio = $3, user_website = $4, user_status = $5
WHERE user_name = $6
`, pfx), uobj.Title, uobj.Email, uobj.Bio, uobj.Website, uobj.Status, name)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) UpdateUserPassword(name string, newPasswordHash string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_user
SET user_password_hash = $1
WHERE user_name = $2
`, pfx), newPasswordHash, name)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) HardDeleteUserByName(name string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
DELETE FROM %s_user WHERE user_name = $1
`, pfx), name)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) UpdateUserStatus(name string, newStatus model.AegisUserStatus) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_user
SET user_status = $1
WHERE user_name = $2
`, pfx), newStatus, name)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) RegisterNamespace(name string, ownerUsername string) (*model.Namespace, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return nil, err }
	defer tx.Rollback(ctx)
	t := time.Now()
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s_namespace(ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, pfx), name, name, new(string), new(string), ownerUsername, t, model.NewACL(), model.NAMESPACE_NORMAL_PUBLIC)
	if err != nil { return nil, err }
	nsPath := path.Join(dbif.config.GitRoot, name)
	err = os.RemoveAll(nsPath)
	if err != nil { return nil, err }
	err = os.Mkdir(nsPath, os.ModeDir|0755)
	if err != nil { return nil, err }
	err = tx.Commit(ctx)
	if err != nil { return nil, err }
	return &model.Namespace{
		Name: name,
		Title: name,
		Description: "",
		Email: "",
		Owner: ownerUsername,
		RegisterTime: t.Unix(),
		Status: model.NAMESPACE_NORMAL_PUBLIC,
		ACL: nil,
		RepositoryList: nil,
		LocalPath: nsPath,
	}, nil
}

func (dbif *PostgresAegisDatabaseInterface) UpdateNamespaceInfo(name string, nsobj *model.Namespace) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_namespace
SET ns_title = $1, ns_description = $2, ns_email = $3, ns_owner = $4, ns_status = $5
WHERE ns_name = $6
`, pfx), nsobj.Name, nsobj.Description, nsobj.Email, nsobj.Owner, nsobj.Status, name)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) UpdateNamespaceOwner(name string, newOwner string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_namespace
SET ns_owner = $1
WHERE ns_name = $2
`, pfx), newOwner, name)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) UpdateNamespaceStatus(name string, newStatus model.AegisNamespaceStatus) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_namespace
SET ns_status = $1
WHERE ns_name = $2
`, pfx), newStatus, name)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) HardDeleteNamespaceByName(name string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
DELETE FROM %s_namespace WHERE ns_name = $1
`, pfx), name)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) CreateRepository(ns string, name string, repoType uint8, owner string) (*model.Repository, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return nil, err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s_repository(repo_type, repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`, pfx), repoType, ns, name, new(string), owner, model.NewACL(), model.REPO_NORMAL_PUBLIC, new(string), new(string))
	if err != nil { return nil, err }
	p := path.Join(dbif.config.GitRoot, ns, name)
	if err = os.RemoveAll(p); err != nil { return nil, err }
	if err = os.MkdirAll(p, os.ModeDir|0775); err != nil { return nil, err }
	lr, err := model.CreateLocalRepository(repoType, ns, name, p)
	if err != nil { return nil, err }
	if err = model.InitLocalRepository(lr); err != nil { return nil, err }
	if err = tx.Commit(ctx); err != nil { return nil, err }
	r, err := model.NewRepository(ns, name, lr)
	if err != nil { return nil, err }
	r.Type = repoType
	r.Owner = owner
	return r, nil
}

func (dbif *PostgresAegisDatabaseInterface) SetUpCloneRepository(originNs string, originName string, targetNs string, targetName string, owner string) (*model.Repository, error) {
	// TODO: fix this for multi vcs support
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return nil, err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s_repository(repo_type, repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`, pfx), model.REPO_TYPE_GIT, targetNs, targetName, new(string), owner, model.NewACL(), model.REPO_NORMAL_PUBLIC, originNs, originName)
	if err != nil { return nil, err }
	originP := path.Join(dbif.config.GitRoot, originNs, originName)
	targetP := path.Join(dbif.config.GitRoot, targetNs, targetName)
	if err = os.RemoveAll(targetP); err != nil { return nil, err }
	if err = os.MkdirAll(targetP, os.ModeDir|0775); err != nil { return nil, err }
	originLr, err := model.CreateLocalRepository(model.REPO_TYPE_GIT, originNs, originName, originP)
	if err != nil { return nil, err }
	targetLr, err := model.CreateLocalForkOf(originLr, targetNs, targetName, targetP)
	if err = tx.Commit(ctx); err != nil { return nil, err }
	r, err := model.NewRepository(targetNs, targetName, targetLr)
	if err != nil { return nil, err }
	r.Type = model.REPO_TYPE_GIT
	r.Owner = owner
	return r, nil
}

func (dbif *PostgresAegisDatabaseInterface) UpdateRepositoryInfo(ns string, name string, robj *model.Repository) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_repository
SET repo_description = $1, repo_owner = $2, repo_status = $3
WHERE repo_namespace = $4 AND repo_name = $5
`, pfx), robj.Description, robj.Owner, robj.Status, ns, name)
	if err != nil { return err }
	if err = tx.Commit(ctx); err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) UpdateRepositoryStatus(ns string, name string, status model.AegisRepositoryStatus) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_repository
SET repo_status = $1
WHERE repo_namespace = $2 AND repo_name = $3
`, pfx), status, ns, name)
	if err != nil { return err }
	if err = tx.Commit(ctx); err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) HardDeleteRepository(ns string, name string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
DELETE FROM %s_repository
WHERE repo_namespace = $1 AND repo_name = $2
`, pfx), ns, name)
	if err != nil { return err }
	if err = tx.Commit(ctx); err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllUsers(pageNum int, pageSize int) ([]*model.AegisUser, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT user_name, user_title, user_email, user_bio, user_website, user_reg_datetime, user_password_hash, user_status
FROM %s_user
ORDER BY user_id ASC LIMIT $1 OFFSET $2
`, pfx), pageSize, pageNum*pageSize)
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make([]*model.AegisUser, 0)
	var name, title, email, bio, website, hash string
	var rt time.Time
	var status int
	for stmt.Next() {
		err := stmt.Scan(&name, &title, &email, &bio, &website, &rt, &hash, &status)
		if err != nil { return nil, err }
		res = append(res, &model.AegisUser{
			Name: name,
			Title: title,
			Email: email,
			Bio: bio,
			Website: website,
			PasswordHash: hash,
			RegisterTime: rt.Unix(),
			Status: model.AegisUserStatus(status),
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllNamespaces(pageNum int, pageSize int) (map[string]*model.Namespace, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status
FROM %s_namespace
ORDER BY ns_absid ASC LIMIT $1 OFFSET $2
`, pfx), pageSize, pageNum*pageSize)
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make(map[string]*model.Namespace, 0)
	var name, title, desc, email, owner, acl string
	var rt time.Time
	var status int
	for stmt.Next() {
		err := stmt.Scan(&name, &title, &desc, &email, &owner, &rt, &acl, &status)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, name)
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res[name] = &model.Namespace{
			Name: name,
			Title: title,
			Description: desc,
			Email: email,
			Owner: owner,
			RegisterTime: rt.Unix(),
			Status: model.AegisNamespaceStatus(status),
			ACL: a,
			RepositoryList: nil,
			LocalPath: p,
		}
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllRepositories(pageNum int, pageSize int) ([]*model.Repository, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
ORDER BY repo_absid ASC LIMIT $1 OFFSET $2
`, pfx), pageSize, pageNum*pageSize)
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make([]*model.Repository, 0)
	var rType int
	var ns, name, desc, owner, acl, forkOriginNs, forkOriginName string
	var rStatus int
	for stmt.Next() {
		err = stmt.Scan(&rType, &ns, &name, &desc, &owner, &acl, &rStatus, &forkOriginNs, &forkOriginName)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, ns, name)
		lr, err := model.CreateLocalRepository(uint8(rType), ns, name, p)
		if err != nil { return nil, err }
		res = append(res, &model.Repository{
			Type: uint8(rType),
			Namespace: ns,
			Name: name,
			Owner: owner,
			Description: desc,
			AccessControlList: a,
			Status: model.AegisRepositoryStatus(rStatus),
			Repository: lr,
			ForkOriginNamespace: forkOriginNs,
			ForkOriginName: forkOriginName,
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) CountAllUser() (int64, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_user
`, pfx))
	var r int64
	err := stmt.Scan(&r)
	if err != nil { return 0, err }
	return r, nil
}

func (dbif *PostgresAegisDatabaseInterface) CountAllNamespace() (int64, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_namespace
`, pfx))
	var r int64
	err := stmt.Scan(&r)
	if err != nil { return 0, err }
	return r, nil
}

func (dbif *PostgresAegisDatabaseInterface) CountAllRepositories() (int64, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_repository
`, pfx))
	var r int64
	err := stmt.Scan(&r)
	if err != nil { return 0, err }
	return r, nil
}

func (dbif *PostgresAegisDatabaseInterface) CountAllRepositoriesSearchResult(q string) (int64, error) {
	if len(q) <= 0 { return dbif.CountAllRepositories() }
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_repository WHERE repo_ns LIKE $1 ESCAPE $2 OR repo_name LIKE $1 ESCAPE $2
`, pfx), db.ToSqlSearchPattern(q), "%")
	var r int64
	err := stmt.Scan(&r)
	if err != nil { return 0, err }
	return r, nil
}

func (dbif *PostgresAegisDatabaseInterface) CountAllVisibleNamespace(username string) (int64, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	var stmt pgx.Row
	if len(username) > 0 {
		stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_namespace WHERE ns_status = 1 OR ns_owner = $1 OR ns_acl->'ACL' ? $1
`, pfx), username)
	} else {
		stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_namespace WHERE ns_status = 1
`, pfx))
	}
	var r int64
	err := stmt.Scan(&r)
	if err != nil { return 0, err }
	return r, nil
}

func (dbif *PostgresAegisDatabaseInterface) CountAllVisibleRepositories(username string) (int64, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	var stmt pgx.Row
	if len(username) > 0 {
		stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_repository
INNER JOIN (SELECT ns_name FROM %s_namespace WHERE ns_status = 1 OR ns_owner = $1 OR ns_acl->'ACL' ? $1) a
ON %s_repository.repo_namespace = a.ns_name
WHERE repo_status = 1 OR repo_status = 4 OR repo_owner = $1 OR repo_acl->'ACL' ? $1
`, pfx, pfx, pfx), username)
	} else {
		stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_repository
INNER JOIN (SELECT ns_name FROM %s_namespace WHERE ns_status = 1) a
ON %s_repository.repo_namespace = a.ns_name
WHERE repo_status = 1 or repo_status = 4
`, pfx, pfx, pfx))
	}
	var r int64
	err := stmt.Scan(&r)
	if err != nil { return 0, err }
	return r, nil
}

func (dbif *PostgresAegisDatabaseInterface) SearchForUser(k string, pageNum int, pageSize int) ([]*model.AegisUser, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	pattern := db.ToSqlSearchPattern(k)
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT user_name, user_title, user_email, user_bio, user_website, user_reg_datetime, user_password_hash, user_status
FROM %s_user
WHERE user_name LIKE $1 ESCAPE $2 OR user_title LIKE $1 ESCAPE $2
ORDER BY user_id ASC LIMIT $3 OFFSET $4
`, pfx), pattern, "%", pageSize, pageNum*pageSize)
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make([]*model.AegisUser, 0)
	var name, title, email, bio, website, hash string
	var dt time.Time
	var st int
	for stmt.Next() {
		err = stmt.Scan(&name, &title, &email, &bio, &website, &dt, &hash, &st)
		if err != nil { return nil, err }
		res = append(res, &model.AegisUser{
			Name: name,
			Title: title,
			Email: email,
			Bio: bio,
			Website: website,
			Status: model.AegisUserStatus(st),
			PasswordHash: hash,
			RegisterTime: dt.Unix(),
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) SearchForNamespace(k string, pageNum int, pageSize int) (map[string]*model.Namespace, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	pattern := db.ToSqlSearchPattern(k)
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_status, ns_acl
FROM %s_namespace
WHERE ns_name LIKE $1 ESCAPE $2 OR ns_title LIKE $1 ESCAPE $2
ORDER BY ns_absid ASC LIMIT $3 OFFSET $4
`, pfx), pattern, "%", pageSize, pageNum*pageSize)
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make(map[string]*model.Namespace, 0)
	for stmt.Next() {
		var name, title, desc, email, owner, acl string
		var regtime time.Time
		var status int64
		err = stmt.Scan(&name, &title, &desc, &email, &owner, &regtime, &status, &acl)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res[name] = &model.Namespace{
			Name: name,
			Title: title,
			Description: desc,
			Email: email,
			Owner: owner,
			RegisterTime: regtime.Unix(),
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		}
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) SearchForRepository(k string, pageNum int, pageSize int) ([]*model.Repository, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	pattern := db.ToSqlSearchPattern(k)
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
WHERE repo_namespace LIKE $1 ESCAPE $2 OR repo_name LIKE $1 ESCAPE $2
ORDER BY repo_absid ASC LIMIT $3 OFFSET $4
`, pfx), pattern, "%", pageSize, pageNum*pageSize)
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make([]*model.Repository, 0)
	for stmt.Next() {
		var rType, rStatus int
		var repoNs, repoName, desc, owner, acl, forkOriginNs, forkOriginName string
		err := stmt.Scan(&rType, &repoNs, &repoName, &desc, &owner, &acl, &rStatus, &forkOriginNs, &forkOriginName)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, repoNs, repoName)
		m, err := model.CreateLocalRepository(uint8(rType), repoNs, repoName, p)
		if err != nil { return nil, err }
		res = append(res, &model.Repository{
			Namespace: repoNs,
			Name: repoName,
			Description: desc,
			Owner: owner,
			AccessControlList: a,
			Status: model.AegisRepositoryStatus(rStatus),
			Type: uint8(rType),
			ForkOriginNamespace: forkOriginNs,
			ForkOriginName: forkOriginName,
			Repository: m,
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) SetNamespaceACL(nsName string, targetUserName string, acl *model.ACLTuple) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	if acl == nil {
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_namespace SET ns_acl = ns_acl #- $1 WHERE ns_name = $2
`, pfx), []string{"acl", targetUserName}, nsName)
	} else {
		var r string
		r, err = acl.SerializeACLTuple()
		if err != nil { return err }
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_namespace SET ns_acl = jsonb_set(ns_acl, $1, $2) WHERE ns_name = $3
`, pfx), []string{"acl", targetUserName}, r, nsName)
	}
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) SetRepositoryACL(nsName string, repoName string, targetUserName string, acl *model.ACLTuple) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	if acl == nil {
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_repository SET repo_acl = repo_acl #- $1 WHERE repo_namespace = $2 AND repo_name = $3
`, pfx), []string{"acl", targetUserName}, nsName, repoName)
	} else {
		r, err := acl.SerializeACLTuple()
		if err != nil { return err }
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_repository SET repo_acl = jsonb_set(repo_acl, $1, $2) WHERE repo_namespace = $3 AND repo_name = $4
`, pfx), []string{"acl", targetUserName}, r, nsName, repoName)
	}
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllComprisingNamespace(username string) (map[string]*model.Namespace, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status
FROM %s_namespace
WHERE ns_owner = $1 OR ns_acl->'ACL' ? $1
`, pfx), username)
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make(map[string]*model.Namespace, 0)
	var name, title, description, email, owner, acl string
	var datetime time.Time
	var status int
	for stmt.Next() {
		err = stmt.Scan(&name, &title, &description, &email, &owner, &datetime, &acl, &status)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res[name] = &model.Namespace{
			Name: name,
			Title: title,
			Description: description,
			Email: email,
			Owner: owner,
			RegisterTime: datetime.Unix(),
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		}
	}
	return  res, nil
}

func (dbif *PostgresAegisDatabaseInterface) CountAllVisibleNamespaceSearchResult(username string, pattern string) (int64, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	var err error
	var stmt pgx.Row
	if len(pattern) > 0 {
		pat := db.ToSqlSearchPattern(pattern)
		if len(username) > 0 {
			stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_namespace
WHERE (ns_owner = $1 OR ns_acl->'ACL' ? $1 OR ns_status = 1)
AND (ns_name LIKE $2 ESCAPE $3 OR ns_title LIKE $2 ESCAPE $3)
`, pfx), username, pat, "%")
		} else {
			stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_namespace
WHERE (ns_status = 1) AND (ns_name LIKE $1 ESCAPE $2 OR ns_title LIKE $1 ESCAPE $2)
`, pfx), pat, "%")
		}
	} else {
		if len(username) > 0 {
			stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_namespace
WHERE (ns_owner = $1 OR ns_acl->'ACL' ? $1) OR ns_status = 1
`, pfx), username)
		} else {
			stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_namespace
WHERE ns_status = 1
`, pfx))
		}
	}
	var res int64
	err = stmt.Scan(&res)
	if err != nil { return 0, err }
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) CountAllVisibleRepositoriesSearchResult(username string, pattern string) (int64, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	var err error
	var stmt pgx.Row
	if len(pattern) > 0 {
		pat := db.ToSqlSearchPattern(pattern)
		if len(username) > 0 {
			stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*)
FROM %s_namespace
INNER JOIN (
    SELECT ns_name FROM %s_namespace
    WHERE (ns_status = 1) OR (ns_owner = $1 OR ns_acl->'ACL' ? $1 OR ns_status = 1)
) a ON %s_repository.repo_namespace = a.ns_name
WHERE (repo_owner = $1 OR repo_acl->'ACL' ? $1 OR repo_status = 1 OR repo_status = 4)
AND (repo_name LIKE $2 ESCAPE $3)
`, pfx, pfx, pfx), username, pat, "%")
		} else {
			stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*)
FROM %s_namespace
INNER JOIN (
    SELECT ns_name FROM %s_namespace
    WHERE (ns_status = 1) OR (ns_owner = $1 OR ns_acl->'ACL' ? $1 OR ns_status = 1)
) a ON %s_repository.repo_namespace = a.ns_name
WHERE (repo_status = 1 OR repo_status = 4)
AND (repo_name LIKE $1 ESCAPE $2)
`, pfx, pfx, pfx), pat, "%")
		}
	} else {
		if len(username) > 0 {
			stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*)
FROM %s_namespace
INNER JOIN (
    SELECT ns_name FROM %s_namespace
    WHERE (ns_status = 1) OR (ns_owner = $1 OR ns_acl->'ACL' ? $1 OR ns_status = 1)
) a ON %s_repository.repo_namespace = a.ns_name
WHERE (repo_owner = $1 OR repo_acl->'ACL' ? $1 OR repo_status = 1 OR repo_status = 4)
`, pfx, pfx, pfx), username)
		} else {
			stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*)
FROM %s_namespace
INNER JOIN (
    SELECT ns_name FROM %s_namespace
    WHERE (ns_status = 1)
) a ON %s_repository.repo_namespace = a.ns_name
WHERE (repo_status = 1 OR repo_status = 4)
`, pfx, pfx, pfx))
		}
	}
	var res int64
	err = stmt.Scan(&res)
	if err != nil { return 0, err }
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllRepositoryIssue(ns string, name string) ([]*model.Issue, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT issue_absid, issue_id, issue_timestamp, issue_author, issue_title, issue_content, issue_status
FROM %s_issue
WHERE repo_namespace = $1 AND repo_name = $2
`, pfx), ns, name)
	if err != nil { return nil, err }
	res := make([]*model.Issue, 0)
	var absid, id, status int64
	var t time.Time
	var author, title, content string
	for stmt.Next() {
		err = stmt.Scan(&absid, &ns, &name, &id, &t, &author, &title, &content, &status)
		if err != nil { return nil, err }
		res = append(res, &model.Issue{
			IssueAbsId: absid,
			RepoNamespace: ns,
			RepoName: name,
			IssueId: int(id),
			IssueAuthor: author,
			IssueTitle: title,
			IssueContent: content,
			IssueTime: t.Unix(),
			IssueStatus: int(status),
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetRepositoryIssue(ns string, name string, iid int) (*model.Issue, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT issue_absid, issue_timestamp, issue_author, issue_title, issue_content, issue_status, issue_priority
FROM %s_issue
WHERE repo_namespace = $1 AND repo_name = $2 AND issue_id = $3
`, pfx), ns, name, iid)
	var absid, id, status int64
	var t time.Time
	var priority int
	var author, title, content string
	err := stmt.Scan(&absid, &t, &author, &title, &content, &status, &priority)
	if err == pgx.ErrNoRows { return nil, db.ErrEntityNotFound }
	if err != nil { return nil, err }
	return &model.Issue{
		IssueAbsId: absid,
		RepoNamespace: ns,
		RepoName: name,
		IssueId: int(id),
		IssueAuthor: author,
		IssueTitle: title,
		IssueContent: content,
		IssueTime: t.Unix(),
		IssueStatus: int(status),
		IssuePriority: priority,
	}, nil
}

func (dbif *PostgresAegisDatabaseInterface) CountAllRepositoryIssue(ns string, name string) (int, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*)
FROM %s_issue
WHERE repo_namespace = $1 AND repo_name = $2
`, pfx), ns, name)
	var res int
	err := stmt.Scan(&res)
	if err == pgx.ErrNoRows { return 0, db.ErrEntityNotFound }
	if err != nil { return 0, err }
	return 0, nil
}

// filterType: 0 - all, 1 - open, 2 - closed, 3 - solved, 4 - discarded
// when query = "" it looks for all issue.
func (dbif *PostgresAegisDatabaseInterface) CountIssue(query string, namespace string, name string, filterType int) (int, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	statusClause := ""
	switch filterType {
	case 0: statusClause = "TRUE"
    case 1: statusClause = "issue_status = 1"
	case 2: statusClause = "NOT (issue_status = 1)"
	case 3: statusClause = "issue_status = 2"
	case 4: statusClause = "issue_status = 3"
	}
	var stmt pgx.Row
	if len(query) > 0 {
		stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_issue
WHERE repo_namespace = $1 AND repo_name = $2
AND %s
AND (issue_title LIKE $3 ESCAPE $4)
`, pfx, statusClause), namespace, name, query, "%")
	} else {
		stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_issue
WHERE repo_namespace = $1 AND repo_name = $2
AND %s
`, pfx, statusClause), namespace, name)
	}
	var res int
	err := stmt.Scan(&res)
	if err == pgx.ErrNoRows { return 0, db.ErrEntityNotFound }
	if err != nil { return 0, err }
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) SearchIssuePaginated(query string, namespace string, name string, filterType int, pageNum int, pageSize int) ([]*model.Issue, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	statusClause := ""
	switch filterType {
	case 0: statusClause = "TRUE"
    case 1: statusClause = "issue_status = 1"
	case 2: statusClause = "NOT (issue_status = 1)"
	case 3: statusClause = "issue_status = 2"
	case 4: statusClause = "issue_status = 3"
	}
	var err error
	var stmt pgx.Rows
	if len(query) > 0 {
		pat := db.ToSqlSearchPattern(query)
		stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT issue_absid, issue_id, issue_timestamp, issue_author, issue_title, issue_content, issue_status, issue_priority
FROM %s_issue
WHERE repo_namespace = $1 AND repo_name = $2 AND %s AND (issue_title LIKE $3 ESCAPE $4)
ORDER BY issue_priority DESC, issue_timestamp DESC LIMIT $5 OFFSET $6
`, pfx, statusClause), namespace, name, pat, "\\", pageSize, pageNum*pageSize)
	} else {
		stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT issue_absid, issue_id, issue_timestamp, issue_author, issue_title, issue_content, issue_status, issue_priority
FROM %s_issue
WHERE repo_namespace = $1 AND repo_name = $2 AND %s
ORDER BY issue_priority DESC, issue_timestamp DESC LIMIT $3 OFFSET $4
`, pfx, statusClause), namespace, name, pageSize, pageNum*pageSize)
	}
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make([]*model.Issue, 0)
	var absid, id, status int64
	var priority int
	var t time.Time
	var author, title, content string
	for stmt.Next() {
		err = stmt.Scan(&absid, &id, &t, &author, &title, &content, &status, &priority)
		if err != nil { return nil, err }
		res = append(res, &model.Issue{
			IssueAbsId: absid,
			RepoNamespace: namespace,
			RepoName: name,
			IssueId: int(id),
			IssueAuthor: author,
			IssueTitle: title,
			IssueContent: content,
			IssueTime: t.Unix(),
			IssueStatus: int(status),
			IssuePriority: priority,
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) NewRepositoryIssue(ns string, name string, author string, title string, content string) (int64, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt1 := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_issue WHERE repo_namespace = $1 AND repo_name = $2
`, pfx), ns, name)
	var newid int64
	err := stmt1.Scan(&newid)
	if err != nil { return 0, err }
	ctx2 := context.Background()
	tx, err := dbif.pool.Begin(ctx2)
	if err != nil { return 0, err }
	defer tx.Rollback(ctx2)
	t := time.Now()
	_, err = dbif.pool.Exec(ctx2, fmt.Sprintf(`
INSERT INTO %s_issue(repo_namespace, repo_name, issue_id, issue_timestamp, issue_author, issue_title, issue_content, issue_status, issue_priority)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`, pfx), ns, name, newid, t, author, title, content, model.ISSUE_OPENED, 0)
	if err != nil { return 0, err }
	err = tx.Commit(ctx2)
	if err != nil { return 0, err }
	return newid, nil
}

func (dbif *PostgresAegisDatabaseInterface) HardDeleteRepositoryIssue(ns string, name string, issueId int) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	_, err = tx.Exec(ctx, fmt.Sprintf(`
DELETE FROM %s_issue WHERE repo_namespace = $1 AND repo_name = $2 AND issue_id = $3
`, pfx), ns, name, issueId)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) SetIssuePriority(namespace string, name string, id int, priority int) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_issue SET issue_priority = $1 WHERE issue_id = $2 AND repo_namespace = $3 AND repo_name = $4
`, pfx), priority, id, namespace, name)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllIssueEvent(ns string, name string, issueId int) ([]*model.IssueEvent, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt1 := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT issue_absid FROM %s_issue WHERE repo_namespace = $1 AND repo_name = $2 AND issue_id = $3
`, pfx), ns, name, issueId)
	var absId int64
	err := stmt1.Scan(&absId);
	if err != nil { return nil, err }
	stmt2, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT issue_event_absid, issue_event_type, issue_event_time, issue_event_author, issue_event_content
FROM %s_issue_event WHERE issue_absid = $1
`, pfx), absId)
	if err != nil { return nil, err }
	defer stmt2.Close()
	var etype int
	var eid int64
	var time time.Time
	var author, content string
	res := make([]*model.IssueEvent, 0)
	for stmt2.Next() {
		err = stmt2.Scan(&eid, &etype, &time, &author, &content)
		if err != nil { return nil, err }
		res = append(res, &model.IssueEvent{
			EventAbsId: eid,
			EventType: etype,
			EventTimestamp: time.Unix(),
			EventAuthor: author,
			EventContent: content,
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) NewRepositoryIssueEvent(ns string, name string, issueId int, eType int, author string, content string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt1 := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT issue_absid, issue_status FROM %s_issue WHERE repo_namespace = $1 AND repo_name = $2 AND issue_id = $3
`, pfx), ns, name, issueId)
	var absId int64
	var issueStatus int
	err := stmt1.Scan(&absId, &issueStatus);
	if err != nil { return err }
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s_issue_event(issue_absid, issue_event_type, issue_event_time, issue_event_author, issue_event_content)
VALUES ($1, $2, $3, $4, $5)
`, pfx), absId, eType, time.Now(), author, content)
	if err != nil { return err }
	newIssueStatus := issueStatus
	switch eType {
	case model.EVENT_CLOSED_AS_SOLVED:
		newIssueStatus = model.ISSUE_CLOSED_AS_SOLVED
	case model.EVENT_CLOSED_AS_DISCARDED:
		newIssueStatus = model.ISSUE_CLOSED_AS_DISCARDED
	case model.EVENT_REOPENED:
		newIssueStatus = model.ISSUE_OPENED
	}
	if newIssueStatus != issueStatus {
		_, err := tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_issue SET issue_status = $1 WHERE issue_absid = $2
`, pfx), newIssueStatus, absId)
		if err != nil { return err }
	}
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) HardDeleteRepositoryIssueEvent(eventAbsId int64) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
DELETE FROM %s_issue_event WHERE issue_event_absid = $1
`, pfx), eventAbsId)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllBelongingNamespace(viewingUser string, user string) ([]*model.Namespace, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	var stmt pgx.Rows
	var err error
	if len(viewingUser) > 0 {
		if viewingUser == user {
			stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status
FROM %s_namespace WHERE ns_owner = $1 OR ns_acl->'ACL' ? $1
`, pfx), viewingUser)
			if err != nil { return nil, err }
		} else {
			stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status
FROM %s_namespace WHERE (ns_owner = $1 OR ns_acl->'ACL' = $1) AND (ns_owner = $2 OR ns_acl->'ACL' ? $2)
`, pfx), viewingUser, user)
			if err != nil { return nil, err }
		}
	} else {
		stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT ns_name, ns_title, ns_description, ns_email, ns_owner, ns_reg_datetime, ns_acl, ns_status
FROM %s_namespace WHERE (ns_status = 1) AND (ns_owner = $1 OR ns_acl->'ACL' ? $1)
`, pfx), user)
		if err != nil { return nil, err }
	}
	res := make([]*model.Namespace, 0)
	var name, title, description, email, owner, acl string
	var datetime time.Time
	var status int
	for stmt.Next() {
		err = stmt.Scan(&name, &title, &description, &email, &owner, &datetime, &acl, &status)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		res = append(res, &model.Namespace{
			Name: name,
			Title: title,
			Description: description,
			Email: email,
			Owner: owner,
			RegisterTime: datetime.Unix(),
			ACL: a,
			Status: model.AegisNamespaceStatus(status),
		})
	}
	return  res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllBelongingRepository(viewingUser string, user string, pageNum int, pageSize int) ([]*model.Repository, error) {
	// TODO: this logic might be wrong. we'll fix this later.
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	var stmt pgx.Rows
	var err error
	if len(viewingUser) > 0 {
		if viewingUser == user {
			stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
WHERE repo_owner = $1 OR repo_acl->'ACL' ? $1
ORDER BY repo_absid ASC LIMIT $2 OFFSET $3
`, pfx), viewingUser, pageSize, pageNum*pageSize)
			if err != nil { return nil, err }
		} else {
			stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
WHERE (repo_status = 1 OR repo_status = 4 OR repo_owner = $1 OR repo_acl->'ACL' ? $1) AND (repo_owner = $2 OR repo_acl -> 'ACL' ? $2)
ORDER BY repo_absid ASC LIMIT $3 OFFSET $4
`, pfx), viewingUser, user, pageSize, pageNum*pageSize)
			if err != nil { return nil, err }
		}
	} else {
		stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_namespace, repo_name, repo_description, repo_owner, repo_acl, repo_status, repo_fork_origin_namespace, repo_fork_origin_name
FROM %s_repository
WHERE (repo_status = 1 OR repo_status = 4) AND (repo_owner = $1 OR repo_acl->'ACL' ? $1)
ORDER BY repo_absid ASC LIMIT $2 OFFSET $3
`, pfx), user, pageSize, pageNum*pageSize)
		if err != nil { return nil, err }
	}
	res := make([]*model.Repository, 0)
	for stmt.Next() {
		var ns, name, desc, acl, owner, forkOriginNamespace, forkOriginName string
		var status int64
		var repoType uint8
		err := stmt.Scan(&repoType, &ns, &name, &desc, &owner, &acl, &status, &forkOriginNamespace, &forkOriginName)
		if err != nil { return nil, err }
		a, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		p := path.Join(dbif.config.GitRoot, ns, name)
		lr, err := model.CreateLocalRepository(repoType, ns, name, p)
		if err != nil { return nil, err }
		res = append(res, &model.Repository{
			Type: repoType,
			Namespace: ns,
			Name: name,
			Description: desc,
			Owner: owner,
			AccessControlList: a,
			Status: model.AegisRepositoryStatus(status),
			Repository: lr,
			ForkOriginNamespace: forkOriginNamespace,
			ForkOriginName: forkOriginName,
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetForkRepositoryOfUser(username string, originNamespace string, originName string) ([]*model.Repository, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT repo_type, repo_namespace, repo_name, repo_description, repo_acl, repo_status
FROM %s_repository
WHERE repo_owner = $1 AND repo_fork_origin_namespace = $2 AND repo_fork_origin_name = $3
`, pfx), username, originNamespace, originName)
	if err != nil { return nil, err }
	defer stmt.Close()
	var ns, name, desc, acl string
	var status int
	var repoType uint8
	res := make([]*model.Repository, 0)
	for stmt.Next() {
		err = stmt.Scan(&repoType, &ns, &name, &desc, &acl, &status)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		p := path.Join(dbif.config.GitRoot, ns, name)
		lr, err := model.CreateLocalRepository(repoType, ns, name, p)
		if err != nil { return nil, err }
		mr, err := model.NewRepository(ns, name, lr)
		mr.Owner = username
		mr.Type = repoType
		mr.Status = model.AegisRepositoryStatus(status)
		mr.ForkOriginNamespace = originNamespace
		mr.ForkOriginName = originName
		aclobj, err := model.ParseACL(acl)
		if err != nil { return nil, err }
		mr.AccessControlList = aclobj
		res = append(res, mr)
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllPullRequestPaginated(namespace string, name string, pageNum int, pageSize int) ([]*model.PullRequest, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT pull_request_absid, pull_request_id, author_username, title, receiver_branch, provider_namespace, provider_name, provider_branch, merge_conflict_check_result, merge_conflict_check_timestamp, pull_request_status, pull_request_timestamp
FROM %s_pull_request
WHERE receiver_namespace = $1 AND receiver_name = $2
ORDER BY pull_request_id ASC LIMIT $3 OFFSET $4
`, pfx), namespace, name, pageSize, pageNum*pageSize)
	if err != nil { return nil, err }
	res := make([]*model.PullRequest, 0)
	var absid, id int64
	var author, title, receiverBranch, providerNs, providerName, providerBranch, mergeCheckResultString string
	var status int
	var mergeCheckTime, pullRequestTime time.Time
	for stmt.Next() {
		err = stmt.Scan(&absid, &id, &author, &title, &receiverBranch, &providerNs, &providerName, &mergeCheckResultString, &mergeCheckTime, &status, &pullRequestTime)
		if err != nil { return nil, err }
		var mergeCheckResult *gitlib.MergeCheckResult = nil
		if len(mergeCheckResultString) > 0 {		
			err = json.Unmarshal([]byte(mergeCheckResultString), &mergeCheckResult)
			if err != nil { return nil, err }
		}
		res = append(res, &model.PullRequest{
			PRId: id,
			PRAbsId: absid,
			Title: title,
			Author: author,
			Timestamp: pullRequestTime.Unix(),
			ReceiverNamespace: namespace,
			ReceiverName: name,
			ReceiverBranch: receiverBranch,
			ProviderNamespace: providerNs,
			ProviderName: providerName,
			ProviderBranch: providerBranch,
			Status: status,
			MergeCheckResult: mergeCheckResult,
			MergeCheckTimestamp: mergeCheckTime.Unix(),
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) NewPullRequest(username string, title string, receiverNamespace string, receiverName string, receiverBranch string, providerNamespace string, providerName string, providerBranch string) (int64, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt1 := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_pull_request WHERE receiver_namespace = $1 AND receiver_name = $2
`, pfx), receiverNamespace, receiverName)
	var newId int64
	err := stmt1.Scan(&newId)
	if err != nil { return 0, err }
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return 0, err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s_pull_request(author_username, pull_request_id, title, receiver_namespace, receiver_name, receiver_branch, provider_namespace, provider_name, provider_branch, merge_conflict_check_result, merge_conflict_check_timestamp, pull_request_status, pull_requst_timestep)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
`, pfx), username, newId, title, receiverNamespace, receiverName, receiverBranch, providerNamespace, providerName, providerBranch, new(string), new(time.Time), model.PULL_REQUEST_OPEN, time.Now())
	if err != nil { return 0, err }
	err = tx.Commit(ctx)
	if err != nil { return 0, err }
	return newId, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetPullRequest(namespace string, name string, id int64) (*model.PullRequest, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT pull_request_absid, author_username, title, receiver_branch, provider_namespace, provider_name, provider_branch, merge_conflict_check_result, merge_conflict_check_timestamp, pull_request_status, pull_request_timestamp
FROM %s_pull_request
WHERE receiver_namespace = $1 AND receiver_name = $2 AND pull_request_id = $3
`, pfx), namespace, name, id)
	var absid int64
	var author, title, receiverBranch, providerNs, providerName, providerBranch, mergeCheckString string
	var mergeConflictTime, pullRequestTime time.Time
	var status int
	err := stmt.Scan(&absid, &author, &title, &receiverBranch, &providerNs, &providerName, &providerBranch, &mergeCheckString, &mergeConflictTime, &status, &pullRequestTime)
	if err != nil { return nil, err }
	var mergeCheckResult *gitlib.MergeCheckResult = nil
	if len(mergeCheckString) > 0 {		
		err = json.Unmarshal([]byte(mergeCheckString), &mergeCheckResult)
		if err != nil { return nil, err }
	}
	return &model.PullRequest{
		PRId: id,
		PRAbsId: absid,
		Title: title,
		Author: author,
		Timestamp: pullRequestTime.Unix(),
		ReceiverNamespace: namespace,
		ReceiverName: name,
		ReceiverBranch: receiverBranch,
		ProviderNamespace: providerNs,
		ProviderName: providerName,
		ProviderBranch: providerBranch,
		Status: status,
		MergeCheckResult: mergeCheckResult,
		MergeCheckTimestamp: mergeConflictTime.Unix(),
	}, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetPullRequestByAbsId(absId int64) (*model.PullRequest, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT author_username, pull_request_id, title, receiver_anmespace, receiver_name, receiver_branch, provider_namespace, provider_name, provider_branch, merge_conflict_check_result, merge_conflict_check_timestamp, pull_request_status, pull_request_timestamp
FROM %s_pull_request
WHERE receiver_namespace = $1 AND receiver_name = $2 AND pull_request_id = $3
`, pfx), absId)
	var id int64
	var author, title, receiverNs, receiverName, receiverBranch, providerNs, providerName, providerBranch, mergeCheckString string
	var mergeConflictTime, pullRequestTime time.Time
	var status int
	err := stmt.Scan(&author, &id, &title, &receiverNs, &receiverName, &receiverBranch, &providerNs, &providerName, &providerBranch, &mergeCheckString, &mergeConflictTime, &status, &pullRequestTime)
	if err != nil { return nil, err }
	var mergeCheckResult *gitlib.MergeCheckResult = nil
	if len(mergeCheckString) > 0 {		
		err = json.Unmarshal([]byte(mergeCheckString), &mergeCheckResult)
		if err != nil { return nil, err }
	}
	return &model.PullRequest{
		PRId: id,
		PRAbsId: absId,
		Title: title,
		Author: author,
		Timestamp: pullRequestTime.Unix(),
		ReceiverNamespace: receiverNs,
		ReceiverName: receiverName,
		ReceiverBranch: receiverBranch,
		ProviderNamespace: providerNs,
		ProviderName: providerName,
		ProviderBranch: providerBranch,
		Status: status,
		MergeCheckResult: mergeCheckResult,
		MergeCheckTimestamp: mergeConflictTime.Unix(),
	}, nil
}

func (dbif *PostgresAegisDatabaseInterface) CheckPullRequestMergeConflict(absId int64) (*gitlib.MergeCheckResult, error) {
	// WARNING: currently only works when when the source &
	// the target is git repo. currently (2025.8.27) this check
	// is performed at the controller side, i.e. users cannot
	// create pull request if the repo is not git repo, but the
	// code can still be called. DO NOT CALL UNLESS YOU KNOW
	// WHAT YOU'RE DOING.
	// TODO: fix this after figuring things out.
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT receiver_namespace, receiver_name, receiver_branch, provider_namespace, provider_name, provider_branch
FROM %s_pull_request
WHERE pull_request_absid = $1
`, pfx), absId)
	var receiverNamespace, receiverName, receiverBranch string
	var providerNamespace, providerName, providerBranch string
	err := stmt.Scan(&receiverNamespace, &receiverName, &receiverBranch, &providerNamespace, &providerName, &providerBranch)
	if err != nil { return nil, err }
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return nil, err }
	defer tx.Rollback(ctx)
	p := path.Join(dbif.config.GitRoot, receiverNamespace, receiverName)
	lgr := gitlib.NewLocalGitRepository(receiverNamespace, receiverName, p)
	remoteName := fmt.Sprintf("%s/%s", providerNamespace, providerName)
	mr, err := lgr.CheckBranchMergeConflict(receiverBranch, remoteName, providerBranch)
	if err != nil { return nil, err }
	mrstr, err := json.Marshal(mr)
	if err != nil { return nil, err }
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_pull_request
SET merge_conflict_check_result = $2, merge_conflict_check_timestamp = $3
WHERE pull_request_absid = $1
`, pfx), absId, string(mrstr), time.Now())
	if err != nil { return nil, err }
	err = tx.Commit(ctx)
	if err != nil { return nil, err }
	return mr, nil
}

func (dbif *PostgresAegisDatabaseInterface) DeletePullRequest(absId int64) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
DELETE FROM %s_pull_request WHERE pull_request_absid = $1
`, pfx), absId)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllPullRequestEventPaginated(absId int64, pageNum int, pageSize int) ([]*model.PullRequestEvent, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT event_type, event_timestamp, event_author, event_content
FROM %s_pull_request_event
WHERE pull_request_absid = $1
ORDER BY event_timestamp ASC LIMIT $2 OFFSET $3
`, pfx), absId, pageSize, pageNum*pageSize)
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make([]*model.PullRequestEvent, 0)
	var etype int
	var timestamp time.Time
	var author, content string
	for stmt.Next() {
		err = stmt.Scan(&etype, &timestamp, &author, &content)
		if err != nil { return nil, err }
		res = append(res, &model.PullRequestEvent{
			PRAbsId: absId,
			EventType: etype,
			EventTimestamp: timestamp.Unix(),
			EventAuthor: author,
			EventContent: content,
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) CheckAndMergePullRequest(absId int64, username string) error {
	// WARNING: currently only works when when the source &
	// the target is git repo. currently (2025.8.27) this check
	// is performed at the controller side, i.e. users cannot
	// create pull request if the repo is not git repo, but the
	// code can still be called. DO NOT CALL UNLESS YOU KNOW
	// WHAT YOU'RE DOING.
	// TODO: fix this after figuring things out.
	r, err := dbif.CheckPullRequestMergeConflict(absId)
	if err != nil { return err }
	// TODO: this would need to be fixed in the future...
	if !r.Successful { return nil }
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt0 := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT user_email, user_title FROM %s_user WHERE user_name = $1
`, pfx), username)
	var email, userTitle string
	err = stmt0.Scan(&email, &userTitle)
	if err != nil { return err }
	// fetch
	buf := new(bytes.Buffer)
	cmd1 := exec.Command("git", "fetch", r.ProviderRemoteName, r.ProviderBranch)
	cmd1.Dir = r.ReceiverLocation
	cmd1.Stderr = buf
	err = cmd1.Run()
	if err != nil { return errors.New(err.Error() + ": " + buf.String()) }
	buf.Reset()
	providerFullName := fmt.Sprintf("%s/%s", r.ProviderRemoteName, r.ProviderBranch)
	cmd2 := exec.Command("git", "merge-tree", "--write-tree", r.ReceiverBranch, providerFullName)
	cmd2.Dir = r.ReceiverLocation
	cmd2.Stdout = buf
	err = cmd2.Run()
	if err != nil { return fmt.Errorf("Failed while merge-tree: %s", err.Error()) }
	treeId := strings.TrimSpace(buf.String())
	mergeMessage := fmt.Sprintf("merge: from %s/%s to %s", r.ProviderRemoteName, r.ProviderBranch, r.ReceiverBranch)
	buf.Reset()
	cmd3 := exec.Command("git", "commit-tree", treeId, "-m", mergeMessage, "-p", r.ReceiverBranch, "-p", providerFullName)
	cmd3.Dir = r.ReceiverLocation
	cmd3.Stdout = buf
	cmd3.Env = os.Environ()
	cmd3.Env = append(cmd3.Env, fmt.Sprintf("GIT_AUTHOR_NAME=%s", userTitle))
	cmd3.Env = append(cmd3.Env, fmt.Sprintf("GIT_AUTHOR_EMAIL=%s", email))
	cmd3.Env = append(cmd3.Env, fmt.Sprintf("GIT_COMMITTER_NAME=%s", userTitle))
	cmd3.Env = append(cmd3.Env, fmt.Sprintf("GIT_COMMITTER_EMAIL=%s", email))
	err = cmd3.Run()
	if err != nil { return fmt.Errorf("Failed while commit-tree: %s", err.Error()) }
	commitId := strings.TrimSpace(buf.String())
	buf.Reset()
	receiverBranchFullName := fmt.Sprintf("refs/heads/%s", r.ReceiverBranch)
	cmd4 := exec.Command("git", "update-ref", receiverBranchFullName, commitId)
	cmd4.Dir = r.ReceiverLocation
	cmd4.Stderr = buf
	err = cmd4.Run()
	if err != nil { return fmt.Errorf("Failed while update-ref: %s; %s", err.Error(), buf.String()) }
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	t := time.Now()
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_pull_request SET pull_request_status = $1, pull_request_timestamp = $2 WHERE pull_request_absid = $3
`, pfx), model.PULL_REQUEST_CLOSED_AS_MERGED, t, absId)
	if err != nil { return err }
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s_pull_request_event(pull_request_abs_id, event_type, event_timestamp, event_author, event_content)
VALUES ($1,$2,$3,$4,$5)
`, pfx), absId, model.PULL_REQUEST_EVENT_CLOSE_AS_MERGED, t, username, "")
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) CommentOnPullRequest(absId int64, author string, content string) (*model.PullRequestEvent, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return nil, err }
	defer tx.Rollback(ctx)
	t := time.Now().Unix()
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s_pull_request_event(pull_request_absid, event_type, event_timestamp, event_author, event_content) VALUES ($1,$2,$3,$4,$5)
`, pfx), absId, model.PULL_REQUEST_EVENT_COMMENT, t, author, content)
	if err != nil { return nil, err }
	err = tx.Commit(ctx)
	if err != nil { return nil, err }
	return &model.PullRequestEvent{
		PRAbsId: absId,
		EventType: model.PULL_REQUEST_EVENT_COMMENT,
		EventTimestamp: t,
		EventAuthor: author,
		EventContent: content,
	}, nil
}

func (dbif *PostgresAegisDatabaseInterface) CommentOnPullRequestCode(absId int64, comment *model.PullRequestCommentOnCode) (*model.PullRequestEvent, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return nil, err }
	defer tx.Rollback(ctx)
	t := time.Now().Unix()
	contentBytes, _ := json.Marshal(comment)
	contentString := string(contentBytes)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s_pull_request_event(pull_request_absid, event_type, event_timestamp, event_author, event_content)
VALUES ($1,$2,$3,$4,$5)
`, pfx), absId, model.PULL_REQUEST_EVENT_COMMENT_ON_CODE, t, comment.Username, contentString)
	if err != nil { return nil, err }
	err = tx.Commit(ctx)
	if err != nil { return nil, err }
	return &model.PullRequestEvent{
		PRAbsId: absId,
		EventType: model.PULL_REQUEST_EVENT_COMMENT_ON_CODE,
		EventTimestamp: t,
		EventAuthor: comment.Username,
		EventContent: contentString,
	}, nil
}

func (dbif *PostgresAegisDatabaseInterface) ClosePullRequestAsNotMerged(absid int64, author string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	t := time.Now()
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s_pull_request_event(pull_request_absid, event_type, event_timestamp, event_author, event_content)
VALUES ($1,$2,$3,$4,$5)
`, pfx), absid, model.PULL_REQUEST_EVENT_CLOSE_AS_NOT_MERGED, t, author, new(string))
	if err != nil { return err }
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_pull_request
SET pull_request_status = $1
WHERE pull_request_absid = $2
`, pfx), model.PULL_REQUEST_CLOSED_AS_NOT_MERGED, absid)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) ReopenPullRequest(absid int64, author string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	t := time.Now()
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s_pull_request_event(pull_request_absid, event_type, event_timestamp, event_author, event_content)
VALUES ($1,$2,$3,$4,$5)
`, pfx), absid, model.PULL_REQUEST_EVENT_REOPEN, t, author, new(string))
	if err != nil { return err }
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_pull_request
SET pull_request_status = $1
WHERE pull_request_absid = $2
`, pfx), model.PULL_REQUEST_OPEN, absid)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) CountPullRequest(query string, namespace string, name string, filterType int) (int, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	statusClause := ""
	switch filterType {
	case 0: statusClause = ""
	case 1: statusClause = "AND pull_request_status = 1"
	case 2: statusClause = "AND NOT (pull_request_status = 1)"
	}
	var stmt pgx.Row
	if len(query) > 0 {
		pat := db.ToSqlSearchPattern(query)
		stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_pull_request
WHERE receiver_namespace = $1 AND receiver_name = $2 %s AND title LIKE $3 ESCAPE $4
`, pfx, statusClause), namespace, name, pat, "\\")
	} else {
		stmt = dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT COUNT(*) FROM %s_pull_request
WHERE receiver_namespace = $1 AND receiver_name = $2 %s
`, pfx, statusClause), namespace, name)
	}
	var res int
	err := stmt.Scan(&res)
	if err != nil { return 0, err }
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) SearchPullRequestPaginated(query string, namespace string, name string, filterType int, pageNum int, pageSize int) ([]*model.PullRequest, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	statusClause := ""
	switch filterType {
	case 0: statusClause = ""
	case 1: statusClause = "AND pull_request_status = 1"
	case 2: statusClause = "AND NOT (pull_request_status = 1)"
	case 3: statusClause = "AND pull_request_status = 2"
	case 4: statusClause = "AND pull_request_status = 3"
	}
	var stmt pgx.Rows
	var err error
	if len(query) >= 0 {
		pat := db.ToSqlSearchPattern(query)
		stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT rowid, username, pull_request_id, title, receiver_branch, provider_namespace, provider_name, provider_branch, merge_conflict_check_result, merge_conflict_check_timestamp, pull_request_status, pull_request_timestamp
FROM %spull_request
WHERE receiver_namespace = ? AND receiver_name = ? %s AND title LIKE $3 ESCAPE $4
ORDER BY pull_request_timestamp DESC LIMIT $1 OFFSET $2
`, pfx, statusClause), pageSize, pageNum*pageSize, pat, "\\")
		if err != nil { return nil, err }
	} else {
		stmt, err = dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT rowid, username, pull_request_id, title, receiver_branch, provider_namespace, provider_name, provider_branch, merge_conflict_check_result, merge_conflict_check_timestamp, pull_request_status, pull_request_timestamp
FROM %spull_request
WHERE receiver_namespace = ? AND receiver_name = ? %s
ORDER BY pull_request_timestamp DESC LIMIT $1 OFFSET $2
`, pfx, statusClause), pageSize, pageNum*pageSize)
		if err != nil { return nil, err }
	}
	res := make([]*model.PullRequest, 0)
	var prid, absid int64
	var prtime, mergeCheckTimestamp time.Time
	var status int
	var username, title, receiverBranch string
	var providerNamespace, providerName, provideBranch string
	var mergeCheckResultString string
	for stmt.Next() {
		err = stmt.Scan(&absid, &username, &prid, &title, &receiverBranch, &providerNamespace, &providerName, &provideBranch, &mergeCheckResultString, &mergeCheckTimestamp, &status, &prtime)
		if err != nil { return nil, err }
		var mergeCheckResult *gitlib.MergeCheckResult = nil
		if len(mergeCheckResultString) > 0 {		
			err = json.Unmarshal([]byte(mergeCheckResultString), &mergeCheckResult)
			if err != nil { return nil, err }
		}
		res = append(res, &model.PullRequest{
			PRId: prid,
			PRAbsId: absid,
			Title: title,
			Author: username,
			Timestamp: prtime.Unix(),
			ReceiverNamespace: namespace,
			ReceiverName: name,
			ReceiverBranch: receiverBranch,
			ProviderNamespace: providerNamespace,
			ProviderName: providerName,
			ProviderBranch: provideBranch,
			Status: status,
			MergeCheckResult: mergeCheckResult,
			MergeCheckTimestamp: mergeCheckTimestamp.Unix(),
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) GetAllRegisteredEmailOfUser(username string) ([]struct{Email string;Verified bool}, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT email, verified FROM %s_user_email WHERE username = $1
`, pfx), username)
	if err != nil { return nil, err }
	defer stmt.Close()
	res := make([]struct{Email string;Verified bool}, 0)
	var email string
	var verified int
	for stmt.Next() {
		err = stmt.Scan(&email, &verified)
		if err != nil { return nil, err }
		res = append(res, struct{Email string;Verified bool}{
			Email: email,
			Verified: verified == 1,
		})
	}
	return res, nil
}

func (dbif *PostgresAegisDatabaseInterface) AddEmail(username string, email string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s_user_email(username, email, verified) VALUES ($1, $2, 0)
`, pfx), username, email)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) VerifyRegisteredEmail(username string, email string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s_user_email SET verified = 1 WHERE username = $1 AND email = $2
`, pfx), username, email)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) DeleteRegisteredEmail(username string, email string) error {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	tx, err := dbif.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
DELETE FROM %s_user_email WHERE username = $1 AND email = $2
`, pfx), username, email)
	if err != nil { return err }
	err = tx.Commit(ctx)
	if err != nil { return err }
	return nil
}

func (dbif *PostgresAegisDatabaseInterface) CheckIfEmailVerified(username string, email string) (bool, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT verified FROM %s_user_email WHERE username = $1 AND email = $2
`, pfx), username, email)
	var r int
	err := stmt.Scan(&r)
	if err != nil { return false, err }
	return r == 1, nil
}

func (dbif *PostgresAegisDatabaseInterface) ResolveEmailToUsername(email string) (string, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	stmt := dbif.pool.QueryRow(ctx, fmt.Sprintf(`
SELECT username FROM %s_user_email WHERE email = $2 AND verified = 1
`, pfx), email)
	var r string
	err := stmt.Scan(&r)
	if err != nil { return "", err }
	return r, nil
}

func (dbif *PostgresAegisDatabaseInterface) ResolveMultipleEmailToUsername(emailList map[string]string) (map[string]string, error) {
	pfx := dbif.config.Database.TablePrefix
	ctx := context.Background()
	l := make([]any, 0)
	q := make([]string, 0)
	i := 1
	for k := range emailList {
		l = append(l, k)
		q = append(q, fmt.Sprintf("$%d", i))
		i += 1
	}
	stmt, err := dbif.pool.Query(ctx, fmt.Sprintf(`
SELECT email, username FROM %s_user_email WHERE verified = 1 AND email IN (%s)
`, pfx, strings.Join(q, ",")), l...)
	if err != nil { return nil, err }
	defer stmt.Close()
	var email, username string
	for stmt.Next() {
		err = stmt.Scan(&email, &username)
		if err != nil { return nil, err }
		emailList[email] = username
	}
	return emailList, nil
}

