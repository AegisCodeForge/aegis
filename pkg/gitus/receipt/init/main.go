package init

import (
	"github.com/GitusCodeForge/Gitus/pkg/gitus"
	"github.com/GitusCodeForge/Gitus/pkg/gitus/receipt"
	"github.com/GitusCodeForge/Gitus/pkg/gitus/receipt/sqlite"
	"github.com/GitusCodeForge/Gitus/pkg/gitus/receipt/postgres"
)

func InitializeReceiptSystem(cfg *gitus.GitusConfig) (receipt.GitusReceiptSystemInterface, error) {
	switch cfg.ReceiptSystem.Type {
	case "sqlite": return sqlite.NewSqliteReceiptSystemInterface(cfg)
	case "postgres": return postgres.NewPostgresReceiptSystemInterface(cfg)
	}
	return nil, receipt.ErrUnsupportedSystemType
}

