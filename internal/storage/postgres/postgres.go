package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"tiny/internal/config"
	"tiny/internal/models"
	"tiny/internal/storage"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	cfg config.Config
	log *slog.Logger
	DB  *pgxpool.Pool
}

func New(cfg config.Config, log *slog.Logger) (*Storage, error) {
	const op = "storage.postgres.New"

	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)
	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &Storage{cfg, log, db}, nil
}

func (s Storage) GetURL(alias string) (*models.URL, error) {
	const op = "storage.postgres.GetURL"
	var url models.URL

	args := pgx.NamedArgs{"alias": alias}
	query := "SELECT id, url, alias, created_at FROM url WHERE alias=@alias LIMIT 1"
	err := s.DB.QueryRow(context.Background(), query, args).Scan(&url.ID, &url.URL, &url.Alias, &url.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr == pgx.ErrNoRows {
			return nil, storage.ErrURLNotFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &url, nil
}

func (s Storage) SaveURL(urlToSave, alias string) error {
	const op = "storage.postgres.SaveURL"

	query := "INSERT INTO url (url, alias) VALUES (@url, @alias)"
	args := pgx.NamedArgs{
		"url":   urlToSave,
		"alias": alias,
	}

	_, err := s.DB.Exec(context.Background(), query, args)
	fmt.Println(err)
	if err != nil {
		var pgErr *pgconn.PgError
		fmt.Println("MY ERROR:", pgErr)
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s Storage) DeleteURL(alias string) error {
	const op = "storage.postgres.DeleteURL"

	res, err := s.DB.Exec(context.Background(), "DELETE FROM url WHERE alias=@alias", pgx.NamedArgs{"alias": alias})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if res.RowsAffected() == 0 {
		return storage.ErrURLNotFound
	}

	return nil
}
