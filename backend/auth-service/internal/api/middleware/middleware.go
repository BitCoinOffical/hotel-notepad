package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Middleware struct {
	logger *zap.Logger
	srvc   AuthService
}

func NewMiddleware(logger *zap.Logger, srvc AuthService) *Middleware {
	return &Middleware{logger: logger, srvc: srvc}
}

func (m *Middleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		session, err := m.srvc.GetSession(c.Request.Context(), sessionID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Set("session", session)
		c.Next()
	}
}
