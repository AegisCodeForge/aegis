//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type SettingEmailTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	EmailList []struct{
		Email string
		Verified bool
		Primary bool
	}
	ErrorMsg struct{
		Type string
		Message string
	}
		
}

