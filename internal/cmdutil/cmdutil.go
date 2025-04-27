package cmdutil

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/hionay/quotes/internal/config"
)

func NewMySQLPool(ctx context.Context, cfg *config.Config) (*sql.DB, error) {
	maxConns := cfg.DBMaxOpenConns()
	if maxConns <= 0 {
		maxConns = 1
	}
	idleConns := cfg.DBMaxIdleConns()
	if idleConns <= 0 {
		idleConns = 1
	}

	db, err := sql.Open("mysql", cfg.MySQLDSN())
	if err != nil {
		return nil, fmt.Errorf("sql.Open(%q): %w", cfg.MySQLDSN(), err)
	}

	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(idleConns)
	db.SetConnMaxLifetime(1 * time.Hour)
	db.SetConnMaxIdleTime(30 * time.Second)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db.PingContext(): %w", err)
	}
	return db, nil
}
