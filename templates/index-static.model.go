//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"

type IndexStaticTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	FrontPage string
}

