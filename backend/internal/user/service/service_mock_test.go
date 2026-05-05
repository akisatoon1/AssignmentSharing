// テストでモックを使用するためのコード

package service_test

import (
	"backend/internal/user/service"
	"errors"

	"github.com/stretchr/testify/mock"
)

// モックから返されるエラーの定義。
// 依存対象のモジュールがエラーを返したときをテストするため。
var (
	hashErr    = errors.New("hash generator error")
	compareErr = errors.New("compare error")
	repoErr    = errors.New("repository error")
)

type RepositoryMock struct {
	mock.Mock
}

func (r *RepositoryMock) Create(usr service.User) error {
	args := r.Called(usr)
	return args.Error(0)
}

func (r *RepositoryMock) Save(usr service.User) error {
	args := r.Called(usr)
	return args.Error(0)
}

func (r *RepositoryMock) FindByID(id int64) (service.User, error) {
	args := r.Called(id)
	return args.Get(0).(service.User), args.Error(1)
}

type PasswordHasherMock struct {
	mock.Mock
}

func (h *PasswordHasherMock) GenerateFromPassword(password string) (string, error) {
	args := h.Called(password)
	return args.String(0), args.Error(1)
}

func (h *PasswordHasherMock) CompareHashAndPassword(hash, password string) error {
	args := h.Called(hash, password)
	return args.Error(0)
}
