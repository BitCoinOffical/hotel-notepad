package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	repo_mocks "github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/api/middleware/mock"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repoErr := errors.New("redis error")

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
		prepare func(repo *repo_mocks.MockSessionRepo)
		status  int
		want    *models.Session
	}{
		{
			caseName: "success",
			data: struct {
				sessionID string
			}{
				sessionID: "session-123",
			},
			prepare: func(repo *repo_mocks.MockSessionRepo) {
				repo.EXPECT().
					GetSession(gomock.Any(), "session-123").
					Return(expectedSession, nil)
			},
			status: http.StatusOK,
			want:   expectedSession,
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
			prepare: func(repo *repo_mocks.MockSessionRepo) {
				repo.EXPECT().
					GetSession(gomock.Any(), "session-123").
					Return(nil, repoErr)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repo_mocks.NewMockSessionRepo(ctrl)

			if tc.prepare != nil {
				tc.prepare(repo)
			}

			middleware := NewMiddleware(zaptest.NewLogger(t), repo)

			var gotSession *models.Session
			var nextCalled bool

			router := gin.New()
			router.Use(middleware.AuthMiddleware())

			router.GET("/protected", func(c *gin.Context) {
				nextCalled = true

				session, err := GetUserSession(c)
				require.NoError(t, err)

				gotSession = session
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

			if tc.want != nil {
				require.True(t, nextCalled)
				require.NotNil(t, gotSession)
				assert.Equal(t, tc.want, gotSession)
			} else {
				assert.False(t, nextCalled)
				assert.Nil(t, gotSession)
			}
		})
	}
}
