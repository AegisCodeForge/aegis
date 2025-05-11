package init

import (
	"github.com/bctnry/gitus/pkg/gitus"
	"github.com/bctnry/gitus/pkg/gitus/db"
	"github.com/bctnry/gitus/pkg/gitus/db/sqlite"
)


func InitializeDatabase(cfg *gitus.GitusConfig) (db.GitusDatabaseInterface, error) {
	switch cfg.DatabaseType {
	case "sqlite": return sqlite.NewSqliteGitusDatabaseInterface(cfg)
	}
	return nil, db.NewGitusDatabaseError(db.DATABASE_NOT_SUPPORTED, "")
}

