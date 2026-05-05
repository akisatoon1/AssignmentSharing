package presentation_test

import (
	"backend/internal/auth/session"
	"errors"

	"github.com/stretchr/testify/mock"
)

var serviceErr = errors.New("service error")

type ServiceMock struct{ mock.Mock }

func (m *ServiceMock) Create(name string, userID int64) (string, error) {
	args := m.Called(name, userID)
	return args.String(0), args.Error(1)
}

func (m *ServiceMock) IssueToken(groupID int64, requesterID int64) (string, error) {
	args := m.Called(groupID, requesterID)
	return args.String(0), args.Error(1)
}

func (m *ServiceMock) AddUser(token string, userID int64) error {
	return m.Called(token, userID).Error(0)
}

func (m *ServiceMock) DeleteUser(groupID int64, requesterID int64, targetUserID int64) error {
	return m.Called(groupID, requesterID, targetUserID).Error(0)
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
