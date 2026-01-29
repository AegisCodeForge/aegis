//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type AdminRegistrationRequestTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	ErrorMsg string
	RequestList []*model.RegistrationRequest
	PageInfo *PageInfoModel
	Query string
}

