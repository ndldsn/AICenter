package service

import (
	"fmt"
	"time"

	"github.com/aicenter/aicenter/internal/auth"
	"github.com/aicenter/aicenter/internal/config"
	"github.com/aicenter/aicenter/internal/models"
	"github.com/aicenter/aicenter/internal/repository"
)

// AuthService coordinates user registration, login and token refresh.
type AuthService struct {
	repo       *repository.UserRepository
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(repo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		repo:       repo,
		secret:     cfg.Auth.Secret,
		accessTTL:  cfg.Auth.AccessTokenTTL,
		refreshTTL: cfg.Auth.RefreshTokenTTL,
	}
}

// Register creates a new user. Username/email must be unique and password is
// bcrypt-hashed before storage. The default role is "viewer" when none given.
func (s *AuthService) Register(req *models.UserCreateRequest) (*models.User, error) {
	if req.Role == "" {
		req.Role = "viewer"
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	u := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
		IsActive:     true,
	}
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	// Return without the password hash.
	u.PasswordHash = ""
	return u, nil
}

// Login verifies credentials and returns an access + refresh token pair.
// On success the user's last_login_at is recorded.
func (s *AuthService) Login(req *models.UserLoginRequest) (loginResult, error) {
	u, err := s.repo.GetByUsername(req.Username)
	if err != nil {
		return loginResult{}, fmt.Errorf("invalid credentials")
	}
	if !u.IsActive {
		return loginResult{}, fmt.Errorf("account disabled")
	}
	if !auth.CheckPasswordHash(req.Password, u.PasswordHash) {
		return loginResult{}, fmt.Errorf("invalid credentials")
	}

	access, refresh, err := auth.GenerateTokenPair(u.ID, u.Username, u.Role, s.secret, s.accessTTL, s.refreshTTL)
	if err != nil {
		return loginResult{}, fmt.Errorf("failed to mint tokens: %w", err)
	}

	if err := s.repo.UpdateLastLogin(u.ID); err != nil {
		// Best-effort; token already minted so we don't roll back.
		return loginResult{}, nil
	}

	return loginResult{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         u,
	}, nil
}

// Refresh accepts a refresh token and mints a new access token for the same user.
// The refresh token's Subject claim is the user ID, used to look up the user.
func (s *AuthService) Refresh(refreshToken string) (loginResult, error) {
	u, err := s.lookupByRefreshToken(refreshToken)
	if err != nil {
		return loginResult{}, fmt.Errorf("invalid refresh token")
	}
	if !u.IsActive {
		return loginResult{}, fmt.Errorf("account disabled")
	}

	access, refresh, err := auth.GenerateTokenPair(u.ID, u.Username, u.Role, s.secret, s.accessTTL, s.refreshTTL)
	if err != nil {
		return loginResult{}, fmt.Errorf("failed to mint tokens: %w", err)
	}

	return loginResult{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         u,
	}, nil
}

// Me returns the current user identified by userID (injected by JWTAuth middleware).
func (s *AuthService) Me(userID string) (*models.User, error) {
	u, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	u.PasswordHash = ""
	return u, nil
}

// lookupByRefreshToken verifies the refresh token signature and extracts the
// user by Subject claim. The caller owns secret + expiry.
func (s *AuthService) lookupByRefreshToken(refreshToken string) (*models.User, error) {
	id, err := auth.ValidateRefreshToken(refreshToken, s.secret)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

type loginResult struct {
	AccessToken  string
	RefreshToken string
	User         *models.User
}