// 入力やモックの振る舞いから、関数の出力やモックへの呼び出しをテストしている。

package user_test

import (
	"backend/internal/user"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 各テストケースを実行するときの共通のボイラープレートをまとめるための関数
func runTest(
	t *testing.T,
	setupMock func(*RepositoryMock, *PasswordHasherMock),
	run func(*user.Service) error,
	expectedErr error) {
	t.Helper()

	repo := &RepositoryMock{}
	hash := &PasswordHasherMock{}

	// これがないとモックの呼び出しが失敗したときにパニックを起こす。
	// パニックを起こすと他のテストが実行されない。
	repo.Test(t)
	hash.Test(t)

	if setupMock != nil {
		// setupMockはモックの振る舞いを定義する関数。
		// テストケースごとにモックの振る舞いは違うため。
		setupMock(repo, hash)
	}

	srv := user.NewService(repo, hash)

	// runはテスト対象の関数を呼び出す関数。
	// テストケースごとに呼び出すServiceのレシーバ関数は違うため。
	err := run(srv)

	assert.ErrorIs(t, err, expectedErr)

	// モックの呼び出しを期待していたのに呼び出されなかったときに、
	// テストが失敗するようにするため。
	repo.AssertExpectations(t)
	hash.AssertExpectations(t)
}

func TestCreate(t *testing.T) {
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
			expectedErr: user.ErrUsernameInvalid,
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
		create := func(srv *user.Service) error {
			return srv.Create(tc.username, tc.password)
		}

		run := func(t *testing.T) {
			runTest(t, tc.setupMock, create, tc.expectedErr)
		}

		t.Run(tc.name, run)
	}
}

func TestUpdateUsername(t *testing.T) {
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
			expectedErr: user.ErrUsernameInvalid,
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
		updateUsername := func(srv *user.Service) error {
			return srv.UpdateUsername(tc.id, tc.newUsername)
		}

		run := func(t *testing.T) {
			runTest(t, tc.setupMock, updateUsername, tc.expectedErr)
		}

		t.Run(tc.name, run)
	}
}

func TestUpdatePassword(t *testing.T) {
	// CompareHashAndPasswordの引数にユーザの現在のパスワードハッシュを渡す。
	// FindByIDの返り値はUser型であり、すべてのテストケースで同じユーザを返す。
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
		updatePassword := func(srv *user.Service) error {
			return srv.UpdatePassword(tc.id, tc.oldPassword, tc.newPassword)
		}

		run := func(t *testing.T) {
			runTest(t, tc.setupMock, updatePassword, tc.expectedErr)
		}

		t.Run(tc.name, run)
	}
}
