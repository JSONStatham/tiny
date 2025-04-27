package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"tiny/internal/config"
	"tiny/internal/models"
	"tiny/internal/storage"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	cfg *config.Config
	DB  *sqlx.DB
}

func New(cfg *config.Config) (*Storage, error) {
	const op = "storage.postgres.New"

	dns := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)
	db, err := sqlx.Connect("pgx", dns)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{cfg, db}, nil
}

func (s Storage) GetURL(alias string) (*models.URL, error) {
	const op = "storage.postgres.GetURL"

	query := "SELECT id, url, alias, created_at FROM url WHERE alias=$1 LIMIT 1"

	url := &models.URL{}
	if err := s.DB.Get(url, query, alias); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrURLNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return url, nil
}

func (s Storage) FetchAll(ctx context.Context) ([]*models.URL, error) {
	const op = "storage.postgres.GetURL"

	query := "SELECT id, url, alias, created_at FROM url"

	var urls []*models.URL
	if err := s.DB.SelectContext(ctx, &urls, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrURLNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return urls, nil
}

func (s Storage) SaveURL(urlToSave, alias string) error {
	const op = "storage.postgres.SaveURL"

	query := "INSERT INTO url (url, alias) VALUES ($1, $2)"

	_, err := s.DB.Exec(query, urlToSave, alias)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.SQLState() == pgerrcode.UniqueViolation {
			return fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s Storage) DeleteURL(alias string) error {
	const op = "storage.postgres.DeleteURL"

	query := "DELETE FROM url WHERE alias=$1"

	res, err := s.DB.ExecContext(context.Background(), query, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return storage.ErrURLNotFound
	}

	return nil
}
