//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitlib"

type CommitInfoTemplateModel struct {
	RootPath string
	Commit *gitlib.CommitObject
	EmailUserMapping map[string]string
}

