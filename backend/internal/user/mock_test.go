package user_test

import (
	"backend/internal/user"

	"github.com/stretchr/testify/mock"
)

type RepositoryMock struct {
	mock.Mock
}

func (r *RepositoryMock) Save(usr user.User) error {
	args := r.Called(usr)
	return args.Error(0)
}

type HashGeneratorMock struct {
	mock.Mock
}

func (h *HashGeneratorMock) GenerateFromPassword(password string) (string, error) {
	args := h.Called(password)
	return args.String(0), args.Error(1)
}
