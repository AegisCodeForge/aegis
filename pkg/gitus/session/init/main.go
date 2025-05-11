package init

import (
	"github.com/bctnry/gitus/pkg/gitus"
	"github.com/bctnry/gitus/pkg/gitus/db"
	"github.com/bctnry/gitus/pkg/gitus/session"
	"github.com/bctnry/gitus/pkg/gitus/session/sqlite"
)


func InitializeDatabase(cfg *gitus.GitusConfig) (session.GitusSessionStore, error) {
	switch cfg.DatabaseType {
	case "sqlite": return sqlite.NewGitusSqliteSessionStore(cfg)
	}
	return nil, db.NewGitusDatabaseError(db.DATABASE_NOT_SUPPORTED, "")
}

