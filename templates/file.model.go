//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitlib"

type FileTemplateModel struct {
	Config *gitus.GitusConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	File BlobTextTemplateModel
	PermaLink string

	ComparisonInfo *gitlib.BranchComparisonInfo
	AllowBlame bool
	TreeFileList *TreeFileListTemplateModel
	TreePath *TreePathTemplateModel
	CommitInfo *CommitInfoTemplateModel
	TagInfo *TagInfoTemplateModel
	
	LoginInfo *LoginInfoModel
}

