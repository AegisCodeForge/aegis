package init

import (
	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/receipt"
	"github.com/bctnry/aegis/pkg/aegis/receipt/sqlite"
	"github.com/bctnry/aegis/pkg/aegis/receipt/postgres"
)

func InitializeReceiptSystem(cfg *aegis.AegisConfig) (receipt.AegisReceiptSystemInterface, error) {
	switch cfg.ReceiptSystem.Type {
	case "sqlite": return sqlite.NewSqliteReceiptSystemInterface(cfg)
	case "postgres": return postgres.NewPostgresReceiptSystemInterface(cfg)
	}
	return nil, receipt.ErrUnsupportedSystemType
}

