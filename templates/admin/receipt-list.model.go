//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/receipt"

type AdminReceiptListTemplateModel struct {
	Config *gitus.GitusConfig
	LoginInfo *LoginInfoModel
	ErrorMsg string
	ReceiptList []*receipt.Receipt
	PageInfo *PageInfoModel
	Query string
}

