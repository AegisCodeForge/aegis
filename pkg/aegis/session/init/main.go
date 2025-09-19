package init

import (
	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/db"
	"github.com/bctnry/aegis/pkg/aegis/session"
	"github.com/bctnry/aegis/pkg/aegis/session/in_memory"
	"github.com/bctnry/aegis/pkg/aegis/session/memcached"
	"github.com/bctnry/aegis/pkg/aegis/session/redis_like"
	"github.com/bctnry/aegis/pkg/aegis/session/sqlite"
)


func InitializeDatabase(cfg *aegis.AegisConfig) (session.AegisSessionStore, error) {
	switch cfg.Session.Type {
	case "sqlite": return sqlite.NewAegisSqliteSessionStore(cfg)
	case "valkey": fallthrough
	case "keydb": fallthrough
	case "redis":
		return redis_like.NewAegisRedisLikeSessionStore(cfg)
	case "memcached":
		return memcached.NewAegisMemcachedSessionStore(cfg)
	case "in_memory":
		return in_memory.NewAegisInMemorySessionStore(cfg)
	}
	return nil, db.ErrDatabaseNotSupported
}

