//go:build ignore

package templates

type FileTemplateModel struct {
	RepoHeaderInfo RepoHeaderTemplateModel
	File BlobTextTemplateModel
	PermaLink string

	TreePath *TreePathTemplateModel
	CommitInfo *CommitInfoTemplateModel
	TagInfo *TagInfoTemplateModel
}

