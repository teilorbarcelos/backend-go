package session

import (
	"context"
	"fmt"
	"log"

	"backend-go/pkg/cache"
)

type SessionManager struct{}

func NewSessionManager() *SessionManager {
	return &SessionManager{}
}

func (s *SessionManager) InvalidateUserSessions(userId string, roleId string) error {
	ctx := context.Background()
	// Sempre usamos wildcard para o papel ao invalidar um usuário específico,
	// para garantir que todas as suas sessões sejam derrubadas independentemente da role.
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
