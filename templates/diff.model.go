//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitlib"

type DiffTemplateModel struct {
	RepoHeaderInfo RepoHeaderTemplateModel
	CommitInfo CommitInfoTemplateModel
	Diff string
}
