package middleware

import (
	"net/http"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/interfaces/sessions"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Middleware struct {
	logger  *zap.Logger
	session *sessions.SessionRepo
}

func NewMiddleware(logger *zap.Logger, rdb *redis.Client) *Middleware {
	session := sessions.NewSessionRepo(rdb)
	return &Middleware{logger: logger, session: session}
}

func (m *Middleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		session, err := m.session.GetSession(c.Request.Context(), sessionID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Set("session", session)
		c.Next()
	}
}
