package init

import (
	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/db"
	"github.com/bctnry/aegis/pkg/aegis/db/postgres"
	"github.com/bctnry/aegis/pkg/aegis/db/sqlite"
)


func InitializeDatabase(cfg *aegis.AegisConfig) (db.AegisDatabaseInterface, error) {
	switch cfg.Database.Type {
	case "sqlite": return sqlite.NewSqliteAegisDatabaseInterface(cfg)
	case "postgres": return postgres.NewPostgresAegisDatabaseInterface(cfg)
	}
	return nil, db.ErrDatabaseNotSupported
}

