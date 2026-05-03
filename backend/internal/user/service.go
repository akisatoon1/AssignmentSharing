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
	ErrUserNotFound     = errors.New("user not found")
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

func validateUsername(username string) error {
	if username == "" {
		return ErrUsernameRequired
	}
	if strings.ContainsFunc(username, unicode.IsSpace) {
		return ErrUsernameInvalid
	}
	return nil
}

func (s *Service) Create(username string, password string) error {
	if err := validateUsername(username); err != nil {
		return err
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

func (s *Service) UpdateUsername(id int64, newUsername string) error {
	if err := validateUsername(newUsername); err != nil {
		return err
	}
	usr := User{
		ID:       id,
		Username: newUsername,
	}
	return s.repo.Save(usr)
}
