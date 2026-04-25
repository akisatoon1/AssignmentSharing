package password_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"backend/internal/password"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateFromPassword(t *testing.T) {
	h := password.New()

	// フォーマットを厳しくすると、内部の実装を変えたときにテストが壊れる可能性があるため。
	t.Run("Positive Case. Check Hash Format", func(t *testing.T) {
		hash, err := h.GenerateFromPassword("password123")
		require.NoError(t, err)
		assert.True(t, isHashFormat(hash), "Invalid hash format: %s", hash)
	})

	t.Run("Positive Case. Check Unique Hash", func(t *testing.T) {
		hash1, err := h.GenerateFromPassword("password123")
		require.NoError(t, err)
		hash2, err := h.GenerateFromPassword("password123")
		require.NoError(t, err)
		assert.NotEqual(t, hash1, hash2)
	})
}

// "<base64_salt>$<base64_hash>" 形式であることを確認するため
// saltの長さは16byte以上、hashの長さは16byte以上とする。
func isHashFormat(hash string) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 2 {
		return false
	}
	salt, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	hashBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	return len(salt) >= 16 && len(hashBytes) >= 16
}
