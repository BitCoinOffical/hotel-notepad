package repo

import (
	"context"
	"testing"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupAuthRepo(t *testing.T) (*AuthRepo, context.Context) {
	t.Helper()

	ctx := context.Background()

	containerCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	container, err := postgres.Run(
		containerCtx,
		"postgres:16-alpine",
		postgres.WithDatabase("hotel_notepad_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)

	testcontainers.CleanupContainer(t, container)

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	t.Cleanup(pool.Close)

	require.NoError(t, pool.Ping(ctx))

	_, err = pool.Exec(ctx, `
		CREATE TYPE user_role AS ENUM ('user', 'admin')
	`)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		CREATE TABLE users (
			id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
			username VARCHAR(16) UNIQUE NOT NULL,
			password_hash VARCHAR NOT NULL,
			role user_role DEFAULT 'user',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	return NewAuthRepo(pool), ctx
}

func TestSaveUser(t *testing.T) {
	authRepo, ctx := setupAuthRepo(t)

	testCases := []struct {
		caseName string
		data     struct {
			username     string
			passwordHash string
		}
		prepare  func(t *testing.T)
		wantErr  bool
		wantCode string
	}{
		{
			caseName: "success",
			data: struct {
				username     string
				passwordHash string
			}{
				username:     "test_user",
				passwordHash: "hashed_password",
			},
		},
		{
			caseName: "username already exists",
			data: struct {
				username     string
				passwordHash string
			}{
				username:     "duplicate",
				passwordHash: "hashed_password",
			},
			prepare: func(t *testing.T) {
				_, err := authRepo.SaveUser(ctx, &models.User{
					UserName:     "duplicate",
					PasswordHash: "hashed_password",
				})
				require.NoError(t, err)
			},
			wantErr:  true,
			wantCode: "23505",
		},
		{
			caseName: "username exceeds maximum length",
			data: struct {
				username     string
				passwordHash string
			}{
				username:     "12345678901234567",
				passwordHash: "hashed_password",
			},
			wantErr:  true,
			wantCode: "22001",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare(t)
			}

			userID, err := authRepo.SaveUser(ctx, &models.User{
				UserName:     tc.data.username,
				PasswordHash: tc.data.passwordHash,
			})

			if tc.wantErr {
				require.Error(t, err)
				assert.Empty(t, userID)

				var pgErr *pgconn.PgError
				require.ErrorAs(t, err, &pgErr)
				assert.Equal(t, tc.wantCode, pgErr.Code)

				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, userID)
		})
	}
}

func TestGetUserByUserName(t *testing.T) {
	authRepo, ctx := setupAuthRepo(t)

	testCases := []struct {
		caseName string
		data     struct {
			username string
		}
		prepare func(t *testing.T)
		wantErr bool
	}{
		{
			caseName: "success",
			data: struct {
				username string
			}{
				username: "test_user",
			},
			prepare: func(t *testing.T) {
				_, err := authRepo.SaveUser(ctx, &models.User{
					UserName:     "test_user",
					PasswordHash: "hashed_password",
				})
				require.NoError(t, err)
			},
		},
		{
			caseName: "user not found",
			data: struct {
				username string
			}{
				username: "unknown_user",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare(t)
			}

			user, err := authRepo.GetUserByUserName(ctx, tc.data.username)

			if tc.wantErr {
				require.ErrorIs(t, err, pgx.ErrNoRows)
				assert.Nil(t, user)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, user)

			assert.NotEmpty(t, user.ID)
			assert.Equal(t, "test_user", user.UserName)
			assert.Equal(t, "hashed_password", user.PasswordHash)
			assert.Equal(t, "user", user.Role)
			assert.False(t, user.CreatedAt.IsZero())
			assert.False(t, user.UpdatedAt.IsZero())
		})
	}
}

func TestCheckUserExists(t *testing.T) {
	authRepo, ctx := setupAuthRepo(t)

	testCases := []struct {
		caseName string
		data     struct {
			username string
		}
		prepare func(t *testing.T)
		want    bool
	}{
		{
			caseName: "user exists",
			data: struct {
				username string
			}{
				username: "test_user",
			},
			prepare: func(t *testing.T) {
				_, err := authRepo.SaveUser(ctx, &models.User{
					UserName:     "test_user",
					PasswordHash: "hashed_password",
				})
				require.NoError(t, err)
			},
			want: true,
		},
		{
			caseName: "user does not exist",
			data: struct {
				username string
			}{
				username: "unknown_user",
			},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare(t)
			}

			exists, err := authRepo.CheckUserExists(ctx, tc.data.username)

			require.NoError(t, err)
			assert.Equal(t, tc.want, exists)
		})
	}
}
