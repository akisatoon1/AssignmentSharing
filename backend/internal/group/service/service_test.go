package service_test

import (
	"backend/internal/group/service"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 返り値がerrorのみの関数をテストするため。
// テストでのモックやアサートのボイラープレートを共通化した。
// runTestStringは返り値が(string, error)の関数をテストするため。
// 今回はたまたま返り値のパターンが(error)と(string, error)だったので、この2つの関数で対応。
func runTest(
	t *testing.T,
	setupMock func(*RepositoryMock, *TokenStoreMock),
	run func(*service.Service) error,
	expectedErr error,
) {
	t.Helper()

	repo := &RepositoryMock{}
	tokenStore := &TokenStoreMock{}

	repo.Test(t)
	tokenStore.Test(t)

	if setupMock != nil {
		setupMock(repo, tokenStore)
	}

	srv := service.NewService(repo, tokenStore)
	err := run(srv)

	assert.ErrorIs(t, err, expectedErr)

	repo.AssertExpectations(t)
	tokenStore.AssertExpectations(t)
}

// runTestとほぼ同じだが、返り値が(string, error)の関数をテストするためのもの。
// 2つ似たような関数を作成した理由は、runTest()にて記載。
func runTestString(
	t *testing.T,
	setupMock func(*RepositoryMock, *TokenStoreMock),
	run func(*service.Service) (string, error),
	expectedString string,
	expectedErr error,
) {
	t.Helper()

	repo := &RepositoryMock{}
	tokenStore := &TokenStoreMock{}

	repo.Test(t)
	tokenStore.Test(t)

	if setupMock != nil {
		setupMock(repo, tokenStore)
	}

	srv := service.NewService(repo, tokenStore)
	str, err := run(srv)

	assert.ErrorIs(t, err, expectedErr)
	assert.Equal(t, expectedString, str)

	repo.AssertExpectations(t)
	tokenStore.AssertExpectations(t)
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name          string
		groupName     string
		userID        int64
		setupMock     func(*RepositoryMock, *TokenStoreMock)
		expectedToken string
		expectedErr   error
	}{
		{
			name:      "Success",
			groupName: "mygroup",
			userID:    1,
			setupMock: func(repo *RepositoryMock, ts *TokenStoreMock) {
				repo.On("CreateGroup", service.Group{Name: "mygroup"}, int64(1)).Return(int64(1), nil)
				ts.On("CreateToken", int64(1)).Return("invite-token", nil)
			},
			expectedToken: "invite-token",
			expectedErr:   nil,
		},
		{
			name:        "Error: empty name",
			groupName:   "",
			userID:      1,
			expectedErr: service.ErrInvalidGroupName,
		},
		{
			name:        "Error: whitespace only name",
			groupName:   "   ",
			userID:      1,
			expectedErr: service.ErrInvalidGroupName,
		},
		{
			name:      "Error: Repository CreateGroup failure",
			groupName: "mygroup",
			userID:    1,
			setupMock: func(repo *RepositoryMock, _ *TokenStoreMock) {
				repo.On("CreateGroup", service.Group{Name: "mygroup"}, int64(1)).Return(int64(0), repoErr)
			},
			expectedErr: repoErr,
		},
		{
			name:      "Error: TokenStore CreateToken failure",
			groupName: "mygroup",
			userID:    1,
			setupMock: func(repo *RepositoryMock, ts *TokenStoreMock) {
				repo.On("CreateGroup", service.Group{Name: "mygroup"}, int64(1)).Return(int64(1), nil)
				ts.On("CreateToken", int64(1)).Return("", tokenErr)
			},
			expectedErr: tokenErr,
		},
	}

	for _, tc := range tests {
		create := func(srv *service.Service) (string, error) {
			return srv.Create(tc.groupName, tc.userID)
		}

		run := func(t *testing.T) {
			runTestString(t, tc.setupMock, create, tc.expectedToken, tc.expectedErr)
		}

		t.Run(tc.name, run)
	}
}
