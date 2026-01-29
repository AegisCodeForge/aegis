//go:build ignore
package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type TreeTemplateModel struct {
	Config *gitus.GitusConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	TreeFileList *TreeFileListTemplateModel
	ParentTreeFileList *TreeFileListTemplateModel
	PermaLink string
	
	ComparisonInfo *gitlib.BranchComparisonInfo
	TreePath *TreePathTemplateModel
	CommitInfo *CommitInfoTemplateModel
	TagInfo *TagInfoTemplateModel

	LoginInfo *LoginInfoModel
}

