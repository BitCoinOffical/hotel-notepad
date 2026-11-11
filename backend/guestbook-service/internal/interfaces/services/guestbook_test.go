package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/dto"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/models"
	repo_mocks "github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/interfaces/services/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateEntry(t *testing.T) {
	repoErr := errors.New("database error")
	var note = "Guest note"
	testCases := []struct {
		caseName string
		data     struct {
			userID string
			req    *dto.CreateGuestDTO
		}
		prepare func(repo *repo_mocks.MockGuestBookRepo)
		wantErr error
	}{
		{
			caseName: "success",
			data: struct {
				userID string
				req    *dto.CreateGuestDTO
			}{
				userID: "user-123",
				req: &dto.CreateGuestDTO{
					FirstName:     "Ivan",
					MidName:       "Ivanovich",
					LastName:      "Petrov",
					Note:          &note,
					Phone:         1234567890,
					ArrivalDate:   time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC),
					DepartureDate: time.Date(2026, 12, 10, 12, 0, 0, 0, time.UTC),
				},
			},
			prepare: func(repo *repo_mocks.MockGuestBookRepo) {
				repo.EXPECT().
					CreateEntry(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, guest *models.Guest) error {
						assert.Equal(t, "user-123", guest.UserID)
						assert.Equal(t, "Ivan", guest.FirstName)
						assert.Equal(t, "Ivanovich", guest.MidName)
						assert.Equal(t, "Petrov", guest.LastName)
						assert.Equal(t, "Guest note", *guest.Note)
						assert.Equal(t, 1234567890, guest.Phone)

						return nil
					})
			},
		},
		{
			caseName: "empty note",
			data: struct {
				userID string
				req    *dto.CreateGuestDTO
			}{
				userID: "user-123",
				req: &dto.CreateGuestDTO{
					FirstName:     "Ivan",
					MidName:       "Ivanovich",
					LastName:      "Petrov",
					Phone:         1234567890,
					ArrivalDate:   time.Date(2026, 12, 1, 12, 0, 0, 0, time.UTC),
					DepartureDate: time.Date(2026, 12, 10, 12, 0, 0, 0, time.UTC),
				},
			},
			prepare: func(repo *repo_mocks.MockGuestBookRepo) {
				repo.EXPECT().
					CreateEntry(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, guest *models.Guest) error {
						assert.Empty(t, guest.Note)
						return nil
					})
			},
		},
		{
			caseName: "repository error",
			data: struct {
				userID string
				req    *dto.CreateGuestDTO
			}{
				userID: "user-123",
				req: &dto.CreateGuestDTO{
					FirstName: "Ivan",
				},
			},
			prepare: func(repo *repo_mocks.MockGuestBookRepo) {
				repo.EXPECT().
					CreateEntry(gomock.Any(), gomock.Any()).
					Return(repoErr)
			},
			wantErr: repoErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repo_mocks.NewMockGuestBookRepo(ctrl)

			if tc.prepare != nil {
				tc.prepare(repo)
			}

			guestService := NewGuestBookService(repo)

			err := guestService.CreateEntry(
				context.Background(),
				tc.data.userID,
				tc.data.req,
			)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGetEntries(t *testing.T) {
	repoErr := errors.New("database error")

	expected := []models.Guest{
		{
			GuestID:   "guest-1",
			UserID:    "user-1",
			FirstName: "Ivan",
			LastName:  "Petrov",
		},
		{
			GuestID:   "guest-2",
			UserID:    "user-1",
			FirstName: "Alex",
			LastName:  "Ivanov",
		},
	}

	testCases := []struct {
		caseName string
		data     struct {
			userID string
		}
		prepare func(repo *repo_mocks.MockGuestBookRepo)
		want    []models.Guest
		wantErr error
	}{
		{
			caseName: "success",
			data: struct {
				userID string
			}{
				userID: "user-1",
			},
			prepare: func(repo *repo_mocks.MockGuestBookRepo) {
				repo.EXPECT().
					GetEntries(gomock.Any(), "user-1").
					Return(expected, nil)
			},
			want: expected,
		},
		{
			caseName: "empty list",
			data: struct {
				userID string
			}{
				userID: "user-1",
			},
			prepare: func(repo *repo_mocks.MockGuestBookRepo) {
				repo.EXPECT().
					GetEntries(gomock.Any(), "user-1").
					Return([]models.Guest{}, nil)
			},
			want: []models.Guest{},
		},
		{
			caseName: "repository error",
			data: struct {
				userID string
			}{
				userID: "user-1",
			},
			prepare: func(repo *repo_mocks.MockGuestBookRepo) {
				repo.EXPECT().
					GetEntries(gomock.Any(), "user-1").
					Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repo_mocks.NewMockGuestBookRepo(ctrl)

			if tc.prepare != nil {
				tc.prepare(repo)
			}

			guestService := NewGuestBookService(repo)

			guests, err := guestService.GetEntries(context.Background(), tc.data.userID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, guests)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, guests)
		})
	}
}

func TestGetEntryByID(t *testing.T) {
	repoErr := errors.New("guest not found")
	var note = "Guest note"
	expected := &models.Guest{
		GuestID:   "guest-123",
		UserID:    "user-123",
		FirstName: "Ivan",
		MidName:   "Ivanovich",
		LastName:  "Petrov",
		Note:      &note,
		Phone:     1234567890,
	}

	testCases := []struct {
		caseName string
		data     struct {
			id     string
			userID string
		}
		prepare func(repo *repo_mocks.MockGuestBookRepo)
		want    *models.Guest
		wantErr error
	}{
		{
			caseName: "success",
			data: struct {
				id     string
				userID string
			}{
				id:     "guest-123",
				userID: "user-123",
			},
			prepare: func(repo *repo_mocks.MockGuestBookRepo) {
				repo.EXPECT().
					GetEntryByID(gomock.Any(), "guest-123", "user-123").
					Return(expected, nil)
			},
			want: expected,
		},
		{
			caseName: "repository error",
			data: struct {
				id     string
				userID string
			}{
				id:     "guest-123",
				userID: "user-123",
			},
			prepare: func(repo *repo_mocks.MockGuestBookRepo) {
				repo.EXPECT().
					GetEntryByID(gomock.Any(), "guest-123", "user-123").
					Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repo_mocks.NewMockGuestBookRepo(ctrl)

			if tc.prepare != nil {
				tc.prepare(repo)
			}

			guestService := NewGuestBookService(repo)

			guest, err := guestService.GetEntryByID(
				context.Background(),
				tc.data.id,
				tc.data.userID,
			)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, guest)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, guest)
			assert.Equal(t, tc.want, guest)
		})
	}
}

func TestUpdateEntry(t *testing.T) {
	repoErr := errors.New("database error")

	firstName := "Alex"
	midName := "Sergeevich"
	lastName := "Ivanov"
	note := "Updated note"
	phone := 987654321
	arrivalDate := time.Date(2026, 12, 5, 12, 0, 0, 0, time.UTC)
	departureDate := time.Date(2026, 12, 15, 12, 0, 0, 0, time.UTC)

	expected := &models.Guest{
		GuestID:       "guest-123",
		UserID:        "user-123",
		FirstName:     firstName,
		MidName:       midName,
		LastName:      lastName,
		Note:          &note,
		Phone:         phone,
		ArrivalDate:   arrivalDate,
		DepartureDate: departureDate,
	}

	testCases := []struct {
		caseName string
		data     struct {
			id     string
			userID string
			req    *dto.UpdateGuestDTO
		}
		prepare func(repo *repo_mocks.MockGuestBookRepo)
		want    *models.Guest
		wantErr error
	}{
		{
			caseName: "success",
			data: struct {
				id     string
				userID string
				req    *dto.UpdateGuestDTO
			}{
				id:     "guest-123",
				userID: "user-123",
				req: &dto.UpdateGuestDTO{
					FirstName:     &firstName,
					MidName:       &midName,
					LastName:      &lastName,
					Note:          &note,
					Phone:         &phone,
					ArrivalDate:   &arrivalDate,
					DepartureDate: &departureDate,
				},
			},
			prepare: func(repo *repo_mocks.MockGuestBookRepo) {
				repo.EXPECT().
					UpdateEntry(gomock.Any(), "guest-123", "user-123", gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						id string,
						userID string,
						guest *models.UpdateGuest,
					) (*models.Guest, error) {
						assert.Equal(t, "Alex", *guest.FirstName)
						assert.Equal(t, "Sergeevich", *guest.MidName)
						assert.Equal(t, "Ivanov", *guest.LastName)
						assert.Equal(t, "Updated note", *guest.Note)
						assert.Equal(t, 987654321, *guest.Phone)
						assert.Equal(t, arrivalDate, *guest.ArrivalDate)
						assert.Equal(t, departureDate, *guest.DepartureDate)

						return expected, nil
					})
			},
			want: expected,
		},
		{
			caseName: "repository error",
			data: struct {
				id     string
				userID string
				req    *dto.UpdateGuestDTO
			}{
				id:     "guest-123",
				userID: "user-123",
				req: &dto.UpdateGuestDTO{
					FirstName: &firstName,
				},
			},
			prepare: func(repo *repo_mocks.MockGuestBookRepo) {
				repo.EXPECT().
					UpdateEntry(gomock.Any(), "guest-123", "user-123", gomock.Any()).
					Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repo_mocks.NewMockGuestBookRepo(ctrl)

			if tc.prepare != nil {
				tc.prepare(repo)
			}

			guestService := NewGuestBookService(repo)

			guest, err := guestService.UpdateEntry(
				context.Background(),
				tc.data.id,
				tc.data.userID,
				tc.data.req,
			)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, guest)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, guest)
			assert.Equal(t, tc.want, guest)
		})
	}
}

func TestDeleteEntry(t *testing.T) {
	repoErr := errors.New("database error")

	testCases := []struct {
		caseName string
		data     struct {
			id     string
			userID string
		}
		prepare func(repo *repo_mocks.MockGuestBookRepo)
		wantErr error
	}{
		{
			caseName: "success",
			data: struct {
				id     string
				userID string
			}{
				id:     "guest-123",
				userID: "user-123",
			},
			prepare: func(repo *repo_mocks.MockGuestBookRepo) {
				repo.EXPECT().
					DeleteEntry(gomock.Any(), "guest-123", "user-123").
					Return(nil)
			},
		},
		{
			caseName: "repository error",
			data: struct {
				id     string
				userID string
			}{
				id:     "guest-123",
				userID: "user-123",
			},
			prepare: func(repo *repo_mocks.MockGuestBookRepo) {
				repo.EXPECT().
					DeleteEntry(gomock.Any(), "guest-123", "user-123").
					Return(repoErr)
			},
			wantErr: repoErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repo_mocks.NewMockGuestBookRepo(ctrl)

			if tc.prepare != nil {
				tc.prepare(repo)
			}

			guestService := NewGuestBookService(repo)

			err := guestService.DeleteEntry(
				context.Background(),
				tc.data.id,
				tc.data.userID,
			)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestSearchEntry(t *testing.T) {
	repoErr := errors.New("database error")

	expected := &models.Guest{
		GuestID:   "guest-123",
		UserID:    "user-123",
		FirstName: "Alexander",
		MidName:   "Sergeevich",
		LastName:  "Ivanov",
	}

	testCases := []struct {
		caseName string
		data     struct {
			userID    string
			firstName string
			midName   string
			lastName  string
		}
		prepare func(repo *repo_mocks.MockGuestBookRepo)
		want    *models.Guest
		wantErr error
	}{
		{
			caseName: "success",
			data: struct {
				userID    string
				firstName string
				midName   string
				lastName  string
			}{
				userID:    "user-123",
				firstName: "Alex",
				midName:   "Sergeevich",
				lastName:  "Ivanov",
			},
			prepare: func(repo *repo_mocks.MockGuestBookRepo) {
				repo.EXPECT().
					SearchEntry(
						gomock.Any(),
						"user-123",
						"Alex",
						"Sergeevich",
						"Ivanov",
					).
					Return(expected, nil)
			},
			want: expected,
		},
		{
			caseName: "repository error",
			data: struct {
				userID    string
				firstName string
				midName   string
				lastName  string
			}{
				userID:    "user-123",
				firstName: "Alex",
				midName:   "",
				lastName:  "",
			},
			prepare: func(repo *repo_mocks.MockGuestBookRepo) {
				repo.EXPECT().
					SearchEntry(
						gomock.Any(),
						"user-123",
						"Alex",
						"",
						"",
					).
					Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := repo_mocks.NewMockGuestBookRepo(ctrl)

			if tc.prepare != nil {
				tc.prepare(repo)
			}

			guestService := NewGuestBookService(repo)

			guest, err := guestService.SearchEntry(
				context.Background(),
				tc.data.userID,
				tc.data.firstName,
				tc.data.midName,
				tc.data.lastName,
			)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, guest)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, guest)
			assert.Equal(t, tc.want, guest)
		})
	}
}
