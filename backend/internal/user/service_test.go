package user_test

import (
	"backend/internal/user"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	hashErr := errors.New("hash generator error")
	repoErr := errors.New("repository error")

	tests := []struct {
		name        string
		username    string
		password    string
		setupMocks  func(repo *RepositoryMock, hash *HashGeneratorMock)
		expectedErr error
	}{
		{
			name:     "Success: Valid user creation",
			username: "testuser",
			password: "password",
			setupMocks: func(repo *RepositoryMock, hash *HashGeneratorMock) {
				hash.On("GenerateFromPassword", "password").Return("thisisHaaash!", nil)
				repo.On("Create", user.User{Username: "testuser", PasswordHash: "thisisHaaash!"}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "Error: Empty username",
			username:    "",
			password:    "password123",
			setupMocks:  func(repo *RepositoryMock, hash *HashGeneratorMock) {},
			expectedErr: user.ErrUsernameRequired,
		},
		{
			name:        "Error: Username contains space",
			username:    "test user",
			password:    "password123",
			setupMocks:  func(repo *RepositoryMock, hash *HashGeneratorMock) {},
			expectedErr: user.ErrUsernameInvalid,
		},
		{
			name:        "Error: Password too short (7 chars)",
			username:    "testuser",
			password:    "1234567",
			setupMocks:  func(repo *RepositoryMock, hash *HashGeneratorMock) {},
			expectedErr: user.ErrPasswordTooShort,
		},
		{
			name:        "Error: Password too short (empty)",
			username:    "testuser",
			password:    "",
			setupMocks:  func(repo *RepositoryMock, hash *HashGeneratorMock) {},
			expectedErr: user.ErrPasswordTooShort,
		},
		{
			name:     "Error: Hash generator failure",
			username: "testuser",
			password: "password",
			setupMocks: func(repo *RepositoryMock, hash *HashGeneratorMock) {
				hash.On("GenerateFromPassword", "password").Return("", hashErr)
			},
			expectedErr: hashErr,
		},
		{
			name:     "Error: Repository save failure",
			username: "testuser",
			password: "password",
			setupMocks: func(repo *RepositoryMock, hash *HashGeneratorMock) {
				hash.On("GenerateFromPassword", "password").Return("thisisHaaash!", nil)
				repo.On("Create", mock.Anything).Return(repoErr)
			},
			expectedErr: repoErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hashGenerator := &HashGeneratorMock{}
			repo := &RepositoryMock{}
			hashGenerator.Test(t)
			repo.Test(t)

			if tc.setupMocks != nil {
				tc.setupMocks(repo, hashGenerator)
			}

			srv := user.NewService(repo, hashGenerator)
			err := srv.Create(tc.username, tc.password)

			assert.ErrorIs(t, err, tc.expectedErr)

			hashGenerator.AssertExpectations(t)
			repo.AssertExpectations(t)
		})
	}
}

func TestUpdateUsername(t *testing.T) {
	repoErr := errors.New("repository error")

	tests := []struct {
		name        string
		id          int64
		newUsername string
		setupMocks  func(repo *RepositoryMock)
		expectedErr error
	}{
		{
			name:        "Success: Valid update",
			id:          1,
			newUsername: "newname",
			setupMocks: func(repo *RepositoryMock) {
				repo.On("Save", user.User{ID: 1, Username: "newname"}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "Error: Empty username",
			id:          1,
			newUsername: "",
			setupMocks:  func(repo *RepositoryMock) {},
			expectedErr: user.ErrUsernameRequired,
		},
		{
			name:        "Error: Username contains space",
			id:          1,
			newUsername: "new name",
			setupMocks:  func(repo *RepositoryMock) {},
			expectedErr: user.ErrUsernameInvalid,
		},
		{
			name:        "Error: Repository failure",
			id:          1,
			newUsername: "newname",
			setupMocks: func(repo *RepositoryMock) {
				repo.On("Save", user.User{ID: 1, Username: "newname"}).Return(repoErr)
			},
			expectedErr: repoErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hashGenerator := &HashGeneratorMock{}
			repo := &RepositoryMock{}
			hashGenerator.Test(t)
			repo.Test(t)

			if tc.setupMocks != nil {
				tc.setupMocks(repo)
			}

			srv := user.NewService(repo, hashGenerator)
			err := srv.UpdateUsername(tc.id, tc.newUsername)

			assert.ErrorIs(t, err, tc.expectedErr)

			hashGenerator.AssertExpectations(t)
			repo.AssertExpectations(t)
		})
	}
}
