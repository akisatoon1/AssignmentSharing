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

func TestIssueToken(t *testing.T) {
	tests := []struct {
		name          string
		groupID       int64
		requesterID   int64
		setupMock     func(*RepositoryMock, *TokenStoreMock)
		expectedToken string
		expectedErr   error
	}{
		{
			name:        "Success",
			groupID:     1,
			requesterID: 10,
			setupMock: func(repo *RepositoryMock, ts *TokenStoreMock) {
				repo.On("IsMember", int64(1), int64(10)).Return(true, nil)
				ts.On("CreateToken", int64(1)).Return("invite-token", nil)
			},
			expectedToken: "invite-token",
			expectedErr:   nil,
		},
		{
			name:        "Error: requester is not a member",
			groupID:     1,
			requesterID: 10,
			setupMock: func(repo *RepositoryMock, _ *TokenStoreMock) {
				repo.On("IsMember", int64(1), int64(10)).Return(false, nil)
			},
			expectedErr: service.ErrNotMember,
		},
		{
			name:        "Error: IsMember repository failure",
			groupID:     1,
			requesterID: 10,
			setupMock: func(repo *RepositoryMock, _ *TokenStoreMock) {
				repo.On("IsMember", int64(1), int64(10)).Return(false, repoErr)
			},
			expectedErr: repoErr,
		},
		{
			name:        "Error: CreateToken failure",
			groupID:     1,
			requesterID: 10,
			setupMock: func(repo *RepositoryMock, ts *TokenStoreMock) {
				repo.On("IsMember", int64(1), int64(10)).Return(true, nil)
				ts.On("CreateToken", int64(1)).Return("", tokenErr)
			},
			expectedErr: tokenErr,
		},
	}

	for _, tc := range tests {
		issueToken := func(srv *service.Service) (string, error) {
			return srv.IssueToken(tc.groupID, tc.requesterID)
		}

		run := func(t *testing.T) {
			runTestString(t, tc.setupMock, issueToken, tc.expectedToken, tc.expectedErr)
		}

		t.Run(tc.name, run)
	}
}

func TestAddUser(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		userID      int64
		setupMock   func(*RepositoryMock, *TokenStoreMock)
		expectedErr error
	}{
		{
			name:   "Success",
			token:  "invite-token",
			userID: 2,
			setupMock: func(repo *RepositoryMock, ts *TokenStoreMock) {
				ts.On("GetGroupIDByToken", "invite-token").Return(int64(1))
				repo.On("AddMember", int64(1), int64(2)).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "Error: invalid token",
			token:  "invalid-token",
			userID: 2,
			setupMock: func(_ *RepositoryMock, ts *TokenStoreMock) {
				ts.On("GetGroupIDByToken", "invalid-token").Return(nil)
			},
			expectedErr: service.ErrInvalidToken,
		},
		{
			name:   "Error: Repository failure",
			token:  "invite-token",
			userID: 2,
			setupMock: func(repo *RepositoryMock, ts *TokenStoreMock) {
				ts.On("GetGroupIDByToken", "invite-token").Return(int64(1))
				repo.On("AddMember", int64(1), int64(2)).Return(repoErr)
			},
			expectedErr: repoErr,
		},
	}

	for _, tc := range tests {
		addUser := func(srv *service.Service) error {
			return srv.AddUser(tc.token, tc.userID)
		}

		run := func(t *testing.T) {
			runTest(t, tc.setupMock, addUser, tc.expectedErr)
		}

		t.Run(tc.name, run)
	}
}
