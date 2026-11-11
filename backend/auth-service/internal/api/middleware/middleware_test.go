package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	srvc_mocks "github.com/BitCoinOffical/hotel-notepad/auth-service/internal/api/middleware/mock"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceErr := errors.New("service error")

	expectedSession := &models.Session{
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
		prepare  func(srvc *srvc_mocks.MockAuthService)
		status   int
		want     *models.Session
		wantNext bool
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
					GetSession(gomock.Any(), "session-123").
					Return(expectedSession, nil)
			},
			status:   http.StatusOK,
			want:     expectedSession,
			wantNext: true,
		},
		{
			caseName: "cookie not found",
			status:   http.StatusUnauthorized,
		},
		{
			caseName: "get session error",
			data: struct {
				sessionID string
			}{
				sessionID: "session-123",
			},
			prepare: func(srvc *srvc_mocks.MockAuthService) {
				srvc.EXPECT().
					GetSession(gomock.Any(), "session-123").
					Return(nil, serviceErr)
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

			middleware := NewMiddleware(zaptest.NewLogger(t), srvc)

			var gotSession *models.Session
			var nextCalled bool

			router := gin.New()
			router.Use(middleware.AuthMiddleware())
			router.GET("/protected", func(c *gin.Context) {
				nextCalled = true

				session, exists := c.Get("session")
				require.True(t, exists)

				gotSession, _ = session.(*models.Session)

				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)

			if tc.data.sessionID != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_id",
					Value: tc.data.sessionID,
				})
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)
			assert.Equal(t, tc.wantNext, nextCalled)

			if tc.want != nil {
				require.NotNil(t, gotSession)
				assert.Equal(t, tc.want, gotSession)
			}
		})
	}
}
