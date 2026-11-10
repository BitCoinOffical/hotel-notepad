package handlers

import (
	"errors"
	"net/http"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/api/response"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/dto"
	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/interfaces/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	srvc   *services.AuthService
	logger *zap.Logger
}

func NewAuthHandler(srvc *services.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{srvc: srvc, logger: logger}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request", h.logger)
		return
	}

	sessionID, err := h.srvc.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrNotMatch) {
			response.BadRequest(c, err, "bad request", h.logger)
			return
		}
		if errors.Is(err, domain.ErrUserConflict) {
			response.Conflict(c, err, "user already exists", h.logger)
			return
		}
		response.InternalServerError(c, err, "failed register user", h.logger)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   2678400,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	c.Status(http.StatusCreated)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginDTO
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BadRequest(c, err, "invalid request", h.logger)
		return
	}

	sessionID, err := h.srvc.Login(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, err, "failed login user", h.logger)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   2678400,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	c.Status(http.StatusCreated)
}

func (h *AuthHandler) Logout(c *gin.Context) {

}
