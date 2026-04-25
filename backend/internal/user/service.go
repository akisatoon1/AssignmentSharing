package user

import (
	"errors"
	"strings"
	"time"
	"unicode"
)

const minPasswordLength = 8

var (
	ErrUsernameRequired = errors.New("username is required")
	ErrUsernameInvalid  = errors.New("username cannot contain whitespace")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
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
	if username == "" {
		return ErrUsernameRequired
	}
	if strings.ContainsFunc(username, unicode.IsSpace) {
		return ErrUsernameInvalid
	}
	if len(password) < minPasswordLength {
		return ErrPasswordTooShort
	}
	hash, err := s.hashGenerator.GenerateFromPassword(password)
	if err != nil {
		return err
	}
	usr := User{
		Username:     username,
		PasswordHash: hash,
	}
	return s.repo.Create(usr)
}
