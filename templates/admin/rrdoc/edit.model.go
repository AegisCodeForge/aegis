//go:build ignore

package templates


import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type AdminRRDocEditTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	DocumentNumber int
	Title string
	Path string
	Content string
}

