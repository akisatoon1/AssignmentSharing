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
		setupMocks  func(repo *RepositoryMock, hash *PasswordHasherMock)
		expectedErr error
	}{
		{
			name:     "Success: Valid user creation",
			username: "testuser",
			password: "password",
			setupMocks: func(repo *RepositoryMock, hash *PasswordHasherMock) {
				hash.On("GenerateFromPassword", "password").Return("thisisHaaash!", nil)
				repo.On("Create", user.User{Username: "testuser", PasswordHash: "thisisHaaash!"}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "Error: Empty username",
			username:    "",
			password:    "password123",
			setupMocks:  nil,
			expectedErr: user.ErrUsernameRequired,
		},
		{
			name:        "Error: Username contains space",
			username:    "test user",
			password:    "password123",
			setupMocks:  nil,
			expectedErr: user.ErrUsernameInvalid,
		},
		{
			name:        "Error: Password too short (7 chars)",
			username:    "testuser",
			password:    "1234567",
			setupMocks:  nil,
			expectedErr: user.ErrPasswordTooShort,
		},
		{
			name:        "Error: Password too short (empty)",
			username:    "testuser",
			password:    "",
			setupMocks:  nil,
			expectedErr: user.ErrPasswordTooShort,
		},
		{
			name:     "Error: Hash generator failure",
			username: "testuser",
			password: "password",
			setupMocks: func(repo *RepositoryMock, hash *PasswordHasherMock) {
				hash.On("GenerateFromPassword", "password").Return("", hashErr)
			},
			expectedErr: hashErr,
		},
		{
			name:     "Error: Repository save failure",
			username: "testuser",
			password: "password",
			setupMocks: func(repo *RepositoryMock, hash *PasswordHasherMock) {
				hash.On("GenerateFromPassword", "password").Return("thisisHaaash!", nil)
				repo.On("Create", mock.Anything).Return(repoErr)
			},
			expectedErr: repoErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hashGenerator := &PasswordHasherMock{}
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
			setupMocks:  nil,
			expectedErr: user.ErrUsernameRequired,
		},
		{
			name:        "Error: Username contains space",
			id:          1,
			newUsername: "new name",
			setupMocks:  nil,
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
			hashGenerator := &PasswordHasherMock{}
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

func TestUpdatePassword(t *testing.T) {
	hashErr := errors.New("hash generator error")
	repoErr := errors.New("repository error")
	compareErr := errors.New("compare error")

	currentUser := user.User{ID: 1, PasswordHash: "currenthash"}

	tests := []struct {
		name        string
		id          int64
		oldPassword string
		newPassword string
		setupMocks  func(repo *RepositoryMock, hash *PasswordHasherMock)
		expectedErr error
	}{
		{
			name:        "Success: Valid update",
			id:          1,
			oldPassword: "oldpassword",
			newPassword: "newpassword",
			setupMocks: func(repo *RepositoryMock, hash *PasswordHasherMock) {
				repo.On("FindByID", int64(1)).Return(currentUser, nil)
				hash.On("CompareHashAndPassword", "currenthash", "oldpassword").Return(nil)
				hash.On("GenerateFromPassword", "newpassword").Return("newhash", nil)
				repo.On("Save", user.User{ID: 1, PasswordHash: "newhash"}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "Error: Repository FindByID failure",
			id:          1,
			oldPassword: "oldpassword",
			newPassword: "newpassword",
			setupMocks: func(repo *RepositoryMock, hash *PasswordHasherMock) {
				repo.On("FindByID", int64(1)).Return(user.User{}, repoErr)
			},
			expectedErr: repoErr,
		},
		{
			name:        "Error: Invalid old password",
			id:          1,
			oldPassword: "wrongpassword",
			newPassword: "newpassword",
			setupMocks: func(repo *RepositoryMock, hash *PasswordHasherMock) {
				repo.On("FindByID", int64(1)).Return(currentUser, nil)
				hash.On("CompareHashAndPassword", "currenthash", "wrongpassword").Return(compareErr)
			},
			expectedErr: user.ErrInvalidPassword,
		},
		{
			name:        "Error: New password too short (Empty)",
			id:          1,
			oldPassword: "oldpassword",
			newPassword: "",
			setupMocks: func(repo *RepositoryMock, hash *PasswordHasherMock) {
				repo.On("FindByID", int64(1)).Return(currentUser, nil)
				hash.On("CompareHashAndPassword", "currenthash", "oldpassword").Return(nil)
			},
			expectedErr: user.ErrPasswordTooShort,
		},
		{
			name:        "Error: New password too short (5 chars)",
			id:          1,
			oldPassword: "oldpassword",
			newPassword: "short",
			setupMocks: func(repo *RepositoryMock, hash *PasswordHasherMock) {
				repo.On("FindByID", int64(1)).Return(currentUser, nil)
				hash.On("CompareHashAndPassword", "currenthash", "oldpassword").Return(nil)
			},
			expectedErr: user.ErrPasswordTooShort,
		},
		{
			name:        "Error: New password too short (7 chars)",
			id:          1,
			oldPassword: "oldpassword",
			newPassword: "shorter",
			setupMocks: func(repo *RepositoryMock, hash *PasswordHasherMock) {
				repo.On("FindByID", int64(1)).Return(currentUser, nil)
				hash.On("CompareHashAndPassword", "currenthash", "oldpassword").Return(nil)
			},
			expectedErr: user.ErrPasswordTooShort,
		},
		{
			name:        "Error: Hash generator failure",
			id:          1,
			oldPassword: "oldpassword",
			newPassword: "newpassword",
			setupMocks: func(repo *RepositoryMock, hash *PasswordHasherMock) {
				repo.On("FindByID", int64(1)).Return(currentUser, nil)
				hash.On("CompareHashAndPassword", "currenthash", "oldpassword").Return(nil)
				hash.On("GenerateFromPassword", "newpassword").Return("", hashErr)
			},
			expectedErr: hashErr,
		},
		{
			name:        "Error: Repository save failure",
			id:          1,
			oldPassword: "oldpassword",
			newPassword: "newpassword",
			setupMocks: func(repo *RepositoryMock, hash *PasswordHasherMock) {
				repo.On("FindByID", int64(1)).Return(currentUser, nil)
				hash.On("CompareHashAndPassword", "currenthash", "oldpassword").Return(nil)
				hash.On("GenerateFromPassword", "newpassword").Return("newhash", nil)
				repo.On("Save", mock.Anything).Return(repoErr)
			},
			expectedErr: repoErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hashGenerator := &PasswordHasherMock{}
			repo := &RepositoryMock{}
			hashGenerator.Test(t)
			repo.Test(t)

			if tc.setupMocks != nil {
				tc.setupMocks(repo, hashGenerator)
			}

			srv := user.NewService(repo, hashGenerator)
			err := srv.UpdatePassword(tc.id, tc.oldPassword, tc.newPassword)

			assert.ErrorIs(t, err, tc.expectedErr)

			hashGenerator.AssertExpectations(t)
			repo.AssertExpectations(t)
		})
	}
}
