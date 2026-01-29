//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"
import "github.com/GitusCodeForge/Gitus/pkg/gitlib"

type SnippetEditFileTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	BelongingUser string
	Name string
	FileName string
	FileContent string
}

