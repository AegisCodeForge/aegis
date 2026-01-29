//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"
import "github.com/GitusCodeForge/Gitus/pkg/gitlib"

type SnippetSettingTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	Snippet *model.Snippet
}

