//go:build ignore

package templates


import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type AdminRRDocListTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
}

