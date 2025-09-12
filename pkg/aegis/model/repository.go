package model

import (
	"encoding/json"
	"fmt"
)

type AegisRepositoryStatus int

const (
	REPO_NORMAL_PUBLIC AegisRepositoryStatus = 1
	REPO_NORMAL_PRIVATE AegisRepositoryStatus = 2
	REPO_ARCHIVED AegisRepositoryStatus = 4
	REPO_INTERNAL AegisRepositoryStatus = 5
	REPO_LIMITED AegisRepositoryStatus = 6
)

const (
	REPO_TYPE_GIT uint8 = 1
)

func ValidRepositoryName(s string) bool {
	colonPassed := false
	for _, k := range s {
		if !(('0' <= k && k <= '9') || ('A' <= k && k <= 'Z') || ('a' <= k && k <= 'z') || k == '_' || k == '-') {
			if k == ':' {
				if colonPassed { return false }
				colonPassed = true
				continue
			} else {
				return false
			}
		}
	}
	return true
}

func ValidStrictRepositoryName(s string) bool {
	for _, k := range s {
		if !(('0' <= k && k <= '9') || ('A' <= k && k <= 'Z') || ('a' <= k && k <= 'z') || k == '_' || k == '-') {
			return false
		}
	}
	return true
}

type LocalRepository any

type Repository struct {
	AbsId int64 `json:"absid"`
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
	WebHookConfig *WebHookConfig `json:"webHookConfig"`
}

func ParseWebHookConfig(s string) (*WebHookConfig, error) {
	var r WebHookConfig
	err := json.Unmarshal([]byte(s), &r)
	if err != nil { return nil, err }
	return &r, nil
}
func (whc *WebHookConfig) String() string {
	r, _ := json.Marshal(whc)
	return string(r)
}

type WebHookConfig struct {
	Enable bool `json:"enable"`
	Secret string `json:"secret"`
	TargetURL string `json:"targetUrl"`
	// the type of the payload. currently only supports "json".
	PayloadType string `json:"payloadType"`
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

