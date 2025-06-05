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
	GetRepositoryByName(nsName string, repoName string) (*model.Repository, error)
	GetAllNamespace() (map[string]*model.Namespace, error)
	GetAllVisibleNamespace(username string) (map[string]*model.Namespace, error)
	GetAllVisibleNamespacePaginated(username string, pageNum int, pageSize int) (map[string]*model.Namespace, error)
	SearchAllVisibleNamespacePaginated(username string, query string, pageNum int, pageSize int) (map[string]*model.Namespace, error)
	GetAllVisibleRepositoryPaginated(username string, pageNum int, pageSize int) ([]*model.Repository, error)
	SearchAllVisibleRepositoryPaginated(username string, query string, pageNum int, pageSize int) ([]*model.Repository, error)
	GetAllNamespaceByOwner(name string) (map[string]*model.Namespace, error)
	GetAllRepositoryFromNamespace(name string) (map[string]*model.Repository, error)
	GetAllVisibleRepositoryFromNamespace(username string, ns string) ([]*model.Repository, error)
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
	CreateRepository(ns string, name string, owner string) (*model.Repository, error)
	UpdateRepositoryInfo(ns string, name string, robj *model.Repository) error
	UpdateRepositoryStatus(ns string, name string, status model.AegisRepositoryStatus) error
	HardDeleteRepository(ns string, name string) error
	MoveRepository(oldNs string, oldName string, newNs string, newName string) error

	GetAllUsers(pageNum int, pageSize int) ([]*model.AegisUser, error)
	GetAllNamespaces(pageNum int, pageSize int) (map[string]*model.Namespace, error)
	GetAllRepositories(pageNum int, pageSize int) ([]*model.Repository, error)

	CountAllUser() (int64, error)
	CountAllNamespace() (int64, error)
	CountAllRepositories() (int64, error)
	CountAllRepositoriesSearchResult(q string) (int64, error)
	CountAllVisibleNamespace(username string) (int64, error)
	CountAllVisibleRepositories(username string) (int64, error)

	// search user name & title containing the string `k`, case
	// insensitive.
	SearchForUser(k string, pageNum int, pageSize int) ([]*model.AegisUser, error)

	// search namespce name & title containing the string `k`, case
	// insensitive.
	SearchForNamespace(k string, pageNum int, pageSize int) (map[string]*model.Namespace, error)

	// search repo namespace name & repository name & title containing
	// the string `k`, case insensitive.
	SearchForRepository(k string, pageNum int, pageSize int) ([]*model.Repository, error)

	// set ACL as specified.  implementer should remove permissions of
	// `targetUserName` when `acl` is nil,
	SetNamespaceACL(nsName string, targetUserName string, acl *model.ACLTuple) error
	SetRepositoryACL(nsName string, repoName string, targetUserName string, acl *model.ACLTuple) error

	GetAllComprisingNamespace(username string) (map[string]*model.Namespace, error)
	
	CountAllVisibleNamespaceSearchResult(username string, pattern string) (int64, error)
	CountAllVisibleRepositoriesSearchResult(username string, pattern string) (int64, error)

	GetAllRepositoryIssue(ns string, name string) ([]*model.Issue, error)
	GetRepositoryIssue(ns string, name string, iid int) (*model.Issue, error)
	CountAllRepositoryIssue(ns string, name string) (int, error)
	// returns the issue_id of the new issue.
	NewRepositoryIssue(ns string, name string, author string, title string, content string) (int64, error)
	HardDeleteRepositoryIssue(ns string, name string, issueId int) error
	GetAllIssueEvent(ns string, name string, issueId int) ([]*model.IssueEvent, error)
	NewRepositoryIssueEvent(ns string, name string, issueId int, eType int, author string, content string) error
	HardDeleteRepositoryIssueEvent(eventAbsId int64) error
	
	GetAllBelongingNamespace(viewingUser string, user string) ([]*model.Namespace, error)
	GetAllBelongingRepository(viewingUser string, user string, pageNum int, pageSize int) ([]*model.Repository, error)
}

// the fact that golang has no parameter default values is
// horrible. it's a simple concept, it's not hard to implement, and
// due to no default values one has to either make the interface
// bloated with functions for each cases or force the caller to endure
// a bloated function.

