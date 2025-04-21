//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitlib"

type CommitInfoTemplateModel struct {
	RepoName string
	Commit *gitlib.CommitObject
}

