package services

import (
	"context"
	"errors"
	"time"

	"github.com/controlewise/backend/internal/config"
	"github.com/controlewise/backend/internal/database"
	"github.com/controlewise/backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db     *database.DB
	jwtCfg config.JWTConfig
}

func NewAuthService(db *database.DB, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{
		db:     db,
		jwtCfg: jwtCfg,
	}
}

type RegisterRequest struct {
	OrganizationName string `json:"organization_name"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Phone            string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Start transaction
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Create organization
	orgID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO organizations (id, name, email, is_active)
		VALUES ($1, $2, $3, $4)
	`, orgID, req.OrganizationName, req.Email, true)
	if err != nil {
		return nil, err
	}

	// Create admin user
	userID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, phone, role, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, userID, orgID, req.Email, string(hashedPassword), req.FirstName, req.LastName, req.Phone, models.RoleAdmin, true)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Get user
	user := &models.User{
		ID:             userID,
		OrganizationID: orgID,
		Email:          req.Email,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Role:           models.RoleAdmin,
		IsActive:       true,
	}

	// Generate token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	var user models.User
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, organization_id, email, password_hash, first_name, last_name, phone, avatar, role, is_active
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`, req.Email).Scan(
		&user.ID,
		&user.OrganizationID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.Avatar,
		&user.Role,
		&user.IsActive,
	)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("user account is not active")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Update last login
	_, err = s.db.Pool.Exec(ctx, `
		UPDATE users SET last_login_at = $1 WHERE id = $2
	`, time.Now(), user.ID)
	if err != nil {
		// Log error but don't fail login
	}

	// Generate token
	token, err := s.generateToken(&user)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  &user,
	}, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, organization_id, email, first_name, last_name, phone, avatar, role, is_active, last_login_at, created_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, userID).Scan(
		&user.ID,
		&user.OrganizationID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.Avatar,
		&user.Role,
		&user.IsActive,
		&user.LastLoginAt,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) generateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":         user.ID.String(),
		"organization_id": user.OrganizationID.String(),
		"role":            user.Role,
		"exp":             time.Now().Add(s.jwtCfg.Expiry).Unix(),
		"iat":             time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtCfg.Secret))
}
