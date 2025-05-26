//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/receipt"

type AdminReceiptListTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	ErrorMsg string
	ReceiptList []*receipt.Receipt
	PageInfo *PageInfoModel
	Query string
}

