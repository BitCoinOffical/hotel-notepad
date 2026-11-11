package repo

import (
	"context"
	"testing"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/models"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupGuestBookRepo(t *testing.T) (*GuestBookRepo, *pgxpool.Pool, context.Context) {
	t.Helper()

	ctx := context.Background()

	containerCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	container, err := postgres.Run(
		containerCtx,
		"postgres:16-alpine",
		postgres.WithDatabase("notebook_test"),
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
		CREATE TABLE guests (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			mid_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			note TEXT NOT NULL,
			phone BIGINT NOT NULL,
			is_delete BOOLEAN NOT NULL DEFAULT FALSE,
			arrival_date TIMESTAMPTZ NOT NULL,
			departure_date TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	return NewGuestBookRepo(pool), pool, ctx
}

func testGuest(id, userID, firstName string, isDelete bool, updatedAt time.Time) models.Guest {
	var note = "Guest note"
	return models.Guest{
		GuestID:       id,
		UserID:        userID,
		FirstName:     firstName,
		MidName:       "Ivanovich",
		LastName:      "Petrov",
		Note:          &note,
		Phone:         1234567890,
		IsDelete:      isDelete,
		ArrivalDate:   time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC),
		DepartureDate: time.Date(2026, 12, 10, 12, 0, 0, 0, time.UTC),
		CreatedAt:     updatedAt,
		UpdatedAt:     updatedAt,
	}
}

func insertTestGuest(t *testing.T, pool *pgxpool.Pool, ctx context.Context, guest models.Guest) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		INSERT INTO guests (
			id, user_id, first_name, mid_name, last_name,
			note, phone, is_delete, arrival_date, departure_date,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		guest.GuestID,
		guest.UserID,
		guest.FirstName,
		guest.MidName,
		guest.LastName,
		guest.Note,
		guest.Phone,
		guest.IsDelete,
		guest.ArrivalDate,
		guest.DepartureDate,
		guest.CreatedAt,
		guest.UpdatedAt,
	)
	require.NoError(t, err)
}

func TestCreateEntry(t *testing.T) {
	var note = "Guest note"
	guest := models.Guest{
		UserID:        "00000000-0000-0000-0000-000000000001",
		FirstName:     "Ivan",
		MidName:       "Ivanovich",
		LastName:      "Petrov",
		Note:          &note,
		Phone:         1234567890,
		ArrivalDate:   time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC),
		DepartureDate: time.Date(2026, 12, 10, 12, 0, 0, 0, time.UTC),
	}

	testCases := []struct {
		caseName  string
		data      models.Guest
		cancelCtx bool
		wantErr   error
		wantCode  string
	}{
		{
			caseName: "success",
			data:     guest,
		},
		{
			caseName: "invalid user ID",
			data: models.Guest{
				UserID:        "invalid-uuid",
				FirstName:     "Ivan",
				MidName:       "Ivanovich",
				LastName:      "Petrov",
				Note:          &note,
				Phone:         1234567890,
				ArrivalDate:   guest.ArrivalDate,
				DepartureDate: guest.DepartureDate,
			},
			wantCode: "22P02",
		},
		{
			caseName:  "context canceled",
			data:      guest,
			cancelCtx: true,
			wantErr:   context.Canceled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			guestRepo, pool, ctx := setupGuestBookRepo(t)

			createCtx := ctx
			if tc.cancelCtx {
				canceledCtx, cancel := context.WithCancel(ctx)
				cancel()
				createCtx = canceledCtx
			}

			err := guestRepo.CreateEntry(createCtx, &tc.data)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			if tc.wantCode != "" {
				require.Error(t, err)

				var pgErr *pgconn.PgError
				require.ErrorAs(t, err, &pgErr)
				assert.Equal(t, tc.wantCode, pgErr.Code)

				return
			}

			require.NoError(t, err)

			var firstName, midName, lastName, note string
			var phone int
			var arrivalDate, departureDate time.Time

			err = pool.QueryRow(ctx, `
				SELECT first_name, mid_name, last_name, note,
					phone, arrival_date, departure_date
				FROM guests
				WHERE user_id = $1
			`, tc.data.UserID).Scan(
				&firstName,
				&midName,
				&lastName,
				&note,
				&phone,
				&arrivalDate,
				&departureDate,
			)
			require.NoError(t, err)

			assert.Equal(t, tc.data.FirstName, firstName)
			assert.Equal(t, tc.data.MidName, midName)
			assert.Equal(t, tc.data.LastName, lastName)
			assert.Equal(t, tc.data.Note, &note)
			assert.Equal(t, tc.data.Phone, phone)
			assert.True(t, tc.data.ArrivalDate.Equal(arrivalDate))
			assert.True(t, tc.data.DepartureDate.Equal(departureDate))
		})
	}
}

func TestGetEntries(t *testing.T) {
	userID := "00000000-0000-0000-0000-000000000001"
	oldTime := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC)

	testCases := []struct {
		caseName string
		data     struct {
			userID string
		}
		prepare func(t *testing.T, pool *pgxpool.Pool, ctx context.Context)
		want    []string
		wantErr error
	}{
		{
			caseName: "success",
			data: struct {
				userID string
			}{
				userID: userID,
			},
			prepare: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000011",
					userID,
					"Old",
					false,
					oldTime,
				))

				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000012",
					userID,
					"New",
					false,
					newTime,
				))

				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000013",
					userID,
					"Deleted",
					true,
					newTime,
				))
			},
			want: []string{"New", "Old"},
		},
		{
			caseName: "empty list",
			data: struct {
				userID string
			}{
				userID: userID,
			},
			want: []string{},
		},
		{
			caseName: "context canceled",
			data: struct {
				userID string
			}{
				userID: userID,
			},
			wantErr: context.Canceled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			guestRepo, pool, ctx := setupGuestBookRepo(t)

			if tc.prepare != nil {
				tc.prepare(t, pool, ctx)
			}

			queryCtx := ctx
			if tc.wantErr != nil {
				canceledCtx, cancel := context.WithCancel(ctx)
				cancel()
				queryCtx = canceledCtx
			}

			guests, err := guestRepo.GetEntries(queryCtx, tc.data.userID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, guests)
				return
			}

			require.NoError(t, err)
			require.Len(t, guests, len(tc.want))

			for i, guest := range guests {
				assert.Equal(t, tc.want[i], guest.FirstName)
				assert.Equal(t, tc.data.userID, guest.UserID)
			}
		})
	}
}

func TestGetEntryByID(t *testing.T) {
	userID := "00000000-0000-0000-0000-000000000001"
	otherUserID := "00000000-0000-0000-0000-000000000002"

	testCases := []struct {
		caseName string
		data     struct {
			id     string
			userID string
		}
		prepare func(t *testing.T, pool *pgxpool.Pool, ctx context.Context)
		want    string
		wantErr bool
	}{
		{
			caseName: "success",
			data: struct {
				id     string
				userID string
			}{
				id:     "00000000-0000-0000-0000-000000000011",
				userID: userID,
			},
			prepare: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000011",
					userID,
					"Ivan",
					false,
					time.Now(),
				))
			},
			want: "Ivan",
		},
		{
			caseName: "entry not found",
			data: struct {
				id     string
				userID string
			}{
				id:     "00000000-0000-0000-0000-000000000099",
				userID: userID,
			},
			wantErr: true,
		},
		{
			caseName: "entry belongs to another user",
			data: struct {
				id     string
				userID string
			}{
				id:     "00000000-0000-0000-0000-000000000012",
				userID: otherUserID,
			},
			prepare: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000012",
					userID,
					"Ivan",
					false,
					time.Now(),
				))
			},
			wantErr: true,
		},
		{
			caseName: "entry deleted",
			data: struct {
				id     string
				userID string
			}{
				id:     "00000000-0000-0000-0000-000000000013",
				userID: userID,
			},
			prepare: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000013",
					userID,
					"Ivan",
					true,
					time.Now(),
				))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			guestRepo, pool, ctx := setupGuestBookRepo(t)

			if tc.prepare != nil {
				tc.prepare(t, pool, ctx)
			}

			guest, err := guestRepo.GetEntryByID(ctx, tc.data.id, tc.data.userID)

			if tc.wantErr {
				require.ErrorIs(t, err, domain.ErrNotFound)
				assert.Nil(t, guest)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, guest)

			assert.Equal(t, tc.data.id, guest.GuestID)
			assert.Equal(t, tc.data.userID, guest.UserID)
			assert.Equal(t, tc.want, guest.FirstName)
			assert.Equal(t, "Ivanovich", guest.MidName)
			assert.Equal(t, "Petrov", guest.LastName)
			assert.False(t, guest.CreatedAt.IsZero())
			assert.False(t, guest.UpdatedAt.IsZero())
		})
	}
}

func TestUpdateEntry(t *testing.T) {
	userID := "00000000-0000-0000-0000-000000000001"
	otherUserID := "00000000-0000-0000-0000-000000000002"

	oldTime := time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)

	newFirstName := "Updated"
	newNote := "Updated note"
	newPhone := 987654321

	testCases := []struct {
		caseName string
		data     struct {
			id     string
			userID string
			guest  *models.UpdateGuest
		}
		prepare func(t *testing.T, pool *pgxpool.Pool, ctx context.Context)
		wantErr bool
	}{
		{
			caseName: "success",
			data: struct {
				id     string
				userID string
				guest  *models.UpdateGuest
			}{
				id:     "00000000-0000-0000-0000-000000000021",
				userID: userID,
				guest: &models.UpdateGuest{
					FirstName: &newFirstName,
					Note:      &newNote,
					Phone:     &newPhone,
				},
			},
			prepare: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000021",
					userID,
					"Ivan",
					false,
					oldTime,
				))
			},
		},
		{
			caseName: "entry not found",
			data: struct {
				id     string
				userID string
				guest  *models.UpdateGuest
			}{
				id:     "00000000-0000-0000-0000-000000000099",
				userID: userID,
				guest:  &models.UpdateGuest{FirstName: &newFirstName},
			},
			wantErr: true,
		},
		{
			caseName: "entry belongs to another user",
			data: struct {
				id     string
				userID string
				guest  *models.UpdateGuest
			}{
				id:     "00000000-0000-0000-0000-000000000022",
				userID: otherUserID,
				guest:  &models.UpdateGuest{FirstName: &newFirstName},
			},
			prepare: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000022",
					userID,
					"Ivan",
					false,
					oldTime,
				))
			},
			wantErr: true,
		},
		{
			caseName: "entry deleted",
			data: struct {
				id     string
				userID string
				guest  *models.UpdateGuest
			}{
				id:     "00000000-0000-0000-0000-000000000023",
				userID: userID,
				guest:  &models.UpdateGuest{FirstName: &newFirstName},
			},
			prepare: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000023",
					userID,
					"Ivan",
					true,
					oldTime,
				))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			guestRepo, pool, ctx := setupGuestBookRepo(t)

			if tc.prepare != nil {
				tc.prepare(t, pool, ctx)
			}

			guest, err := guestRepo.UpdateEntry(
				ctx,
				tc.data.id,
				tc.data.userID,
				tc.data.guest,
			)

			if tc.wantErr {
				require.ErrorIs(t, err, domain.ErrNotFound)
				assert.Nil(t, guest)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, guest)

			assert.Equal(t, newFirstName, guest.FirstName)
			assert.Equal(t, newNote, *guest.Note)
			assert.Equal(t, newPhone, guest.Phone)

			assert.Equal(t, "Ivanovich", guest.MidName)
			assert.Equal(t, "Petrov", guest.LastName)

			assert.True(t, guest.UpdatedAt.After(oldTime))
			assert.False(t, guest.CreatedAt.IsZero())
		})
	}
}

func TestDeleteEntry(t *testing.T) {
	userID := "00000000-0000-0000-0000-000000000001"
	otherUserID := "00000000-0000-0000-0000-000000000002"

	testCases := []struct {
		caseName string
		data     struct {
			id     string
			userID string
		}
		prepare func(t *testing.T, pool *pgxpool.Pool, ctx context.Context)
		wantErr bool
	}{
		{
			caseName: "success",
			data: struct {
				id     string
				userID string
			}{
				id:     "00000000-0000-0000-0000-000000000031",
				userID: userID,
			},
			prepare: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000031",
					userID,
					"Ivan",
					false,
					time.Now(),
				))
			},
		},
		{
			caseName: "entry not found",
			data: struct {
				id     string
				userID string
			}{
				id:     "00000000-0000-0000-0000-000000000099",
				userID: userID,
			},
			wantErr: true,
		},
		{
			caseName: "entry belongs to another user",
			data: struct {
				id     string
				userID string
			}{
				id:     "00000000-0000-0000-0000-000000000032",
				userID: otherUserID,
			},
			prepare: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000032",
					userID,
					"Ivan",
					false,
					time.Now(),
				))
			},
			wantErr: true,
		},
		{
			caseName: "entry already deleted",
			data: struct {
				id     string
				userID string
			}{
				id:     "00000000-0000-0000-0000-000000000033",
				userID: userID,
			},
			prepare: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000033",
					userID,
					"Ivan",
					true,
					time.Now(),
				))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			guestRepo, pool, ctx := setupGuestBookRepo(t)

			if tc.prepare != nil {
				tc.prepare(t, pool, ctx)
			}

			err := guestRepo.DeleteEntry(ctx, tc.data.id, tc.data.userID)

			if tc.wantErr {
				require.ErrorIs(t, err, domain.ErrNotFound)
				return
			}

			require.NoError(t, err)

			var isDelete bool
			err = pool.QueryRow(ctx, `
				SELECT is_delete FROM guests WHERE id = $1
			`, tc.data.id).Scan(&isDelete)

			require.NoError(t, err)
			assert.True(t, isDelete)
		})
	}
}

func TestSearchEntry(t *testing.T) {
	userID := "00000000-0000-0000-0000-000000000001"
	otherUserID := "00000000-0000-0000-0000-000000000002"

	testCases := []struct {
		caseName string
		data     struct {
			userID    string
			firstName string
			midName   string
			lastName  string
		}
		prepare func(t *testing.T, pool *pgxpool.Pool, ctx context.Context)
		want    string
		wantErr bool
	}{
		{
			caseName: "success partial case insensitive search",
			data: struct {
				userID    string
				firstName string
				midName   string
				lastName  string
			}{
				userID:    userID,
				firstName: "lex",
				midName:   "SERGEE",
				lastName:  "ivan",
			},
			prepare: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
				guest := testGuest(
					"00000000-0000-0000-0000-000000000041",
					userID,
					"Alexander",
					false,
					time.Now(),
				)
				guest.MidName = "Sergeevich"
				guest.LastName = "Ivanov"

				insertTestGuest(t, pool, ctx, guest)
			},
			want: "Alexander",
		},
		{
			caseName: "entry not found",
			data: struct {
				userID    string
				firstName string
				midName   string
				lastName  string
			}{
				userID:    userID,
				firstName: "Unknown",
			},
			wantErr: true,
		},
		{
			caseName: "entry deleted",
			data: struct {
				userID    string
				firstName string
				midName   string
				lastName  string
			}{
				userID:    userID,
				firstName: "Ivan",
			},
			prepare: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000042",
					userID,
					"Ivan",
					true,
					time.Now(),
				))
			},
			wantErr: true,
		},
		{
			caseName: "entry belongs to another user",
			data: struct {
				userID    string
				firstName string
				midName   string
				lastName  string
			}{
				userID:    otherUserID,
				firstName: "Ivan",
			},
			prepare: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000043",
					userID,
					"Ivan",
					false,
					time.Now(),
				))
			},
			wantErr: true,
		},
		{
			caseName: "SQL injection treated as ordinary value",
			data: struct {
				userID    string
				firstName string
				midName   string
				lastName  string
			}{
				userID:    userID,
				firstName: "' OR 1=1 --",
			},
			prepare: func(t *testing.T, pool *pgxpool.Pool, ctx context.Context) {
				insertTestGuest(t, pool, ctx, testGuest(
					"00000000-0000-0000-0000-000000000044",
					userID,
					"Ivan",
					false,
					time.Now(),
				))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			guestRepo, pool, ctx := setupGuestBookRepo(t)

			if tc.prepare != nil {
				tc.prepare(t, pool, ctx)
			}

			guest, err := guestRepo.SearchEntry(
				ctx,
				tc.data.userID,
				tc.data.firstName,
				tc.data.midName,
				tc.data.lastName,
			)

			if tc.wantErr {
				require.ErrorIs(t, err, domain.ErrNotFound)
				assert.Nil(t, guest)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, guest)

			assert.Equal(t, tc.want, guest.FirstName)
			assert.Equal(t, tc.data.userID, guest.UserID)
		})
	}
}
