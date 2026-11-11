package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/notebook-service/internal/domain/models"
	"github.com/redis/go-redis/v9"
)

const (
	sessionPrefix = "session:"
	SessionTTL    = 31 * 24 * time.Hour
)

type SessionRepo struct {
	rdb *redis.Client
}

func NewSessionRepo(rdb *redis.Client) *SessionRepo {
	return &SessionRepo{rdb: rdb}
}

func (s *SessionRepo) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	key := sessionPrefix + sessionID

	value, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("s.rdb.Get: %w", err)
	}

	var session models.Session
	if err := json.Unmarshal([]byte(value), &session); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return &session, nil
}
