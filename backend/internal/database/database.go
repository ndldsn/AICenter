package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

var db *sql.DB

// Initialize sets up the database connection
func Initialize(url string, log *zap.Logger) (*sql.DB, error) {
	var dsn string
	if strings.HasPrefix(url, "sqlite:") {
		dsn = strings.TrimPrefix(url, "sqlite:")
		// Handle relative paths
		if strings.HasPrefix(dsn, "//") {
			dsn = strings.TrimPrefix(dsn, "/")
		}
		if !filepath.IsAbs(dsn) {
			wd, _ := os.Getwd()
			dsn = filepath.Join(wd, dsn)
		}
		log.Info(fmt.Sprintf("Using SQLite database: %s", dsn))
	} else if strings.HasPrefix(url, "postgresql:") || strings.HasPrefix(url, "postgres:") {
		dsn = strings.TrimPrefix(url, "postgresql:")
		dsn = strings.TrimPrefix(dsn, "//")
		log.Info("Using PostgreSQL database")
	} else {
		dsn = url
		log.Info(fmt.Sprintf("Using database URL: %s", url))
	}

	// Ensure directory exists for SQLite
	if !strings.Contains(dsn, "host=") {
		dir := filepath.Dir(dsn)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	driver := "sqlite"
	if strings.Contains(dsn, "host=") || strings.Contains(dsn, "postgres") {
		driver = "postgres"
	}

	sqlDB, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	db = sqlDB
	return sqlDB, nil
}

// Get returns the global database connection
func Get() *sql.DB {
	return db
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
