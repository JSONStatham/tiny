package clickhouse

import (
	"analytics/internal/config"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type repository struct {
	db *sql.DB
}

func New(ctx context.Context, cfg *config.Config) (*repository, error) {
	addr := fmt.Sprintf("%s:%s", cfg.DB.Host, cfg.DB.Port)
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: cfg.DB.Name,
			Username: cfg.DB.Username,
			Password: cfg.DB.Username,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: time.Second * 30,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Debug:                true,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
	})
	conn.SetMaxIdleConns(5)
	conn.SetMaxOpenConns(10)
	conn.SetConnMaxLifetime(time.Hour)

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &repository{db: conn}, nil
}

func (r *repository) Close() {
	r.db.Close()
}
