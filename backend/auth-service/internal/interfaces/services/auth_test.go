package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/models"
	repo_mocks "github.com/BitCoinOffical/hotel-notepad/auth-service/internal/interfaces/services/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	repoErr := errors.New("database error")
	sessionErr := errors.New("redis error")

	testCases := []struct {
		caseName string
		data     struct {
			req *dto.RegisterDTO
		}
		prepare func(
			t *testing.T,
			authRepo *repo_mocks.MockAuthRepo,
			sessionRepo *repo_mocks.MockSessionRepo,
		)
		wantErr error
	}{
		{
			caseName: "success",
			data: struct {
				req *dto.RegisterDTO
			}{
				req: &dto.RegisterDTO{
					UserName:      "test_user",
					Password:      "password123",
					PasswordRetry: "password123",
				},
			},
			prepare: func(
				t *testing.T,
				authRepo *repo_mocks.MockAuthRepo,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				authRepo.EXPECT().
					CheckUserExists(gomock.Any(), "test_user").
					Return(false, nil)

				authRepo.EXPECT().
					SaveUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, user *models.User) (string, error) {
						assert.Equal(t, "test_user", user.UserName)
						assert.NoError(t, bcrypt.CompareHashAndPassword(
							[]byte(user.PasswordHash),
							[]byte("password123"),
						))

						return "user-123", nil
					})

				sessionRepo.EXPECT().
					SaveSession(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, sessionID string, session *models.Session) error {
						assert.NotEmpty(t, sessionID)
						assert.Equal(t, "user-123", session.UserID)
						assert.Equal(t, "test_user", session.Username)
						assert.Equal(t, "user", session.Role)
						assert.False(t, session.CreatedAt.IsZero())

						return nil
					})
			},
		},
		{
			caseName: "passwords do not match",
			data: struct {
				req *dto.RegisterDTO
			}{
				req: &dto.RegisterDTO{
					UserName:      "test_user",
					Password:      "password123",
					PasswordRetry: "password456",
				},
			},
			wantErr: domain.ErrNotMatch,
		},
		{
			caseName: "check user exists error",
			data: struct {
				req *dto.RegisterDTO
			}{
				req: &dto.RegisterDTO{
					UserName:      "test_user",
					Password:      "password123",
					PasswordRetry: "password123",
				},
			},
			prepare: func(
				t *testing.T,
				authRepo *repo_mocks.MockAuthRepo,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				authRepo.EXPECT().
					CheckUserExists(gomock.Any(), "test_user").
					Return(false, repoErr)
			},
			wantErr: repoErr,
		},
		{
			caseName: "user already exists",
			data: struct {
				req *dto.RegisterDTO
			}{
				req: &dto.RegisterDTO{
					UserName:      "test_user",
					Password:      "password123",
					PasswordRetry: "password123",
				},
			},
			prepare: func(
				t *testing.T,
				authRepo *repo_mocks.MockAuthRepo,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				authRepo.EXPECT().
					CheckUserExists(gomock.Any(), "test_user").
					Return(true, nil)
			},
			wantErr: domain.ErrUserConflict,
		},
		{
			caseName: "password exceeds maximum length",
			data: struct {
				req *dto.RegisterDTO
			}{
				req: &dto.RegisterDTO{
					UserName:      "test_user",
					Password:      strings.Repeat("a", 73),
					PasswordRetry: strings.Repeat("a", 73),
				},
			},
			prepare: func(
				t *testing.T,
				authRepo *repo_mocks.MockAuthRepo,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				authRepo.EXPECT().
					CheckUserExists(gomock.Any(), "test_user").
					Return(false, nil)
			},
			wantErr: bcrypt.ErrPasswordTooLong,
		},
		{
			caseName: "save user error",
			data: struct {
				req *dto.RegisterDTO
			}{
				req: &dto.RegisterDTO{
					UserName:      "test_user",
					Password:      "password123",
					PasswordRetry: "password123",
				},
			},
			prepare: func(
				t *testing.T,
				authRepo *repo_mocks.MockAuthRepo,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				authRepo.EXPECT().
					CheckUserExists(gomock.Any(), "test_user").
					Return(false, nil)

				authRepo.EXPECT().
					SaveUser(gomock.Any(), gomock.Any()).
					Return("", repoErr)
			},
			wantErr: repoErr,
		},
		{
			caseName: "save session error",
			data: struct {
				req *dto.RegisterDTO
			}{
				req: &dto.RegisterDTO{
					UserName:      "test_user",
					Password:      "password123",
					PasswordRetry: "password123",
				},
			},
			prepare: func(
				t *testing.T,
				authRepo *repo_mocks.MockAuthRepo,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				authRepo.EXPECT().
					CheckUserExists(gomock.Any(), "test_user").
					Return(false, nil)

				authRepo.EXPECT().
					SaveUser(gomock.Any(), gomock.Any()).
					Return("user-123", nil)

				sessionRepo.EXPECT().
					SaveSession(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(sessionErr)
			},
			wantErr: sessionErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			authRepo := repo_mocks.NewMockAuthRepo(ctrl)
			sessionRepo := repo_mocks.NewMockSessionRepo(ctrl)

			if tc.prepare != nil {
				tc.prepare(t, authRepo, sessionRepo)
			}

			authService := NewAuthService(authRepo, sessionRepo)

			sessionID, err := authService.Register(context.Background(), tc.data.req)

			if tc.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, sessionID)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, sessionID)
		})
	}
}

func TestLogin(t *testing.T) {
	repoErr := errors.New("database error")
	sessionErr := errors.New("redis error")

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte("password123"),
		bcrypt.DefaultCost,
	)
	require.NoError(t, err)

	testCases := []struct {
		caseName string
		data     struct {
			req *dto.LoginDTO
		}
		prepare func(
			t *testing.T,
			authRepo *repo_mocks.MockAuthRepo,
			sessionRepo *repo_mocks.MockSessionRepo,
		)
		wantErr error
	}{
		{
			caseName: "success",
			data: struct {
				req *dto.LoginDTO
			}{
				req: &dto.LoginDTO{
					UserName: "test_user",
					Password: "password123",
				},
			},
			prepare: func(
				t *testing.T,
				authRepo *repo_mocks.MockAuthRepo,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				authRepo.EXPECT().
					GetUserByUserName(gomock.Any(), "test_user").
					Return(&models.User{
						ID:           "user-123",
						UserName:     "test_user",
						PasswordHash: string(passwordHash),
						Role:         "admin",
					}, nil)

				sessionRepo.EXPECT().
					SaveSession(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, sessionID string, session *models.Session) error {
						assert.NotEmpty(t, sessionID)
						assert.Equal(t, "user-123", session.UserID)
						assert.Equal(t, "test_user", session.Username)
						assert.Equal(t, "admin", session.Role)
						assert.False(t, session.CreatedAt.IsZero())

						return nil
					})
			},
		},
		{
			caseName: "get user error",
			data: struct {
				req *dto.LoginDTO
			}{
				req: &dto.LoginDTO{
					UserName: "test_user",
					Password: "password123",
				},
			},
			prepare: func(
				t *testing.T,
				authRepo *repo_mocks.MockAuthRepo,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				authRepo.EXPECT().
					GetUserByUserName(gomock.Any(), "test_user").
					Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
		{
			caseName: "invalid password",
			data: struct {
				req *dto.LoginDTO
			}{
				req: &dto.LoginDTO{
					UserName: "test_user",
					Password: "wrong_password",
				},
			},
			prepare: func(
				t *testing.T,
				authRepo *repo_mocks.MockAuthRepo,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				authRepo.EXPECT().
					GetUserByUserName(gomock.Any(), "test_user").
					Return(&models.User{
						ID:           "user-123",
						UserName:     "test_user",
						PasswordHash: string(passwordHash),
						Role:         "user",
					}, nil)
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			caseName: "invalid password hash",
			data: struct {
				req *dto.LoginDTO
			}{
				req: &dto.LoginDTO{
					UserName: "test_user",
					Password: "password123",
				},
			},
			prepare: func(
				t *testing.T,
				authRepo *repo_mocks.MockAuthRepo,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				authRepo.EXPECT().
					GetUserByUserName(gomock.Any(), "test_user").
					Return(&models.User{
						ID:           "user-123",
						UserName:     "test_user",
						PasswordHash: "invalid_hash",
						Role:         "user",
					}, nil)
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			caseName: "save session error",
			data: struct {
				req *dto.LoginDTO
			}{
				req: &dto.LoginDTO{
					UserName: "test_user",
					Password: "password123",
				},
			},
			prepare: func(
				t *testing.T,
				authRepo *repo_mocks.MockAuthRepo,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				authRepo.EXPECT().
					GetUserByUserName(gomock.Any(), "test_user").
					Return(&models.User{
						ID:           "user-123",
						UserName:     "test_user",
						PasswordHash: string(passwordHash),
						Role:         "user",
					}, nil)

				sessionRepo.EXPECT().
					SaveSession(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(sessionErr)
			},
			wantErr: sessionErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			authRepo := repo_mocks.NewMockAuthRepo(ctrl)
			sessionRepo := repo_mocks.NewMockSessionRepo(ctrl)

			if tc.prepare != nil {
				tc.prepare(t, authRepo, sessionRepo)
			}

			authService := NewAuthService(authRepo, sessionRepo)

			sessionID, err := authService.Login(context.Background(), tc.data.req)

			if tc.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, sessionID)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, sessionID)
		})
	}
}

func TestLogout(t *testing.T) {
	repoErr := errors.New("redis error")

	testCases := []struct {
		caseName string
		data     struct {
			sessionID string
		}
		prepare func(
			t *testing.T,
			sessionRepo *repo_mocks.MockSessionRepo,
		)
		wantErr error
	}{
		{
			caseName: "success",
			data: struct {
				sessionID string
			}{
				sessionID: "session-123",
			},
			prepare: func(
				t *testing.T,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				sessionRepo.EXPECT().
					DeleteSession(gomock.Any(), "session-123").
					Return(nil)
			},
		},
		{
			caseName: "delete session error",
			data: struct {
				sessionID string
			}{
				sessionID: "session-123",
			},
			prepare: func(
				t *testing.T,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				sessionRepo.EXPECT().
					DeleteSession(gomock.Any(), "session-123").
					Return(repoErr)
			},
			wantErr: repoErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			sessionRepo := repo_mocks.NewMockSessionRepo(ctrl)

			if tc.prepare != nil {
				tc.prepare(t, sessionRepo)
			}

			authService := NewAuthService(nil, sessionRepo)

			err := authService.Logout(context.Background(), tc.data.sessionID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGetSession(t *testing.T) {
	repoErr := errors.New("redis error")

	expected := &models.Session{
		UserID:    "user-123",
		Username:  "test_user",
		Role:      "user",
		CreatedAt: time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC),
	}

	testCases := []struct {
		caseName string
		data     struct {
			sessionID string
		}
		prepare func(
			t *testing.T,
			sessionRepo *repo_mocks.MockSessionRepo,
		)
		want    *models.Session
		wantErr error
	}{
		{
			caseName: "success",
			data: struct {
				sessionID string
			}{
				sessionID: "session-123",
			},
			prepare: func(
				t *testing.T,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				sessionRepo.EXPECT().
					GetSession(gomock.Any(), "session-123").
					Return(expected, nil)
			},
			want: expected,
		},
		{
			caseName: "get session error",
			data: struct {
				sessionID string
			}{
				sessionID: "session-123",
			},
			prepare: func(
				t *testing.T,
				sessionRepo *repo_mocks.MockSessionRepo,
			) {
				sessionRepo.EXPECT().
					GetSession(gomock.Any(), "session-123").
					Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			sessionRepo := repo_mocks.NewMockSessionRepo(ctrl)

			if tc.prepare != nil {
				tc.prepare(t, sessionRepo)
			}

			authService := NewAuthService(nil, sessionRepo)

			session, err := authService.GetSession(context.Background(), tc.data.sessionID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, session)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, session)
			assert.Equal(t, tc.want, session)
		})
	}
}
