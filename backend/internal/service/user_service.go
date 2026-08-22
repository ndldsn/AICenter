package service

import (
	"fmt"

	"github.com/aicenter/aicenter/internal/database"
	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/repository"
)

// UserService handles administrative user management (list/update/role/delete).
type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

type UserListResult struct {
	Users []*models.User
	Total int64
}

func (s *UserService) List(q string, page, pageSize int) (*UserListResult, error) {
	// Safety bounds.
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	users, total, err := s.repo.List(q, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &UserListResult{Users: users, Total: total}, nil
}

func (s *UserService) GetByID(id string) (*models.User, error) {
	u, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	u.PasswordHash = ""
	return u, nil
}

type UpdateUserRequest struct {
	Username  *string `json:"username,omitempty"`
	Email     *string `json:"email,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

func (s *UserService) Update(id string, req UpdateUserRequest) (*models.User, error) {
	err := s.repo.Update(id, req.Username, req.Email, req.IsActive)
	if err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

type ChangeRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

func (s *UserService) ChangeRole(userID string, req ChangeRoleRequest) (*models.User, error) {
	// Role existence is enforced against the roles table so admins can also
	// assign custom roles they created — keeping the assignment surface in
	// sync with the extensible role catalog.
	db := database.Get()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM roles WHERE name = ?", req.Role).Scan(&count); err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, fmt.Errorf("role %q does not exist", req.Role)
	}
	if err := s.repo.UpdateRole(userID, req.Role); err != nil {
		return nil, err
	}
	return s.GetByID(userID)
}

func (s *UserService) Delete(id string) error {
	return s.repo.Delete(id)
}