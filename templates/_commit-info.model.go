//go:build ignore

package templates

type CommitInfoTemplateModel struct {
	RepoName string
	Commit *gitlib.CommitObject
}

