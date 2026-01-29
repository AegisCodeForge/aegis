//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"
import "github.com/GitusCodeForge/Gitus/pkg/gitlib"

type TagTemplateModel struct {
	Config *gitus.GitusConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	Tag *gitlib.TagObject
	TagInfo *TagInfoTemplateModel
	LoginInfo *LoginInfoModel
}

