package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)


type PostgresAegisDatabaseInterface struct {
	config *aegis.AegisConfig
	pool *pgxpool.Pool
}

func NewPostgresAegisDatabaseInterface(cfg *aegis.AegisConfig) (*PostgresAegisDatabaseInterface, error) {
	u := &url.URL{
		Scheme: "postgres",
		User: url.UserPassword(cfg.Database.UserName, cfg.Database.Password),
		Host: cfg.Database.URL,
		Path: cfg.Database.DatabaseName,
	}
	pool, err := pgxpool.New(context.TODO(), u.String())
	if err != nil { return nil, err }
	return &PostgresAegisDatabaseInterface{
		config: cfg,
		pool: pool,
	}, nil
}

func (dbif *PostgresAegisDatabaseInterface) Dispose() error {
	dbif.pool.Close()
	return nil
}

var requiredTableList = []string{
	"user_authkey",
	"user_signkey",
	"user",
	"namespace",
	"repository",
	"issue",
	"issue_event",
	"pull_request",
	"pull_request_event",
}

func (dbif *PostgresAegisDatabaseInterface) IsDatabaseUsable() (bool, error) {
	ctx := context.Background()
	queryStr := `
SELECT EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = $1)
`
	for _, item := range requiredTableList {
		tableName := fmt.Sprintf("%s_%s", dbif.config.Database.TablePrefix, item)
		stmt := dbif.pool.QueryRow(ctx, queryStr, tableName)
		var a bool
		err := stmt.Scan(&a)
		if errors.Is(err, pgx.ErrNoRows) { return false, nil }
		if err != nil { return false, err }
		if (!a) { return false, nil }
	}
	return true, nil
}

