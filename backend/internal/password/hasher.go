package password

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	iterationTime = uint32(2)
	memory        = uint32(19 * 1024)
	threads       = uint8(4)
	keyLen        = uint32(32)
	saltLen       = 16
)

type Hasher struct{}

func New() *Hasher {
	return &Hasher{}
}

// GenerateFromPassword はargon2idアルゴリズムでパスワードをハッシュ化し、
// "<base64_salt>$<base64_hash>" 形式の文字列を返す。
func (h *Hasher) GenerateFromPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, iterationTime, memory, threads, keyLen)

	encoded := fmt.Sprintf(
		"%s$%s",
		base64.StdEncoding.EncodeToString(salt),
		base64.StdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}
