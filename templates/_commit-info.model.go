//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/gitlib"

type CommitInfoTemplateModel struct {
	RootPath string
	Commit *gitlib.CommitObject
}

