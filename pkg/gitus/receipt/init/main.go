package init

import (
	"github.com/bctnry/gitus/pkg/gitus"
	"github.com/bctnry/gitus/pkg/gitus/receipt"
	"github.com/bctnry/gitus/pkg/gitus/receipt/sqlite"
)

func InitializeReceiptSystem(cfg *gitus.GitusConfig) (receipt.GitusReceiptSystemInterface, error) {
	switch cfg.ReceiptSystem.Type {
	case "sqlite": return sqlite.NewSqliteReceiptSystemInterface(cfg)
	}
	return nil, receipt.ErrUnsupportedSystemType
}

