package model

import (
	"encoding/json"
	"os"
)

const (
	SIMPLE_MODE_VISIBILITY_PUBLIC = "public"
	SIMPLE_MODE_VISIBILITY_PRIVATE = "private"
	SIMPLE_MODE_PERM_ALLOW = "allow"
	SIMPLE_MODE_PERM_DISALLOW = "disallow"
	SIMPLE_MODE_ACTION_BRANCH_PUSH = "commit"
	SIMPLE_MODE_ACTION_BRANCH_DELETE = "delete"
	SIMPLE_MODE_ACTION_TAG_UNANNONATED = "commit"
	SIMPLE_MODE_ACTION_TAG_DELETE = "delete"
	SIMPLE_MODE_ACTION_TAG_ADD = "tag"
)

type SimpleModeNamespaceConfig struct {
	Namespace struct {
		Title string `json:"title"`
		Description string `json:"description"`
		Visibility string `json:"visibility"`
	} `json:"namespace"`
	RepositoryList map[string]*SimpleModeRepositoryConfig
}

type SimpleModeUserACL struct {
	Default string `json:"default"`
	// NOTE(2025.10.31): Patterns are postponed until we come back to the pack format.
	// Pattern map[string][]string `json:"patterns"`
}

type SimpleModeRepositoryConfig struct {
	Repository struct {
		Description string `json:"description"`
		Visibility string `json:"visibility"`
	} `json:"repo"`
	Hooks map[string]string `json:"hooks"`
	Users map[string]*SimpleModeUserACL `json:"users"`
}

type SimpleModeConfigCache map[string]*SimpleModeNamespaceConfig

func ReadRepositoryConfigFromFile(filePath string) (*SimpleModeRepositoryConfig, error) {
	s, err := os.ReadFile(filePath)
	if err != nil { return nil, err }
	var res SimpleModeRepositoryConfig
	err = json.Unmarshal(s, &res)
	if err != nil { return nil, err }
	return &res, nil
}

func ReadNamespaceConfigFromFile(filePath string) (*SimpleModeNamespaceConfig, error) {
	s, err := os.ReadFile(filePath)
	if err != nil { return nil, err }
	var res SimpleModeNamespaceConfig
	err = json.Unmarshal(s, &res)
	if err != nil { return nil, err }
	return &res, nil
}

