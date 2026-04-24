package user

import (
	"time"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type Service struct {
	repo          Repository
	hashGenerator HashGenerator
}

func NewService(repo Repository, hashGenerator HashGenerator) *Service {
	return &Service{repo: repo, hashGenerator: hashGenerator}
}

type HashGenerator interface {
	GenerateFromPassword(password string) (hash string, err error)
}

func (s *Service) Create(username string, password string) error {
	hash, err := s.hashGenerator.GenerateFromPassword(password)
	if err != nil {
		return err
	}
	usr := User{
		Username:     username,
		PasswordHash: hash,
	}
	return s.repo.Save(usr)
}

/*
func hashAndSalt(password string) (passwdHash string, passwdSalt string, errr error) {
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
*/
