package user

import (
	"errors"
	"strings"
	"time"
	"unicode"
)

var (
	ErrUsernameRequired = errors.New("username is required")
	ErrUsernameInvalid  = errors.New("username cannot contain whitespace")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrInvalidPassword  = errors.New("invalid password")
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
	CompareHashAndPassword(hash, password string) error
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

func validatePassword(password string) error {
	const minPasswordLength = 8
	if len(password) < minPasswordLength {
		return ErrPasswordTooShort
	}
	return nil
}

func (s *Service) Create(username string, password string) error {
	if err := validateUsername(username); err != nil {
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
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

func (s *Service) UpdatePassword(id int64, oldPassword, newPassword string) error {
	usr, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if err := s.hashGenerator.CompareHashAndPassword(usr.PasswordHash, oldPassword); err != nil {
		return ErrInvalidPassword
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	hash, err := s.hashGenerator.GenerateFromPassword(newPassword)
	if err != nil {
		return err
	}
	return s.repo.Save(User{ID: id, PasswordHash: hash})
}
