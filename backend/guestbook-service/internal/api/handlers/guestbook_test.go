package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	srvc_mocks "github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/api/handlers/mock"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/dto"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
)

func TestCreateEntry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	note := "Guest note"

	body, err := json.Marshal(&dto.CreateGuestDTO{
		FirstName: "Ivan",
		MidName:   "Ivanovich",
		LastName:  "Petrov",
		Note:      &note,
		Phone:     12345678901,
		ArrivalDate: time.Date(
			2026, 12, 1, 12, 0, 0, 0, time.UTC,
		),
		DepartureDate: time.Date(
			2026, 12, 10, 12, 0, 0, 0, time.UTC,
		),
	})
	require.NoError(t, err)
	serviceErr := errors.New("service error")

	testCases := []struct {
		caseName    string
		data        struct{ body string }
		withSession bool
		prepare     func(srvc *srvc_mocks.MockGuestBookService)
		status      int
	}{
		{
			caseName: "success",
			data: struct{ body string }{
				body: string(body),
			},
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					CreateEntry(gomock.Any(), "user-123", gomock.Any()).
					DoAndReturn(func(_ context.Context, userID string, req *dto.CreateGuestDTO) error {
						assert.Equal(t, "user-123", userID)
						assert.Equal(t, "Ivan", req.FirstName)
						assert.Equal(t, "Ivanovich", req.MidName)
						assert.Equal(t, "Petrov", req.LastName)
						require.NotNil(t, req.Note)
						assert.Equal(t, "Guest note", *req.Note)

						return nil
					})
			},
			status: http.StatusCreated,
		},
		{
			caseName: "session not found",
			data: struct{ body string }{
				body: string(body),
			},
			status: http.StatusUnauthorized,
		},
		{
			caseName: "invalid request",
			data: struct{ body string }{
				body: "{invalid json",
			},
			withSession: true,
			status:      http.StatusBadRequest,
		},
		{
			caseName: "service error",
			data: struct{ body string }{
				body: string(body),
			},
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					CreateEntry(gomock.Any(), "user-123", gomock.Any()).
					Return(serviceErr)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockGuestBookService(ctrl)

			if tc.prepare != nil {
				tc.prepare(srvc)
			}

			handler := NewGuestBookHandler(zaptest.NewLogger(t), srvc)

			router := gin.New()

			if tc.withSession {
				router.Use(func(c *gin.Context) {
					c.Set("session", &models.Session{
						UserID:   "user-123",
						Username: "test_user",
					})
					c.Next()
				})
			}

			router.POST("/entries", handler.CreateEntry)

			req := httptest.NewRequest(
				http.MethodPost,
				"/entries",
				strings.NewReader(tc.data.body),
			)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)
		})
	}
}

func TestGetEntries(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceErr := errors.New("service error")

	expected := []models.Guest{
		{
			GuestID:   "guest-1",
			UserID:    "user-123",
			FirstName: "Ivan",
		},
		{
			GuestID:   "guest-2",
			UserID:    "user-123",
			FirstName: "Alex",
		},
	}

	testCases := []struct {
		caseName    string
		withSession bool
		prepare     func(srvc *srvc_mocks.MockGuestBookService)
		status      int
		want        []models.Guest
	}{
		{
			caseName:    "success",
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					GetEntries(gomock.Any(), "user-123").
					Return(expected, nil)
			},
			status: http.StatusOK,
			want:   expected,
		},
		{
			caseName:    "empty list",
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					GetEntries(gomock.Any(), "user-123").
					Return([]models.Guest{}, nil)
			},
			status: http.StatusOK,
			want:   []models.Guest{},
		},
		{
			caseName: "session not found",
			status:   http.StatusUnauthorized,
		},
		{
			caseName:    "service error",
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					GetEntries(gomock.Any(), "user-123").
					Return(nil, serviceErr)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockGuestBookService(ctrl)

			if tc.prepare != nil {
				tc.prepare(srvc)
			}

			handler := NewGuestBookHandler(zaptest.NewLogger(t), srvc)

			router := gin.New()

			if tc.withSession {
				router.Use(func(c *gin.Context) {
					c.Set("session", &models.Session{
						UserID:   "user-123",
						Username: "test_user",
					})
					c.Next()
				})
			}

			router.GET("/entries", handler.GetEntries)

			req := httptest.NewRequest(http.MethodGet, "/entries", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)

			if tc.status == http.StatusOK {
				var guests []models.Guest

				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &guests))
				assert.Equal(t, tc.want, guests)
			}
		})
	}
}

func TestGetEntryByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceErr := errors.New("service error")

	expected := &models.Guest{
		GuestID:   "guest-123",
		UserID:    "user-123",
		FirstName: "Ivan",
	}

	testCases := []struct {
		caseName    string
		data        struct{ guestID string }
		withSession bool
		prepare     func(srvc *srvc_mocks.MockGuestBookService)
		status      int
		want        *models.Guest
	}{
		{
			caseName: "success",
			data: struct{ guestID string }{
				guestID: "guest-123",
			},
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					GetEntryByID(gomock.Any(), "guest-123", "user-123").
					Return(expected, nil)
			},
			status: http.StatusOK,
			want:   expected,
		},
		{
			caseName: "session not found",
			data: struct{ guestID string }{
				guestID: "guest-123",
			},
			status: http.StatusUnauthorized,
		},
		{
			caseName: "entry not found",
			data: struct{ guestID string }{
				guestID: "guest-123",
			},
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					GetEntryByID(gomock.Any(), "guest-123", "user-123").
					Return(nil, domain.ErrNotFound)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "service error",
			data: struct{ guestID string }{
				guestID: "guest-123",
			},
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					GetEntryByID(gomock.Any(), "guest-123", "user-123").
					Return(nil, serviceErr)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockGuestBookService(ctrl)

			if tc.prepare != nil {
				tc.prepare(srvc)
			}

			handler := NewGuestBookHandler(zaptest.NewLogger(t), srvc)

			router := gin.New()

			if tc.withSession {
				router.Use(func(c *gin.Context) {
					c.Set("session", &models.Session{
						UserID:   "user-123",
						Username: "test_user",
					})
					c.Next()
				})
			}

			router.GET("/entries/:guest_id", handler.GetEntryByID)

			req := httptest.NewRequest(
				http.MethodGet,
				"/entries/"+tc.data.guestID,
				nil,
			)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)

			if tc.want != nil {
				var guest models.Guest
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &guest))
				assert.Equal(t, tc.want.GuestID, guest.GuestID)
				assert.Equal(t, tc.want.UserID, guest.UserID)
				assert.Equal(t, tc.want.FirstName, guest.FirstName)
			}
		})
	}
}

func TestUpdateEntry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body, err := json.Marshal(&dto.UpdateGuestDTO{
		FirstName: stringPointer("Alex"),
	})
	require.NoError(t, err)

	serviceErr := errors.New("service error")

	expected := &models.Guest{
		GuestID:   "guest-123",
		UserID:    "user-123",
		FirstName: "Alex",
	}

	testCases := []struct {
		caseName    string
		data        struct{ body string }
		withSession bool
		prepare     func(srvc *srvc_mocks.MockGuestBookService)
		status      int
		want        *models.Guest
	}{
		{
			caseName: "success",
			data: struct{ body string }{
				body: string(body),
			},
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					UpdateEntry(gomock.Any(), "guest-123", "user-123", gomock.Any()).
					DoAndReturn(func(
						_ context.Context,
						guestID string,
						userID string,
						req *dto.UpdateGuestDTO,
					) (*models.Guest, error) {
						require.NotNil(t, req.FirstName)
						assert.Equal(t, "Alex", *req.FirstName)

						return expected, nil
					})
			},
			status: http.StatusOK,
			want:   expected,
		},
		{
			caseName: "invalid request",
			data: struct{ body string }{
				body: "{invalid json",
			},
			withSession: true,
			status:      http.StatusBadRequest,
		},
		{
			caseName: "session not found",
			data: struct{ body string }{
				body: string(body),
			},
			status: http.StatusUnauthorized,
		},
		{
			caseName: "service error",
			data: struct{ body string }{
				body: string(body),
			},
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					UpdateEntry(gomock.Any(), "guest-123", "user-123", gomock.Any()).
					Return(nil, serviceErr)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockGuestBookService(ctrl)

			if tc.prepare != nil {
				tc.prepare(srvc)
			}

			handler := NewGuestBookHandler(zaptest.NewLogger(t), srvc)

			router := gin.New()

			if tc.withSession {
				router.Use(func(c *gin.Context) {
					c.Set("session", &models.Session{
						UserID:   "user-123",
						Username: "test_user",
					})
					c.Next()
				})
			}

			router.PATCH("/entries/:guest_id", handler.UpdateEntry)

			req := httptest.NewRequest(
				http.MethodPatch,
				"/entries/guest-123",
				strings.NewReader(tc.data.body),
			)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)

			if tc.want != nil {
				var guest models.Guest
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &guest))
				assert.Equal(t, tc.want.GuestID, guest.GuestID)
				assert.Equal(t, tc.want.UserID, guest.UserID)
				assert.Equal(t, tc.want.FirstName, guest.FirstName)
			}
		})
	}
}

func TestDeleteEntry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceErr := errors.New("service error")

	testCases := []struct {
		caseName    string
		withSession bool
		prepare     func(srvc *srvc_mocks.MockGuestBookService)
		status      int
	}{
		{
			caseName:    "success",
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					DeleteEntry(gomock.Any(), "guest-123", "user-123").
					Return(nil)
			},
			status: http.StatusNoContent,
		},
		{
			caseName: "session not found",
			status:   http.StatusUnauthorized,
		},
		{
			caseName:    "entry not found",
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					DeleteEntry(gomock.Any(), "guest-123", "user-123").
					Return(domain.ErrNotFound)
			},
			status: http.StatusNotFound,
		},
		{
			caseName:    "service error",
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					DeleteEntry(gomock.Any(), "guest-123", "user-123").
					Return(serviceErr)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockGuestBookService(ctrl)

			if tc.prepare != nil {
				tc.prepare(srvc)
			}

			handler := NewGuestBookHandler(zaptest.NewLogger(t), srvc)

			router := gin.New()

			if tc.withSession {
				router.Use(func(c *gin.Context) {
					c.Set("session", &models.Session{
						UserID:   "user-123",
						Username: "test_user",
					})
					c.Next()
				})
			}

			router.DELETE("/entries/:guest_id", handler.DeleteEntry)

			req := httptest.NewRequest(http.MethodDelete, "/entries/guest-123", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)
		})
	}
}

func TestSearchEntry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceErr := errors.New("service error")

	expected := &models.Guest{
		GuestID:   "guest-123",
		UserID:    "user-123",
		FirstName: "Alexander",
	}

	testCases := []struct {
		caseName    string
		data        struct{ query string }
		withSession bool
		prepare     func(srvc *srvc_mocks.MockGuestBookService)
		status      int
		want        *models.Guest
	}{
		{
			caseName: "success",
			data: struct{ query string }{
				query: "?first_name=Alex&mid_name=Sergeevich&last_name=Ivanov",
			},
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					SearchEntry(gomock.Any(), "user-123", "Alex", "Sergeevich", "Ivanov").
					Return(expected, nil)
			},
			status: http.StatusOK,
			want:   expected,
		},
		{
			caseName: "empty search",
			data: struct{ query string }{
				query: "",
			},
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					SearchEntry(gomock.Any(), "user-123", "", "", "").
					Return(expected, nil)
			},
			status: http.StatusOK,
			want:   expected,
		},
		{
			caseName: "session not found",
			status:   http.StatusUnauthorized,
		},
		{
			caseName: "entry not found",
			data: struct{ query string }{
				query: "?first_name=Unknown",
			},
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					SearchEntry(gomock.Any(), "user-123", "Unknown", "", "").
					Return(nil, domain.ErrNotFound)
			},
			status: http.StatusNotFound,
		},
		{
			caseName: "service error",
			data: struct{ query string }{
				query: "?first_name=Ivan",
			},
			withSession: true,
			prepare: func(srvc *srvc_mocks.MockGuestBookService) {
				srvc.EXPECT().
					SearchEntry(gomock.Any(), "user-123", "Ivan", "", "").
					Return(nil, serviceErr)
			},
			status: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			srvc := srvc_mocks.NewMockGuestBookService(ctrl)

			if tc.prepare != nil {
				tc.prepare(srvc)
			}

			handler := NewGuestBookHandler(zaptest.NewLogger(t), srvc)

			router := gin.New()

			if tc.withSession {
				router.Use(func(c *gin.Context) {
					c.Set("session", &models.Session{
						UserID:   "user-123",
						Username: "test_user",
					})
					c.Next()
				})
			}

			router.GET("/entries/search", handler.SearchEntry)

			req := httptest.NewRequest(
				http.MethodGet,
				"/entries/search"+tc.data.query,
				nil,
			)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code)

			if tc.want != nil {
				var guest models.Guest
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &guest))
				assert.Equal(t, tc.want.GuestID, guest.GuestID)
				assert.Equal(t, tc.want.UserID, guest.UserID)
				assert.Equal(t, tc.want.FirstName, guest.FirstName)
			}
		})
	}
}

func stringPointer(value string) *string {
	return &value
}
