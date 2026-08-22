package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/pkg/utils"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// UserRepository persists users rows.
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user and returns the populated user.
func (r *UserRepository) Create(u *models.User) error {
	u.ID = uuid.New().String()
	now := time.Now().Format("2006-01-02 15:04:05")
	u.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", now)
	u.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", now)

	result, err := r.db.Exec(`
		INSERT INTO users (id, username, email, password_hash, role, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Username, u.Email, u.PasswordHash, u.Role, u.IsActive, now, now)
	if err != nil {
		// SQLite surfaces uniqueness violations as a constraint error; we map
		// both username and email collisions to a single business error.
		if isUniquenessViolation(err) {
			return ErrUserAlreadyExists
		}
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrUserAlreadyExists
	}
	return nil
}

// GetByID loads a single user by id.
func (r *UserRepository) GetByID(id string) (*models.User, error) {
	u, err := r.scanUser(r.db.QueryRow(
		"SELECT id, username, email, password_hash, role, is_active, last_login_at, created_at, updated_at FROM users WHERE id = ?", id))
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetByUsername loads a single user by username.
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	u, err := r.scanUser(r.db.QueryRow(
		"SELECT id, username, email, password_hash, role, is_active, last_login_at, created_at, updated_at FROM users WHERE username = ?", username))
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateLastLogin records the user's last login time.
func (r *UserRepository) UpdateLastLogin(id string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := r.db.Exec("UPDATE users SET last_login_at = ?, updated_at = ? WHERE id = ?", now, now, id)
	return err
}

// List returns users with optional username filter and pagination.
func (r *UserRepository) List(q string, page, pageSize int) ([]*models.User, int64, error) {
	where := ""
	args := []interface{}{}
	if q != "" {
		where = " WHERE username LIKE ? OR email LIKE ?"
		like := "%" + q + "%"
		args = append(args, like, like)
	}
	args = append(args, pageSize, (page-1)*pageSize)

	var total int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM users"+where, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query("SELECT id, username, email, password_hash, role, is_active, last_login_at, created_at, updated_at FROM users"+where+" ORDER BY created_at DESC LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []*models.User
	for rows.Next() {
		var u models.User
		var lastLogin sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &lastLogin, &createdAt, &updatedAt); err != nil {
			return nil, 0, err
		}
		u.CreatedAt, _ = utils.ParseTimestamp(createdAt)
		u.UpdatedAt, _ = utils.ParseTimestamp(updatedAt)
		if lastLogin.Valid {
			t, err := utils.ParseTimestamp(lastLogin.String)
			if err == nil && !t.IsZero() {
				u.LastLoginAt = &t
			}
		}
		out = append(out, &u)
	}
	return out, total, nil
}

// Update changes editable fields of a user.
func (r *UserRepository) Update(id string, username *string, email *string, isActive *bool) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	if username == nil && email == nil && isActive == nil {
		_, err := r.db.Exec("UPDATE users SET updated_at = ? WHERE id = ?", now, id)
		return err
	}
	set, args := []string{"updated_at = ?"}, []interface{}{now}
	if username != nil {
		set = append(set, "username = ?")
		args = append(args, *username)
	}
	if email != nil {
		set = append(set, "email = ?")
		args = append(args, *email)
	}
	if isActive != nil {
		set = append(set, "is_active = ?")
		args = append(args, *isActive)
	}
	args = append(args, id)
	if _, err := r.db.Exec("UPDATE users SET "+strings.Join(set, ", ")+` WHERE id = ?`, args...); err != nil {
		return err
	}
	return nil
}

// UpdateRole changes a user's role (admin-only assignment).
func (r *UserRepository) UpdateRole(id, role string) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := r.db.Exec("UPDATE users SET role = ?, updated_at = ? WHERE id = ?", role, now, id)
	return err
}

// Delete removes a user. It is a no-op when the user does not exist.
func (r *UserRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func (r *UserRepository) scanUser(row interface{ Scan(dest ...interface{}) error }) (*models.User, error) {
	var u models.User
	var lastLogin sql.NullString
	var createdAt, updatedAt string
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &lastLogin, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	u.CreatedAt, _ = utils.ParseTimestamp(createdAt)
	u.UpdatedAt, _ = utils.ParseTimestamp(updatedAt)
	if lastLogin.Valid {
		t, err := utils.ParseTimestamp(lastLogin.String)
		if err == nil && !t.IsZero() {
			u.LastLoginAt = &t
		}
	}
	return &u, nil
}

// isUniquenessViolation detects the SQLite uniqueness constraint error that
// surfaces when username/email already exists.
func isUniquenessViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint") || strings.Contains(msg, "SQLITE_CONSTRAINT")
}