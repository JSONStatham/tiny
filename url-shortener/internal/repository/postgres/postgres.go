package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"urlshortener/internal/config"
	"urlshortener/internal/models"
	"urlshortener/internal/repository"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	cfg *config.Config
	DB  *sqlx.DB
}

func New(cfg *config.Config) (*Repository, error) {
	const op = "repository.postgres.New"

	dns := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.DB.Username,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)
	db, err := sqlx.Connect("pgx", dns)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Repository{cfg, db}, nil
}

func (r *Repository) GetURL(ctx context.Context, short_url string) (*models.URL, error) {
	const op = "repository.postgres.GetURL"

	query := "SELECT id, original_url, short_url, created_at FROM url WHERE short_url=$1 LIMIT 1"

	url := &models.URL{}
	if err := r.DB.GetContext(ctx, url, query, short_url); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrURLNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return url, nil
}

func (r *Repository) FetchAll(ctx context.Context) ([]*models.URL, error) {
	const op = "repository.postgres.GetURL"

	query := "SELECT id, short_url, original_url, created_at FROM url"

	var urls []*models.URL
	if err := r.DB.SelectContext(ctx, &urls, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrURLNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return urls, nil
}

func (r *Repository) SaveURL(ctx context.Context, urlToSave, short_url string) error {
	const op = "repository.postgres.SaveURL"

	query := "INSERT INTO url (original_url, short_url) VALUES ($1, $2)"

	_, err := r.DB.Exec(query, urlToSave, short_url)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.SQLState() == pgerrcode.UniqueViolation {
			return fmt.Errorf("%s: %w", op, repository.ErrURLExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Repository) DeleteURL(ctx context.Context, short_url string) error {
	const op = "repository.postgres.DeleteURL"

	query := "DELETE FROM url WHERE short_url=$1"

	res, err := r.DB.ExecContext(context.Background(), query, short_url)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, repository.ErrURLNotFound)
	}

	return nil
}
