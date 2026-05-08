package session

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/redis/go-redis/v9"
	"github.com/teilorbarcelos/backend-go/pkg/cache"
	"github.com/teilorbarcelos/backend-go/pkg/config"
)

func TestMain(m *testing.M) {
	// Setup test environment
	os.Setenv("ENVIRONMENT", "test")
	config.LoadConfig()
	cache.ConnectRedis()

	code := m.Run()
	os.Exit(code)
}

func TestSessionManager_InvalidateUserSessions(t *testing.T) {
	sm := NewSessionManager()
	ctx := context.Background()
	userId := "user123"
	roleId := "admin"

	// Setup: Add some session keys to Redis
	key1 := fmt.Sprintf("session:role:%s:user:%s:1", roleId, userId)
	key2 := fmt.Sprintf("session:role:%s:user:%s:2", roleId, userId)
	cache.RedisClient.Set(ctx, key1, "data", 0)
	cache.RedisClient.Set(ctx, key2, "data", 0)

	t.Run("Invalidate existing user sessions", func(t *testing.T) {
		err := sm.InvalidateUserSessions(userId, roleId)
		assert.NoError(t, err)

		// Verify keys are deleted
		val1 := cache.RedisClient.Exists(ctx, key1).Val()
		val2 := cache.RedisClient.Exists(ctx, key2).Val()
		assert.Equal(t, int64(0), val1)
		assert.Equal(t, int64(0), val2)
	})

	t.Run("Invalidate non-existing user sessions", func(t *testing.T) {
		err := sm.InvalidateUserSessions("nonexistent", "role")
		assert.NoError(t, err)
	})
}

func TestSessionManager_InvalidateRoleSessions(t *testing.T) {
	sm := NewSessionManager()
	ctx := context.Background()
	roleId := "manager"

	// Setup: Add some session keys to Redis
	key1 := fmt.Sprintf("session:role:%s:user:u1:1", roleId)
	key2 := fmt.Sprintf("session:role:%s:user:u2:2", roleId)
	cache.RedisClient.Set(ctx, key1, "data", 0)
	cache.RedisClient.Set(ctx, key2, "data", 0)

	t.Run("Invalidate existing role sessions", func(t *testing.T) {
		err := sm.InvalidateRoleSessions(roleId)
		assert.NoError(t, err)

		// Verify keys are deleted
		val1 := cache.RedisClient.Exists(ctx, key1).Val()
		val2 := cache.RedisClient.Exists(ctx, key2).Val()
		assert.Equal(t, int64(0), val1)
		assert.Equal(t, int64(0), val2)
	})

	t.Run("Invalidate non-existing role sessions", func(t *testing.T) {
		err := sm.InvalidateRoleSessions("nonexistent_role")
		assert.NoError(t, err)
	})
}

func TestSessionManager_DeleteByPattern_Error(t *testing.T) {
	sm := NewSessionManager()

	t.Run("Redis Scan error", func(t *testing.T) {
		// Backup and close client to trigger error on Scan
		originalClient := cache.RedisClient
		cache.RedisClient.Close()
		
		err := sm.InvalidateRoleSessions("any")
		assert.Error(t, err)
		
		// Restore client
		cache.RedisClient = originalClient
		cache.ConnectRedis()
	})

	t.Run("Redis Del error", func(t *testing.T) {
		// Setup: Add a key so Del is called
		userId := "error-user"
		key := fmt.Sprintf("session:role:admin:user:%s:1", userId)
		cache.RedisClient.Set(context.Background(), key, "data", 0)

		// Use hook to fail specifically on "del"
		hook := &delErrorHook{enabled: true}
		cache.RedisClient.AddHook(hook)
		
		// This should call Scan (success) then Del (fail)
		err := sm.InvalidateUserSessions(userId, "admin")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "forced del error")

		hook.enabled = false
	})
}

type delErrorHook struct {
	enabled bool
}

func (h *delErrorHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *delErrorHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if h.enabled && cmd.Name() == "del" {
			return fmt.Errorf("forced del error")
		}
		return next(ctx, cmd)
	}
}

func (h *delErrorHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

// Since I need to refer to redis types that might not be exported or are from another package,
// I should check how they are imported in the test file.
// Wait, I can just use the redis package directly in the test.
