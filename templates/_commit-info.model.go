//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/gitlib"

type CommitInfoTemplateModel struct {
	RepoName string
	Commit *gitlib.CommitObject
}

