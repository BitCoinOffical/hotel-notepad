package services

import (
	"context"
	"fmt"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/models"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/interfaces/repo"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/interfaces/sessions"
	jwtpkg "github.com/BitCoinOffical/hotel-notepad/auth-service/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo    *repo.AuthRepo
	session *sessions.SessionRepo
}

func NewAuthService(repo *repo.AuthRepo, session *sessions.SessionRepo) *AuthService {
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
		return "", fmt.Errorf("s.repo.SaveUser")
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

func (s *AuthService) Login()
