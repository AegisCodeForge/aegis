package db

import "github.com/bctnry/gitus/pkg/gitus/model"

type GitusDatabaseInterface interface {
	// we have to discern between "database unusable" and "error while detecting".
	IsDatabaseUsable() (bool, error)
	InstallTables() error
	GetUserByName(name string) (*model.GitusUser, error)
	GetAllAuthKeyByUsername(name string) ([]model.GitusAuthKey, error)
	GetAuthKeyByName(userName string, keyName string) (*model.GitusAuthKey, error)
	RegisterAuthKey(username string, keyname string, keytext string) error
	RemoveAuthKey(username string, keyname string) error
	GetAllSignKeyByUsername(name string) ([]model.GitusSigningKey, error)
	RegisterSignKey(username string, keyname string, keytext string) error
	RemoveSignKey(username string, keyname string) error
	GetNamespaceByName(name string) (*model.Namespace, error)
	GetAllNamespace() (map[string]*model.Namespace, error)
	GetAllNamespaceByOwner(name string) (map[string]*model.Namespace, error)
	GetAllRepositoryFromNamespace(name string) (map[string]*model.Repository, error)
	RegisterUser(name string, email string, passwordHash string) (*model.GitusUser, error)
	UpdateUserInfo(name string, uobj *model.GitusUser) error
	UpdateUserPassword(name string, newPasswordHash string) error
	// soft delete vs. hard delete:
	// "soft delete" mark the user as deleted.
	// "hard delete" removes the data from the database for good.
	HardDeleteUserByName(name string) error
	UpdateUserStatus(name string, newStatus model.GitusUserStatus) error
	RegisterNamespace(name string, ownerUsername string) (*model.Namespace, error)
	UpdateNamespaceInfo(name string, nsobj *model.Namespace) error
	UpdateNamespaceOwner(name string, newOwner string) error
	UpdateNamespaceStatus(name string, newStatus model.GitusNamespaceStatus) error
	HardDeleteNamespaceByName(name string) error
	CreateRepository(ns string, name string) (*model.Repository, error)
	UpdateRepositoryInfo(ns string, name string, robj *model.Repository) error
	UpdateRepositoryStatus(ns string, name string, status model.GitusRepositoryStatus) error
	HardDeleteRepository(ns string, name string) error
	MoveRepository(oldNs string, oldName string, newNs string, newName string) error
	
}

