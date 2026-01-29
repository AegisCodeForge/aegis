//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"
import "github.com/GitusCodeForge/Gitus/pkg/gitlib"

type SnippetAllFileTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	Snippet *model.Snippet
	DisplayingFileList map[string]string
}

