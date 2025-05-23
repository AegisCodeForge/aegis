//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"

type IndexStaticTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	FrontPage string
}

