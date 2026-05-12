package session

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"backend-go/pkg/cache"
)

type SessionStore interface {
	CreateSession(ctx context.Context, userId string, roleId string, tokenHash string, payload interface{}, expiration time.Duration) error
	CreateRefreshToken(ctx context.Context, userId string, roleId string, refreshTokenHash string, expiration time.Duration) error
	InvalidateUserSessions(userId string, roleId string) error
	InvalidateRoleSessions(roleId string) error
}

type SessionManager struct{}

func NewSessionManager() SessionStore {
	return &SessionManager{}
}

func (s *SessionManager) CreateSession(ctx context.Context, userId string, roleId string, tokenHash string, payload interface{}, expiration time.Duration) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("session:role:%s:user:%s:access:%s", roleId, userId, tokenHash)
	return cache.RedisClient.Set(ctx, key, payloadJSON, expiration).Err()
}

func (s *SessionManager) CreateRefreshToken(ctx context.Context, userId string, roleId string, refreshTokenHash string, expiration time.Duration) error {
	key := fmt.Sprintf("session:role:%s:user:%s:refresh:%s", roleId, userId, refreshTokenHash)
	return cache.RedisClient.Set(ctx, key, "1", expiration).Err()
}

func (s *SessionManager) InvalidateUserSessions(userId string, roleId string) error {
	ctx := context.Background()
	pattern := fmt.Sprintf("session:role:*:user:%s:*", userId)

	log.Printf("[SessionManager] Invalidating sessions for user %s with pattern %s", userId, pattern)

	return s.deleteByPattern(ctx, pattern)
}

func (s *SessionManager) InvalidateRoleSessions(roleId string) error {
	ctx := context.Background()
	pattern := fmt.Sprintf("session:role:%s:*", roleId)

	log.Printf("[SessionManager] Invalidating all sessions for role %s with pattern %s", roleId, pattern)

	return s.deleteByPattern(ctx, pattern)
}

func (s *SessionManager) deleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := cache.RedisClient.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			log.Printf("[SessionManager] Found keys to delete: %v", keys)
			if err := cache.RedisClient.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			log.Printf("[SessionManager] Deleted %d session keys", len(keys))
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
