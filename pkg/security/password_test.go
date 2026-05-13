package security

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestPasswordHashing(t *testing.T) {
	password := "my_secret_password"

	t.Run("HashPassword success", func(t *testing.T) {
		hash, err := HashPassword(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, password, hash)
	})

	t.Run("CheckPasswordHash success", func(t *testing.T) {
		hash, _ := HashPassword(password)
		isValid := CheckPasswordHash(password, hash)
		assert.True(t, isValid)
	})

	t.Run("CheckPasswordHash failure", func(t *testing.T) {
		hash, _ := HashPassword(password)
		isValid := CheckPasswordHash("wrong_password", hash)
		assert.False(t, isValid)
	})

	t.Run("CheckPasswordHash empty hash", func(t *testing.T) {
		isValid := CheckPasswordHash(password, "")
		assert.False(t, isValid)
	})
}
