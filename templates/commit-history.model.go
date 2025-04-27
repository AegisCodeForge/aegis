//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitlib"

type CommitHistoryModel struct {
	RepoHeaderInfo RepoHeaderTemplateModel
	Commit gitlib.CommitObject
	CommitHistory []gitlib.CommitObject
}

