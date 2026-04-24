package user

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"golang.org/x/crypto/argon2"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	PasswordSalt string
	CreatedAt    time.Time
}

type Service struct {
	repo Repository
}

func (s *Service) Create(username string, password string) error {
	hash, salt, err := hashPassword(password)
	if err != nil {
		return err
	}
	usr := &User{
		Username:     username,
		PasswordHash: hash,
		PasswordSalt: salt,
	}
	return s.repo.Save(usr)
}

func hashPassword(password string) (passwdHash string, passwdSalt string, errr error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", "", err
	}

	// ハッシュアルゴリズムの強度を設定するためのパラメータ
	iterationTime := uint32(2)
	memory := uint32(19 * 1024) // KiB単位なので19Mibになる
	threads := uint8(4)

	keyLen := uint32(32)

	hash := argon2.IDKey([]byte(password), salt, iterationTime, memory, threads, keyLen)
	return base64.StdEncoding.EncodeToString(hash), base64.StdEncoding.EncodeToString(salt), nil
}
