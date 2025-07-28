package model

import (
	"fmt"

)

type AegisRepositoryStatus int

const (
	REPO_NORMAL_PUBLIC AegisRepositoryStatus = 1
	REPO_NORMAL_PRIVATE AegisRepositoryStatus = 2
	REPO_ARCHIVED AegisRepositoryStatus = 4
)

const (
	REPO_TYPE_GIT uint8 = 1
)

type LocalRepository any

type Repository struct {
	Type uint8 `json:"type"`
	Namespace string `json:"namespace"`
	Name string `json:"name"`
	Description string `json:"description"`
	AccessControlList *ACL `json:"acl"`
	Owner string `json:"owner"`
	Status AegisRepositoryStatus `json:"status"`
	Repository LocalRepository
	LocalPath string `json:"localPath"`
	ForkOriginNamespace string `json:"forkOriginNamespace"`
	ForkOriginName string `json:"forkOriginName"`
	// reserved for features in the future.
	RepoLabelList []string `json:"labelList"`
}

func NewRepository(ns string, name string, localgr LocalRepository) (*Repository, error) {
	return &Repository{
		Namespace: ns,
		Name: name,
		Description: "",
		AccessControlList: nil,
		Status: REPO_NORMAL_PUBLIC,
		Repository: localgr,
		LocalPath: GetLocalRepositoryLocalPath(localgr),
	}, nil
}

func (repo *Repository) FullName() string {
	if len(repo.Namespace) > 0 {
		return fmt.Sprintf("%s:%s", repo.Namespace, repo.Name)
	} else {
		return repo.Name
	}
}

