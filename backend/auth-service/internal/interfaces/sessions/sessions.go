package sessions

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"time"

	"github.com/BitCoinOffical/hotel-notepad/auth-service/internal/domain/models"
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

func (s *SessionRepo) SaveSession(ctx context.Context, sessionID string, session *models.Session) error {
	value, err := json.Marshal(&session)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	key := sessionPrefix + sessionID
	if err := s.rdb.Set(ctx, key, value, SessionTTL).Err(); err != nil {
		return fmt.Errorf("s.rdb.Set: %w", err)
	}

	return nil
}

func (s *SessionRepo) DeleteSession(ctx context.Context, sessionID string) error {
	key := sessionPrefix + sessionID
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("s.rdb.Del: %w", err)
	}
	return nil
}
