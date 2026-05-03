// テストでモックを使用するためのコード

package user_test

import (
	"backend/internal/user"

	"github.com/stretchr/testify/mock"
)

type RepositoryMock struct {
	mock.Mock
}

func (r *RepositoryMock) Create(usr user.User) error {
	args := r.Called(usr)
	return args.Error(0)
}

func (r *RepositoryMock) Save(usr user.User) error {
	args := r.Called(usr)
	return args.Error(0)
}

func (r *RepositoryMock) FindByID(id int64) (user.User, error) {
	args := r.Called(id)
	return args.Get(0).(user.User), args.Error(1)
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
