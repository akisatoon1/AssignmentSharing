package service_test

import (
	"backend/internal/group/service"
	"errors"

	"github.com/stretchr/testify/mock"
)

var (
	repoErr  = errors.New("repository error")
	tokenErr = errors.New("token store error")
)

type RepositoryMock struct{ mock.Mock }

func (m *RepositoryMock) CreateGroup(grp service.Group, userID int64) (int64, error) {
	args := m.Called(grp, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *RepositoryMock) IsMember(groupID, userID int64) (bool, error) {
	args := m.Called(groupID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *RepositoryMock) AddMember(groupID, userID int64) error {
	return m.Called(groupID, userID).Error(0)
}

func (m *RepositoryMock) RemoveMember(groupID, userID int64) error {
	return m.Called(groupID, userID).Error(0)
}

type TokenStoreMock struct{ mock.Mock }

func (m *TokenStoreMock) CreateToken(groupID int64) (string, error) {
	args := m.Called(groupID)
	return args.String(0), args.Error(1)
}

func (m *TokenStoreMock) GetGroupIDByToken(token string) *int64 {
	args := m.Called(token)
	result := args.Get(0)
	if result == nil {
		return nil
	}
	v := result.(int64)
	return &v
}
