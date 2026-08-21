package database

import (
	"database/sql"
	"time"
)

// SeedData adds initial data for development
func SeedData(db *sql.DB) error {
	// Check if already seeded
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM servers").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // Already has data
	}

	// Add test servers
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := db.Exec(`
		INSERT INTO servers (id, name, host, port, username, auth_type, status, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "srv-demo-001", "Demo Web Server", "192.168.1.10", 22, "ubuntu", "key", "offline", `["web","production"]`, now, now)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO servers (id, name, host, port, username, auth_type, status, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "srv-demo-002", "Demo DB Server", "192.168.1.11", 22, "root", "password", "offline", `["database","mysql"]`, now, now)
	if err != nil {
		return err
	}

	return nil
}
