package database

import (
	"fmt"
	"strings"

	"github.com/111zxc/rss-communicator/internal/repository"
	"github.com/111zxc/rss-communicator/internal/repository/postgres"
	"github.com/111zxc/rss-communicator/internal/repository/sqlite"
)

func Open(driver string, dsn string) (repository.Database, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", "postgres", "postgresql":
		return postgres.New(dsn)
	case "sqlite", "sqlite3":
		return sqlite.New(dsn)
	default:
		return nil, fmt.Errorf("unsupported db driver %q", driver)
	}
}
