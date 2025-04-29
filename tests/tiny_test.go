package tests

import (
	"database/sql"
	"net/url"
	"testing"
	"tiny/internal/config"
	"tiny/internal/handlers"
	"tiny/internal/storage/postgres"
	random "tiny/internal/utils"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const host = "localhost:8080"

var cfg = config.Config{
	DB: config.DB{
		User:     "test",
		Password: "test",
		Host:     "test",
		Port:     "5433",
		Name:     "test",
	},
}

func setupTestDB(t *testing.T) func() {
	// os.Setenv("CONFIG_PATH", "../configs/test.yml")

	// Database connection string for testing
	connStr := "postgres://test:test@localhost:5433/test?sslmode=disable"

	// Create a new database connection pool (pgx v5)
	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)

	// Initialize the migration driver
	driver, err := pgx.WithInstance(db, &pgx.Config{})
	require.NoError(t, err)

	// Initialize migration instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://../migrations",
		"postgres",
		driver,
	)
	require.NoError(t, err)

	// Apply migrations
	err = m.Up()
	require.NoError(t, err)

	// Return cleanup function to rollback migrations and close DB
	cleanup := func() {
		// Rollback migrations (if needed)
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			t.Fatalf("Failed to rollback migrations: %v", err)
		}
		db.Close()
	}

	return cleanup
}

func TestTiny_HappyPath(t *testing.T) {
	t.Parallel()

	cleanup := setupTestDB(t)
	defer cleanup()

	u := url.URL{
		Host:   host,
		Scheme: "http",
	}

	e := httpexpect.Default(t, u.String())

	alias := random.RandomString(10)
	// Create new url
	e.POST("/url").
		WithJSON(handlers.Request{
			URL:   gofakeit.URL(),
			Alias: alias,
		}).
		WithBasicAuth("user", "root").
		Expect().
		Status(201).
		JSON().Object().
		ContainsKey("alias")

	// Ensures url exists in db
	storage, err := postgres.New(&cfg)
	if err != nil {
		require.Error(t, err)
	}

	url, err := storage.GetURL(alias)
	require.Error(t, err)

	assert.NotNil(t, url)

	assert.Equal(t, alias, url.Alias)
}
