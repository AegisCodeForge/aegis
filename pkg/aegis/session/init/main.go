package init

import (
	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/db"
	"github.com/bctnry/aegis/pkg/aegis/session"
	"github.com/bctnry/aegis/pkg/aegis/session/sqlite"
)


func InitializeDatabase(cfg *aegis.AegisConfig) (session.AegisSessionStore, error) {
	switch cfg.Session.Type {
	case "sqlite": return sqlite.NewAegisSqliteSessionStore(cfg)
	}
	return nil, db.NewAegisDatabaseError(db.DATABASE_NOT_SUPPORTED, "")
}

