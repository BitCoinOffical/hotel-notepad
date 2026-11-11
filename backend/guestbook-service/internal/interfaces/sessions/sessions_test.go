package sessions

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/models"
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
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	options, err := redis.ParseURL(connStr)
	require.NoError(t, err)

	rdb := redis.NewClient(options)
	t.Cleanup(func() {
		assert.NoError(t, rdb.Close())
	})

	require.NoError(t, rdb.Ping(ctx).Err())

	return NewSessionRepo(rdb), rdb, ctx
}

func TestGetSession(t *testing.T) {
	testCases := []struct {
		caseName string
		data     struct {
			sessionID string
		}
		prepare   func(t *testing.T, rdb *redis.Client, ctx context.Context)
		cancelCtx bool
		wantErr   error
	}{
		{
			caseName: "success",
			data: struct {
				sessionID string
			}{
				sessionID: "session-123",
			},
			prepare: func(t *testing.T, rdb *redis.Client, ctx context.Context) {
				err := rdb.Set(ctx, sessionPrefix+"session-123", `{}`, 0).Err()
				require.NoError(t, err)
			},
		},
		{
			caseName: "session not found",
			data: struct {
				sessionID string
			}{
				sessionID: "unknown-session",
			},
			wantErr: redis.Nil,
		},
		{
			caseName: "invalid JSON",
			data: struct {
				sessionID string
			}{
				sessionID: "invalid-session",
			},
			prepare: func(t *testing.T, rdb *redis.Client, ctx context.Context) {
				err := rdb.Set(ctx, sessionPrefix+"invalid-session", `{invalid json`, 0).Err()
				require.NoError(t, err)
			},
		},
		{
			caseName: "context canceled",
			data: struct {
				sessionID string
			}{
				sessionID: "session-123",
			},
			cancelCtx: true,
			wantErr:   context.Canceled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			sessionRepo, rdb, ctx := setupSessionRepo(t)

			if tc.prepare != nil {
				tc.prepare(t, rdb, ctx)
			}

			getCtx := ctx
			if tc.cancelCtx {
				canceledCtx, cancel := context.WithCancel(ctx)
				cancel()
				getCtx = canceledCtx
			}

			session, err := sessionRepo.GetSession(getCtx, tc.data.sessionID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, session)
				return
			}

			if err != nil {
				require.Error(t, err)

				var jsonErr *json.SyntaxError
				require.ErrorAs(t, err, &jsonErr)

				assert.Nil(t, session)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, session)
			assert.Equal(t, &models.Session{}, session)
		})
	}
}
