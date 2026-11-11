package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/api/middleware"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/api/response"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/dto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GuestBookHandler struct {
	srvc   GuestBookService
	logger *zap.Logger
}

func NewGuestBookHandler(logger *zap.Logger, srvc GuestBookService) *GuestBookHandler {
	return &GuestBookHandler{logger: logger, srvc: srvc}
}

func (h *GuestBookHandler) CreateEntry(c *gin.Context) {
	session, err := middleware.GetUserSession(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user session", h.logger)
		return
	}
	var req dto.CreateGuestDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request", h.logger)
		return
	}
	if err := h.srvc.CreateEntry(c.Request.Context(), session.UserID, &req); err != nil {
		response.InternalServerError(c, err, "failed create entry", h.logger)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *GuestBookHandler) GetEntries(c *gin.Context) {
	session, err := middleware.GetUserSession(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user session", h.logger)
		return
	}

	guests, err := h.srvc.GetEntries(c.Request.Context(), session.UserID)
	if err != nil {
		response.InternalServerError(c, err, "failed get entries", h.logger)
		return
	}

	c.JSON(http.StatusOK, guests)
}
func (h *GuestBookHandler) GetEntryByID(c *gin.Context) {
	guestID := c.Param("guest_id")
	session, err := middleware.GetUserSession(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user session", h.logger)
		return
	}

	guest, err := h.srvc.GetEntryByID(c.Request.Context(), guestID, session.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "entry not found", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed get entries", h.logger)
		return
	}

	c.JSON(http.StatusOK, guest)
}
func (h *GuestBookHandler) UpdateEntry(c *gin.Context) {
	var req dto.UpdateGuestDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request", h.logger)
		return
	}

	guestID := c.Param("guest_id")
	session, err := middleware.GetUserSession(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user session", h.logger)
		return
	}

	newguest, err := h.srvc.UpdateEntry(c.Request.Context(), guestID, session.UserID, &req)
	if err != nil {
		response.InternalServerError(c, err, "failed update entry", h.logger)
		return
	}

	c.JSON(http.StatusOK, newguest)
}
func (h *GuestBookHandler) DeleteEntry(c *gin.Context) {
	guestID := c.Param("guest_id")
	session, err := middleware.GetUserSession(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user session", h.logger)
		return
	}
	if err := h.srvc.DeleteEntry(c.Request.Context(), guestID, session.UserID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "entry not found", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed delete entry", h.logger)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *GuestBookHandler) SearchEntry(c *gin.Context) {
	firstName := c.Query("first_name")
	midName := c.Query("mid_name")
	lastName := c.Query("last_name")
	session, err := middleware.GetUserSession(c)
	if err != nil {
		response.Unauthorized(c, err, "failed get user session", h.logger)
		return
	}

	guest, err := h.srvc.SearchEntry(c.Request.Context(), session.UserID, firstName, midName, lastName)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, err, "entry not found", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed search entry", h.logger)
		return
	}

	c.JSON(http.StatusOK, guest)
}
