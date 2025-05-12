package model

import (
	"fmt"

	"github.com/bctnry/gitus/pkg/gitlib"
)

type GitusRepositoryStatus int

const (
	REPO_NORMAL_PUBLIC GitusRepositoryStatus = 1
	REPO_NORMAL_PRIVATE GitusRepositoryStatus = 2
	REPO_DELETED GitusRepositoryStatus = 3
	REPO_ARCHIVED GitusRepositoryStatus = 4
)

type Repository struct {
	Namespace string `json:"namespace"`
	Name string `json:"name"`
	Description string `json:"description"`
	AccessControlList string `json:"acl"`
	Status GitusRepositoryStatus `json:"status"`
	Repository *gitlib.LocalGitRepository `json:"localGitRepo"`
	LocalPath string `json:"localPath"`
}

func NewRepository(ns string, name string, localgr *gitlib.LocalGitRepository) (*Repository, error) {
	return &Repository{
		Namespace: ns,
		Name: name,
		Description: localgr.Description,
		AccessControlList: "",
		Status: REPO_NORMAL_PUBLIC,
		Repository: localgr,
		LocalPath: localgr.GitDirectoryPath,
	}, nil
}

func (repo *Repository) FullName() string {
	return fmt.Sprintf("%s:%s", repo.Namespace, repo.Name)
}

