package postgres

import (
	"analytics/internal/config"
	"context"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func New(ctx context.Context, cfg *config.Config) (*Repository, error) {
	const op = "repository.postgres.New"

	dns := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.DB.Username,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)
	db, err := sqlx.ConnectContext(ctx, "pgx", dns)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Repository{db}, nil
}

func (r *Repository) Close() {
	r.db.Close()
}
