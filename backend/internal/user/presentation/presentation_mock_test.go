// テストでモックを使用するためのコード（プレゼンテーション層用）

package presentation_test

import (
	"backend/internal/auth/session"
	"errors"

	"github.com/stretchr/testify/mock"
)

// ServiceMockからエラーが返されるときをテストするため
var serviceErr = errors.New("service error")

type ServiceMock struct{ mock.Mock }

func (m *ServiceMock) Create(username, password string) error {
	return m.Called(username, password).Error(0)
}

func (m *ServiceMock) UpdateUsername(id int64, newUsername string) error {
	return m.Called(id, newUsername).Error(0)
}

func (m *ServiceMock) UpdatePassword(id int64, oldPassword, newPassword string) error {
	return m.Called(id, oldPassword, newPassword).Error(0)
}

type SessionStoreMock struct{ mock.Mock }

func (m *SessionStoreMock) Get(id string) *session.Session {
	args := m.Called(id)
	result := args.Get(0)
	if result == nil {
		return nil
	}
	return result.(*session.Session)
}
