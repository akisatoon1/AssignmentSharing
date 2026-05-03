package user_test

import (
	"backend/internal/user"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func runTest(t *testing.T, setupMock func(*RepositoryMock, *PasswordHasherMock), run func(*user.Service) error, expectedErr error) {
	t.Helper()

	repo := &RepositoryMock{}
	hash := &PasswordHasherMock{}
	repo.Test(t)
	hash.Test(t)

	if setupMock != nil {
		setupMock(repo, hash)
	}

	srv := user.NewService(repo, hash)
	err := run(srv)

	assert.ErrorIs(t, err, expectedErr)
	repo.AssertExpectations(t)
	hash.AssertExpectations(t)
}

func TestCreate(t *testing.T) {
	hashErr := errors.New("hash generator error")
	repoErr := errors.New("repository error")

	tests := []struct {
		name        string
		username    string
		password    string
		setupMock   func(*RepositoryMock, *PasswordHasherMock)
		expectedErr error
	}{
		{
			name:     "Success: Valid user creation",
			username: "testuser",
			password: "password",
			setupMock: func(repo *RepositoryMock, hash *PasswordHasherMock) {
				hash.On("GenerateFromPassword", "password").Return("thisisHaaash!", nil)
				repo.On("Create", user.User{Username: "testuser", PasswordHash: "thisisHaaash!"}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "Error: Empty username",
			username:    "",
			password:    "password123",
			expectedErr: user.ErrUsernameRequired,
		},
		{
			name:        "Error: Username contains space",
			username:    "test user",
			password:    "password123",
			expectedErr: user.ErrUsernameInvalid,
		},
		{
			name:        "Error: Password too short (7 chars)",
			username:    "testuser",
			password:    "1234567",
			expectedErr: user.ErrPasswordTooShort,
		},
		{
			name:        "Error: Password too short (empty)",
			username:    "testuser",
			password:    "",
			expectedErr: user.ErrPasswordTooShort,
		},
		{
			name:     "Error: Hash generator failure",
			username: "testuser",
			password: "password",
			setupMock: func(_ *RepositoryMock, hash *PasswordHasherMock) {
				hash.On("GenerateFromPassword", "password").Return("", hashErr)
			},
			expectedErr: hashErr,
		},
		{
			name:     "Error: Repository save failure",
			username: "testuser",
			password: "password",
			setupMock: func(repo *RepositoryMock, hash *PasswordHasherMock) {
				hash.On("GenerateFromPassword", "password").Return("thisisHaaash!", nil)
				repo.On("Create", mock.Anything).Return(repoErr)
			},
			expectedErr: repoErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTest(t, tc.setupMock, func(srv *user.Service) error {
				return srv.Create(tc.username, tc.password)
			}, tc.expectedErr)
		})
	}
}

func TestUpdateUsername(t *testing.T) {
	repoErr := errors.New("repository error")

	tests := []struct {
		name        string
		id          int64
		newUsername string
		setupMock   func(*RepositoryMock, *PasswordHasherMock)
		expectedErr error
	}{
		{
			name:        "Success: Valid update",
			id:          1,
			newUsername: "newname",
			setupMock: func(repo *RepositoryMock, _ *PasswordHasherMock) {
				repo.On("Save", user.User{ID: 1, Username: "newname"}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "Error: Empty username",
			id:          1,
			newUsername: "",
			expectedErr: user.ErrUsernameRequired,
		},
		{
			name:        "Error: Username contains space",
			id:          1,
			newUsername: "new name",
			expectedErr: user.ErrUsernameInvalid,
		},
		{
			name:        "Error: Repository failure",
			id:          1,
			newUsername: "newname",
			setupMock: func(repo *RepositoryMock, _ *PasswordHasherMock) {
				repo.On("Save", user.User{ID: 1, Username: "newname"}).Return(repoErr)
			},
			expectedErr: repoErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runTest(t, tc.setupMock, func(srv *user.Service) error {
				return srv.UpdateUsername(tc.id, tc.newUsername)
			}, tc.expectedErr)
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
		setupMock   func(*RepositoryMock, *PasswordHasherMock)
		expectedErr error
	}{
		{
			name:        "Success: Valid update",
			id:          1,
			oldPassword: "oldpassword",
			newPassword: "newpassword",
			setupMock: func(repo *RepositoryMock, hash *PasswordHasherMock) {
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
			setupMock: func(repo *RepositoryMock, _ *PasswordHasherMock) {
				repo.On("FindByID", int64(1)).Return(user.User{}, repoErr)
			},
			expectedErr: repoErr,
		},
		{
			name:        "Error: Invalid old password",
			id:          1,
			oldPassword: "wrongpassword",
			newPassword: "newpassword",
			setupMock: func(repo *RepositoryMock, hash *PasswordHasherMock) {
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
			setupMock: func(repo *RepositoryMock, hash *PasswordHasherMock) {
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
			setupMock: func(repo *RepositoryMock, hash *PasswordHasherMock) {
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
			setupMock: func(repo *RepositoryMock, hash *PasswordHasherMock) {
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
			setupMock: func(repo *RepositoryMock, hash *PasswordHasherMock) {
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
			setupMock: func(repo *RepositoryMock, hash *PasswordHasherMock) {
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
			runTest(t, tc.setupMock, func(srv *user.Service) error {
				return srv.UpdatePassword(tc.id, tc.oldPassword, tc.newPassword)
			}, tc.expectedErr)
		})
	}
}
