package service

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/repository"
)

func newAuthTestHarness(t *testing.T) (*AuthService, *sql.DB, func()) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "auth_test.db")
	os.Remove(dbPath)

	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	// Replicate the user+role tables that migration 001_init.up.sql creates,
	// so AuthService tests do not depend on the database package (which itself
	// imports service).
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT DEFAULT 'viewer',
		is_active INTEGER DEFAULT 1,
		last_login_at TEXT,
		created_at TEXT DEFAULT (datetime('now')),
		updated_at TEXT DEFAULT (datetime('now'))
	);`
	if _, err := sqlDB.Exec(schema); err != nil {
		t.Fatalf("schema: %v", err)
	}

	cfg := authTestConfig{
		secret:     "test-secret-that-is-long-enough",
		accessTTL:  5 * time.Minute,
		refreshTTL: 7 * 24 * time.Hour,
	}
	repo := repository.NewUserRepository(sqlDB)
	svc := NewAuthServiceForTest(repo, cfg)
	return svc, sqlDB, func() { sqlDB.Close(); os.Remove(dbPath) }
}

type authTestConfig struct {
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthServiceForTest(repo *repository.UserRepository, cfg authTestConfig) *AuthService {
	return &AuthService{repo: repo, secret: cfg.secret, accessTTL: cfg.accessTTL, refreshTTL: cfg.refreshTTL}
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	svc, _, cleanup := newAuthTestHarness(t)
	defer cleanup()

	req := &models.UserCreateRequest{Username: "alice", Email: "alice@example.com", Password: "CorrectHorseBattery0", Role: "admin"}
	u, err := svc.Register(req)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if u.PasswordHash != "" {
		t.Fatal("password hash must be stripped from response")
	}

	res, err := svc.Login(&models.UserLoginRequest{Username: "alice", Password: "CorrectHorseBattery0"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if res.AccessToken == "" || res.RefreshToken == "" {
		t.Fatal("login must return tokens")
	}
	if res.User.Username != "alice" {
		t.Fatalf("unexpected user: %+v", res.User)
	}

	me, err := svc.Me(u.ID)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	if me.Username != "alice" {
		t.Fatalf("me returned wrong user: %+v", me)
	}
}

func TestAuthService_RegisterDuplicateUser(t *testing.T) {
	svc, _, cleanup := newAuthTestHarness(t)
	defer cleanup()

	req := &models.UserCreateRequest{Username: "dup", Email: "dup@example.com", Password: "Password123!", Role: "viewer"}
	if _, err := svc.Register(req); err != nil {
		t.Fatalf("first register: %v", err)
	}
	_, err := svc.Register(req)
	if err != repository.ErrUserAlreadyExists {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestAuthService_LoginWrongPassword(t *testing.T) {
	svc, _, cleanup := newAuthTestHarness(t)
	defer cleanup()

	svc.Register(&models.UserCreateRequest{Username: "bob", Email: "bob@example.com", Password: "RightOne123!", Role: "viewer"})
	_, err := svc.Login(&models.UserLoginRequest{Username: "bob", Password: "WrongOne123!"})
	if err == nil {
		t.Fatal("expected login failure with wrong password")
	}
}

func TestAuthService_LoginDisabledAccount(t *testing.T) {
	svc, sqlDB, cleanup := newAuthTestHarness(t)
	defer cleanup()

	svc.Register(&models.UserCreateRequest{Username: "locked", Email: "locked@example.com", Password: "Pass123!", Role: "viewer"})
	if _, err := sqlDB.Exec("UPDATE users SET is_active = 0 WHERE username = ?", "locked"); err != nil {
		t.Fatalf("disable user: %v", err)
	}

	_, err := svc.Login(&models.UserLoginRequest{Username: "locked", Password: "Pass123!"})
	if err == nil {
		t.Fatal("expected login failure for disabled account")
	}
}

func TestAuthService_RefreshEndToEnd(t *testing.T) {
	svc, _, cleanup := newAuthTestHarness(t)
	defer cleanup()

	svc.Register(&models.UserCreateRequest{Username: "rf", Email: "rf@example.com", Password: "Pass123!", Role: "operator"})
	login, err := svc.Login(&models.UserLoginRequest{Username: "rf", Password: "Pass123!"})
	if err != nil {
		t.Fatalf("login for refresh test: %v", err)
	}
	refreshed, err := svc.Refresh(login.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshed.AccessToken == "" || refreshed.User.Username != "rf" {
		t.Fatalf("unexpected refresh result: %+v", refreshed)
	}
}