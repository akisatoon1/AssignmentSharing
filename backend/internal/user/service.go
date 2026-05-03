// サービス層
// ユーザ関連の操作を実行して結果を返す役割を持ち、サービス層に渡る前に認証は済んでいる前提。
// 認証もサービス層で行うと、サービス層が依存する対象が増え(session管理のモジュール)、テストが複雑になるから。

package user

import (
	"errors"
	"strings"
	"time"
	"unicode"
)

// サービス層で返されるエラーを定義したいが、今は完全に定義できておらず。
// 現状、RepositoryやPasswordHasherのエラーはそのまま返すが、後々はラップして返すかもしれない
var (
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

// Repositoryはデータを保存するため
// PasswordHasherはパスワードのハッシュ化とパスワードの比較のため
type Service struct {
	repo     Repository
	pdHasher PasswordHasher
}

func NewService(repo Repository, pdHasher PasswordHasher) *Service {
	return &Service{repo: repo, pdHasher: pdHasher}
}

// パスワードのハッシュ化とパスワードの比較のため
type PasswordHasher interface {
	GenerateFromPassword(password string) (hash string, err error)
	CompareHashAndPassword(hash, password string) error
}

// usernameの条件は
// - 空であってはならない
// - 空白を含んではいけない
func validateUsername(username string) error {
	if username == "" {
		return ErrUsernameInvalid
	}
	if strings.ContainsFunc(username, unicode.IsSpace) {
		return ErrUsernameInvalid
	}
	return nil
}

// パスワードの条件は
// - 8文字以上でなければならない
func validatePassword(password string) error {
	const minPasswordLength = 8
	if len(password) < minPasswordLength {
		return ErrPasswordTooShort
	}
	return nil
}

// ユーザを作成するため
func (s *Service) Create(username string, password string) error {
	if err := validateUsername(username); err != nil {
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
	}
	hash, err := s.pdHasher.GenerateFromPassword(password)
	if err != nil {
		return err
	}
	usr := User{
		Username:     username,
		PasswordHash: hash,
	}
	return s.repo.Create(usr)
}

// ユーザネームを変更するため
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

// パスワードを変更するため
// UpdateUsernameと分けたのは、パスワードの変更時に変更前のパスワードを正しく知っているかどうかを確認するため
func (s *Service) UpdatePassword(id int64, oldPassword, newPassword string) error {
	usr, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if err := s.pdHasher.CompareHashAndPassword(usr.PasswordHash, oldPassword); err != nil {
		return ErrInvalidPassword
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	hash, err := s.pdHasher.GenerateFromPassword(newPassword)
	if err != nil {
		return err
	}
	return s.repo.Save(User{ID: id, PasswordHash: hash})
}
