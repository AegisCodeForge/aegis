package db

import "github.com/bctnry/aegis/pkg/aegis/model"

type AegisDatabaseInterface interface {
	// we have to discern between "database unusable" and "error while detecting".
	IsDatabaseUsable() (bool, error)
	InstallTables() error
	GetUserByName(name string) (*model.AegisUser, error)
	GetAllAuthKeyByUsername(name string) ([]model.AegisAuthKey, error)
	GetAuthKeyByName(userName string, keyName string) (*model.AegisAuthKey, error)
	RegisterAuthKey(username string, keyname string, keytext string) error
	RemoveAuthKey(username string, keyname string) error
	GetAllSignKeyByUsername(name string) ([]model.AegisSigningKey, error)
	RegisterSignKey(username string, keyname string, keytext string) error
	RemoveSignKey(username string, keyname string) error
	GetNamespaceByName(name string) (*model.Namespace, error)
	GetAllNamespace() (map[string]*model.Namespace, error)
	GetAllNamespaceByOwner(name string) (map[string]*model.Namespace, error)
	GetAllRepositoryFromNamespace(name string) (map[string]*model.Repository, error)
	RegisterUser(name string, email string, passwordHash string, status model.AegisUserStatus) (*model.AegisUser, error)
	// update user info. NOTE THAT any implementers MUST update the
	// status field as well if that has changed.
	UpdateUserInfo(name string, uobj *model.AegisUser) error
	UpdateUserPassword(name string, newPasswordHash string) error
	// soft delete vs. hard delete:
	// "soft delete" mark the user as deleted.
	// "hard delete" removes the data from the database for good.
	HardDeleteUserByName(name string) error
	UpdateUserStatus(name string, newStatus model.AegisUserStatus) error
	RegisterNamespace(name string, ownerUsername string) (*model.Namespace, error)
	UpdateNamespaceInfo(name string, nsobj *model.Namespace) error
	UpdateNamespaceOwner(name string, newOwner string) error
	UpdateNamespaceStatus(name string, newStatus model.AegisNamespaceStatus) error
	// the implementer should remove the directory as well.
	HardDeleteNamespaceByName(name string) error
	CreateRepository(ns string, name string) (*model.Repository, error)
	UpdateRepositoryInfo(ns string, name string, robj *model.Repository) error
	UpdateRepositoryStatus(ns string, name string, status model.AegisRepositoryStatus) error
	HardDeleteRepository(ns string, name string) error
	MoveRepository(oldNs string, oldName string, newNs string, newName string) error

	GetAllUsers(pageNum int, pageSize int) ([]*model.AegisUser, error)
	GetAllNamespaces(pageNum int, pageSize int) ([]*model.Namespace, error)
	GetAllRepositories(pageNum int, pageSize int) ([]*model.Repository, error)

	CountAllUser() (int64, error)
	CountAllNamespace() (int64, error)
	CountAllRepositories() (int64, error)

	// search user name & title containing the string `k`, case
	// insensitive.
	SearchForUser(k string, pageNum int, pageSize int) ([]*model.AegisUser, error)

	// search namespce name & title containing the string `k`, case
	// insensitive.
	SearchForNamespace(k string, pageNum int, pageSize int) ([]*model.Namespace, error)

	// search repo namespace name & repository name & title containing
	// the string `k`, case insensitive.
	SearchForRepository(k string, pageNum int, pageSize int) ([]*model.Repository, error)
}

