package middleware

import (
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain"
	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/models"
	"github.com/gin-gonic/gin"
)

func GetUserSession(c *gin.Context) (*models.Session, error) {
	value, exists := c.Get("session")
	if !exists {
		return nil, domain.ErrNotFound
	}

	v, ok := value.(*models.Session)
	if !ok {
		return nil, domain.ErrIncorrectType
	}

	return v, nil
}
