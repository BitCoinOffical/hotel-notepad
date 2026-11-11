package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserSession(t *testing.T) {
	expectedSession := &models.Session{
		UserID:    "user-123",
		Username:  "test_user",
		Role:      "user",
		CreatedAt: time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC),
	}

	testCases := []struct {
		caseName string
		prepare  func(c *gin.Context)
		want     *models.Session
		wantErr  error
	}{
		{
			caseName: "success",
			prepare: func(c *gin.Context) {
				c.Set("session", expectedSession)
			},
			want: expectedSession,
		},
		{
			caseName: "session not found",
			wantErr:  domain.ErrNotFound,
		},
		{
			caseName: "incorrect session type",
			prepare: func(c *gin.Context) {
				c.Set("session", "invalid session")
			},
			wantErr: domain.ErrIncorrectType,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())

			if tc.prepare != nil {
				tc.prepare(c)
			}

			session, err := GetUserSession(c)

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
