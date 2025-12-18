package user

import (
	"context"
	"errors"

	"fitslot/internal/auth"
)

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*User, string, string, error)
	Login(ctx context.Context, req LoginRequest) (*User, string, string, error)
	GetByID(ctx context.Context, userID int) (*User, error)
	RefreshToken(ctx context.Context, refreshToken, jwtSecret string) (string, *User, error)
}

type service struct {
	repo      Repository
	jwtSecret string
}

func NewService(repo Repository, jwtSecret string) Service {
	return &service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*User, string, string, error) {
	exists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, "", "", err
	}
	if exists {
		return nil, "", "", ErrEmailExists
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, "", "", err
	}

	user, err := s.repo.Create(ctx, req.Name, req.Email, passwordHash, "member")
	if err != nil {
		return nil, "", "", err
	}

	accessToken, refreshToken, err := auth.GenerateTokens(
		user.ID,
		user.Email,
		user.Role,
		s.jwtSecret,
		s.jwtSecret,
	)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*User, string, string, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		return nil, "", "", ErrInvalidCredentials
	}

	accessToken, refreshToken, err := auth.GenerateTokens(
		user.ID,
		user.Email,
		user.Role,
		s.jwtSecret,
		s.jwtSecret,
	)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *service) GetByID(ctx context.Context, userID int) (*User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken, jwtSecret string) (string, *User, error) {
	_, claims, err := auth.RefreshAccessToken(refreshToken, jwtSecret, jwtSecret)
	if err != nil {
		return "", nil, err
	}

	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		return "", nil, ErrUserNotFound
	}

	newAccessToken, err := auth.GenerateAccessToken(user.ID, user.Email, user.Role, jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return newAccessToken, user, nil
}
