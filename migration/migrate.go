package migration

import (
	"embed"
	"fmt"
	"go-pg-demo/pkgs"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
)

//go:embed db/*.sql
var migrationsFS embed.FS

func RunMigrations(db *sqlx.DB, config *pkgs.Config) error {
	sourceDriver, err := iofs.New(migrationsFS, "db")
	if err != nil {
		return fmt.Errorf("failed to create source driver: %w", err)
	}

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.Database.Username,
		url.QueryEscape(config.Database.Password),
		config.Database.Host,
		config.Database.Port,
		config.Database.DBName,
		config.Database.SSLMode,
	)

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, connStr)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
