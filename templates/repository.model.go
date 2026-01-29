//go:build ignore
package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitlib"

type RepositoryModel struct{
	Config *gitus.GitusConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	BranchList map[string]*gitlib.Branch
	TagList map[string]*gitlib.Tag
	LoginInfo *LoginInfoModel

	MajorBranchPermaLink string
	ReadmeString string
	TreeFileList *TreeFileListTemplateModel
	CommitInfo *CommitInfoTemplateModel
	ComparisonInfo *gitlib.BranchComparisonInfo
}
