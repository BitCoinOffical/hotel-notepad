package sessions

import (
	"context"
	"encoding/json/v2"
	"testing"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/models"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupSessionRepo(t *testing.T) (*SessionRepo, *redis.Client, context.Context) {
	t.Helper()

	ctx := context.Background()

	containerCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	container, err := tcredis.Run(containerCtx, "redis:7-alpine")
	require.NoError(t, err)

	testcontainers.CleanupContainer(t, container)

	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	options, err := redis.ParseURL(connStr)
	require.NoError(t, err)

	rdb := redis.NewClient(options)
	t.Cleanup(func() {
		require.NoError(t, rdb.Close())
	})

	require.NoError(t, rdb.Ping(ctx).Err())

	return NewSessionRepo(rdb), rdb, ctx
}

func TestSaveSession(t *testing.T) {
	sessionRepo, rdb, ctx := setupSessionRepo(t)

	testCases := []struct {
		caseName  string
		sessionID string
		session   *models.Session
		prepare   func(t *testing.T)
		cancelCtx bool
		wantErr   error
	}{
		{
			caseName:  "success",
			sessionID: "session1",
			session: &models.Session{
				UserID:    "user1",
				Username:  "test_user",
				Role:      "user",
				CreatedAt: time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC),
			},
		},
		{
			caseName:  "overwrite existing session",
			sessionID: "session2",
			session: &models.Session{
				UserID:    "user2",
				Username:  "updated_user",
				Role:      "admin",
				CreatedAt: time.Date(2026, 11, 11, 11, 0, 0, 0, time.UTC),
			},
			prepare: func(t *testing.T) {
				err := sessionRepo.SaveSession(ctx, "session2", &models.Session{
					UserID:   "user1",
					Username: "old_user",
					Role:     "user",
				})
				require.NoError(t, err)
			},
		},
		{
			caseName:  "context canceled",
			sessionID: "session3",
			session: &models.Session{
				UserID:   "user3",
				Username: "test_user",
				Role:     "user",
			},
			cancelCtx: true,
			wantErr:   context.Canceled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare(t)
			}

			saveCtx := ctx
			if tc.cancelCtx {
				canceledCtx, cancel := context.WithCancel(ctx)
				cancel()
				saveCtx = canceledCtx
			}

			err := sessionRepo.SaveSession(saveCtx, tc.sessionID, tc.session)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)

			value, err := rdb.Get(ctx, sessionPrefix+tc.sessionID).Bytes()
			require.NoError(t, err)

			var got models.Session
			require.NoError(t, json.Unmarshal(value, &got))
			assert.Equal(t, tc.session, &got)

			ttl, err := rdb.TTL(ctx, sessionPrefix+tc.sessionID).Result()
			require.NoError(t, err)
			assert.Greater(t, ttl, time.Duration(0))
			assert.LessOrEqual(t, ttl, SessionTTL)
		})
	}
}

func TestGetSession(t *testing.T) {
	sessionRepo, rdb, ctx := setupSessionRepo(t)

	expected := &models.Session{
		UserID:    "user1",
		Username:  "test_user",
		Role:      "user",
		CreatedAt: time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC),
	}

	testCases := []struct {
		caseName  string
		sessionID string
		prepare   func(t *testing.T)
		want      *models.Session
		wantErr   bool
		wantErrIs error
	}{
		{
			caseName:  "success",
			sessionID: "session1",
			prepare: func(t *testing.T) {
				err := sessionRepo.SaveSession(ctx, "session1", expected)
				require.NoError(t, err)
			},
			want: expected,
		},
		{
			caseName:  "session not found",
			sessionID: "unknown_session",
			wantErr:   true,
			wantErrIs: redis.Nil,
		},
		{
			caseName:  "invalid JSON",
			sessionID: "invalid_session",
			prepare: func(t *testing.T) {
				err := rdb.Set(ctx, sessionPrefix+"invalid_session", "{invalid json", SessionTTL).Err()
				require.NoError(t, err)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare(t)
			}

			session, err := sessionRepo.GetSession(ctx, tc.sessionID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, session)

				if tc.wantErrIs != nil {
					assert.ErrorIs(t, err, tc.wantErrIs)
				}

				return
			}

			require.NoError(t, err)
			require.NotNil(t, session)
			assert.Equal(t, tc.want, session)
		})
	}
}

func TestDeleteSession(t *testing.T) {
	sessionRepo, rdb, ctx := setupSessionRepo(t)

	testCases := []struct {
		caseName  string
		sessionID string
		prepare   func(t *testing.T)
		cancelCtx bool
		wantErr   error
		wantExist int64
	}{
		{
			caseName:  "success",
			sessionID: "session1",
			prepare: func(t *testing.T) {
				err := sessionRepo.SaveSession(ctx, "session1", &models.Session{
					UserID:   "user1",
					Username: "test_user",
					Role:     "user",
				})
				require.NoError(t, err)
			},
			wantExist: 0,
		},
		{
			caseName:  "session not found",
			sessionID: "unknown_session",
			wantExist: 0,
		},
		{
			caseName:  "context canceled",
			sessionID: "session3",
			prepare: func(t *testing.T) {
				err := sessionRepo.SaveSession(ctx, "session3", &models.Session{
					UserID:   "user3",
					Username: "test_user",
					Role:     "user",
				})
				require.NoError(t, err)
			},
			cancelCtx: true,
			wantErr:   context.Canceled,
			wantExist: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare(t)
			}

			deleteCtx := ctx
			if tc.cancelCtx {
				canceledCtx, cancel := context.WithCancel(ctx)
				cancel()
				deleteCtx = canceledCtx
			}

			err := sessionRepo.DeleteSession(deleteCtx, tc.sessionID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}

			exists, err := rdb.Exists(ctx, sessionPrefix+tc.sessionID).Result()
			require.NoError(t, err)
			assert.Equal(t, tc.wantExist, exists)
		})
	}
}
