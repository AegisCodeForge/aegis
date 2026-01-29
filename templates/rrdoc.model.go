//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"

type RRDocTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	Title string
	DocumentContent string
}

