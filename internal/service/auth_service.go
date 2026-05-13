package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/wmp/auth-service/internal/model"
	"github.com/wmp/auth-service/internal/repository"
)

var (
	ErrEmailExists      = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrValidation       = errors.New("validation error")
)

type AuthService struct {
	userRepo        *repository.UserRepository
	jwtService      *JWTService
	portfolioClient *PortfolioClient
}

func NewAuthService(
	userRepo *repository.UserRepository,
	jwtService *JWTService,
	portfolioClient *PortfolioClient,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		jwtService:      jwtService,
		portfolioClient: portfolioClient,
	}
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, error) {
	if err := validateRegisterRequest(req); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrValidation, err.Error())
	}

	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hash),
		FullName:     req.FullName,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	slog.Info("user registered", "id", user.ID, "email", user.Email)

	// Sync call to portfolio-service to create user and seed starter portfolio
	if err := s.portfolioClient.CreateUser(user.ID, user.Email, user.FullName); err != nil {
		slog.Warn("failed to create user in portfolio-service", "userId", user.ID, "error", err)
		// Registration still succeeds — portfolio-side user can be created later
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &model.AuthResponse{
		Token:    token,
		UserID:   user.ID,
		Email:    user.Email,
		FullName: user.FullName,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, fmt.Errorf("%w: email and password are required", ErrValidation)
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &model.AuthResponse{
		Token:    token,
		UserID:   user.ID,
		Email:    user.Email,
		FullName: user.FullName,
	}, nil
}

func validateRegisterRequest(req model.RegisterRequest) error {
	if strings.TrimSpace(req.Email) == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return errors.New("invalid email format")
	}
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if strings.TrimSpace(req.FullName) == "" {
		return errors.New("fullName is required")
	}
	return nil
}
