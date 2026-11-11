package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	srvc_mocks "github.com/BitCoinOffical/hotel-notepad/auth-service/internal/api/handlers/mock"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	registerBody, err := json.Marshal(&dto.RegisterDTO{
		UserName:      "test_user",
		Password:      "password123",
		PasswordRetry: "password123",
	})
	require.NoError(t, err)

	serviceErr := errors.New("service error")

	testCases := []struct {
		caseName string
		data     struct {
			body string
		}
		prepare func(srvc *srvc_mocks.MockAuthService)
		status  int
		cookie  string
	}{
		{
			caseName: "success",
			data: struct {
				body string
			}{
				body: string(registerBody),
			},
			prepare: func(srvc *srvc_mocks.MockAuthService) {
				srvc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, req *dto.RegisterDTO) (string, error) {
						assert.Equal(t, "test_user", req.UserName)
						assert.Equal(t, "password123", req.Password)
						assert.Equal(t, "password123", req.PasswordRetry)

						return "session-123", nil
					})
			},
			status: http.StatusCreated,
			cookie: "session-123",
		},
		{
			caseName: "invalid request",
			data: struct {
				body string
			}{
				body: "{invalid json",
			},
			status: http.StatusBadRequest,
		},
		{
			caseName: "passwords do not match",
			data: struct {
				body string
			}{
				body: string(registerBody),
			},
			prepare: func(srvc *srvc_mocks.MockAuthService) {
				srvc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return("", domain.ErrNotMatch)
			},
			status: http.StatusBadRequest,
		},
		{
			caseName: "user already exists",
			data: struct {
				body string
			}{
				body: string(registerBody),
			},
			prepare: func(srvc *srvc_mocks.MockAuthService) {
				srvc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return("", domain.ErrUserConflict)
			},
			status: http.StatusConflict,
		},
		{
			caseName: "service error",
			data: struct {
				body string
			}{
				body: string(registerBody),
			},
			prepare: func(srvc *srvc_mocks.MockAuthService) {
				srvc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return("", serviceErr)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockAuthService(ctrl)

			if tc.prepare != nil {
				tc.prepare(srvc)
			}

			handler := NewAuthHandler(srvc, zaptest.NewLogger(t))

			router := gin.New()
			router.POST("/register", handler.Register)

			req := httptest.NewRequest(
				http.MethodPost,
				"/register",
				strings.NewReader(tc.data.body),
			)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)

			cookies := rec.Result().Cookies()
			if tc.cookie != "" {
				require.Len(t, cookies, 1)

				assert.Equal(t, "session_id", cookies[0].Name)
				assert.Equal(t, tc.cookie, cookies[0].Value)
				assert.Equal(t, "/", cookies[0].Path)
				assert.Equal(t, 2678400, cookies[0].MaxAge)
				assert.True(t, cookies[0].Secure)
				assert.True(t, cookies[0].HttpOnly)
				assert.Equal(t, http.SameSiteStrictMode, cookies[0].SameSite)
			} else {
				assert.Empty(t, cookies)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	loginBody, err := json.Marshal(&dto.LoginDTO{
		UserName: "test_user",
		Password: "password123",
	})
	require.NoError(t, err)

	serviceErr := errors.New("service error")

	testCases := []struct {
		caseName string
		data     struct {
			body string
		}
		prepare func(srvc *srvc_mocks.MockAuthService)
		status  int
		cookie  string
	}{
		{
			caseName: "success",
			data: struct {
				body string
			}{
				body: string(loginBody),
			},
			prepare: func(srvc *srvc_mocks.MockAuthService) {
				srvc.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, req *dto.LoginDTO) (string, error) {
						assert.Equal(t, "test_user", req.UserName)
						assert.Equal(t, "password123", req.Password)

						return "session-123", nil
					})
			},
			status: http.StatusCreated,
			cookie: "session-123",
		},
		{
			caseName: "invalid request",
			data: struct {
				body string
			}{
				body: "{invalid json",
			},
			status: http.StatusBadRequest,
		},
		{
			caseName: "service error",
			data: struct {
				body string
			}{
				body: string(loginBody),
			},
			prepare: func(srvc *srvc_mocks.MockAuthService) {
				srvc.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return("", serviceErr)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockAuthService(ctrl)

			if tc.prepare != nil {
				tc.prepare(srvc)
			}

			handler := NewAuthHandler(srvc, zaptest.NewLogger(t))

			router := gin.New()
			router.POST("/login", handler.Login)

			req := httptest.NewRequest(
				http.MethodPost,
				"/login",
				strings.NewReader(tc.data.body),
			)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)

			cookies := rec.Result().Cookies()
			if tc.cookie != "" {
				require.Len(t, cookies, 1)

				assert.Equal(t, "session_id", cookies[0].Name)
				assert.Equal(t, tc.cookie, cookies[0].Value)
				assert.Equal(t, "/", cookies[0].Path)
				assert.Equal(t, 2678400, cookies[0].MaxAge)
				assert.True(t, cookies[0].Secure)
				assert.True(t, cookies[0].HttpOnly)
				assert.Equal(t, http.SameSiteStrictMode, cookies[0].SameSite)
			} else {
				assert.Empty(t, cookies)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceErr := errors.New("service error")

	testCases := []struct {
		caseName string
		data     struct {
			sessionID string
		}
		prepare func(srvc *srvc_mocks.MockAuthService)
		status  int
	}{
		{
			caseName: "success",
			data: struct {
				sessionID string
			}{
				sessionID: "session-123",
			},
			prepare: func(srvc *srvc_mocks.MockAuthService) {
				srvc.EXPECT().
					Logout(gomock.Any(), "session-123").
					Return(nil)
			},
			status: http.StatusNoContent,
		},
		{
			caseName: "cookie not found",
			data: struct {
				sessionID string
			}{},
			status: http.StatusUnauthorized,
		},
		{
			caseName: "service error",
			data: struct {
				sessionID string
			}{
				sessionID: "session-123",
			},
			prepare: func(srvc *srvc_mocks.MockAuthService) {
				srvc.EXPECT().
					Logout(gomock.Any(), "session-123").
					Return(serviceErr)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockAuthService(ctrl)

			if tc.prepare != nil {
				tc.prepare(srvc)
			}

			handler := NewAuthHandler(srvc, zaptest.NewLogger(t))

			router := gin.New()
			router.POST("/logout", handler.Logout)

			req := httptest.NewRequest(http.MethodPost, "/logout", nil)

			if tc.data.sessionID != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_id",
					Value: tc.data.sessionID,
				})
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)
		})
	}
}
