package services

import (
	"context"
	"fmt"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/models"
	jwtpkg "github.com/BitCoinOffical/hotel-notepad/auth-service/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo    AuthRepo
	session SessionRepo
}

func NewAuthService(repo AuthRepo, session SessionRepo) *AuthService {
	return &AuthService{repo: repo, session: session}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterDTO) (string, error) {
	if req.Password != req.PasswordRetry {
		return "", domain.ErrNotMatch
	}

	exists, err := s.repo.CheckUserExists(ctx, req.UserName)
	if err != nil {
		return "", fmt.Errorf("s.repo.CheckUserExists: %w", err)
	}

	if exists {
		return "", domain.ErrUserConflict
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
	}

	user := models.User{
		UserName:     req.UserName,
		PasswordHash: string(pass),
	}

	userID, err := s.repo.SaveUser(ctx, &user)
	if err != nil {
		return "", fmt.Errorf("s.repo.SaveUser: %w", err)
	}
	sessionID, err := jwtpkg.GenerateRandomString()
	if err != nil {
		return "", fmt.Errorf("jwtpkg.GenerateRandomString: %w", err)
	}

	session := models.Session{
		UserID:    userID,
		Username:  req.UserName,
		Role:      "user",
		CreatedAt: time.Now(),
	}

	if err := s.session.SaveSession(ctx, sessionID, &session); err != nil {
		return "", fmt.Errorf("s.session.SaveSession: %w", err)
	}

	return sessionID, nil

}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginDTO) (string, error) {
	user, err := s.repo.GetUserByUserName(ctx, req.UserName)
	if err != nil {
		return "", fmt.Errorf("s.repo.GetUserByUserName: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", fmt.Errorf("bcrypt.CompareHashAndPassword: %w", domain.ErrInvalidCredentials)
	}

	sessionID, err := jwtpkg.GenerateRandomString()
	if err != nil {
		return "", fmt.Errorf("jwtpkg.GenerateRandomString: %w", err)
	}

	session := models.Session{
		UserID:    user.ID,
		Username:  user.UserName,
		Role:      user.Role,
		CreatedAt: time.Now(),
	}

	if err := s.session.SaveSession(ctx, sessionID, &session); err != nil {
		return "", fmt.Errorf("s.session.SaveSession: %w", err)
	}

	return sessionID, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	if err := s.session.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("s.session.DeleteSession: %w", err)
	}
	return nil
}

func (s *AuthService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := s.session.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("s.session.GetSession: %w", err)
	}
	return session, nil
}
