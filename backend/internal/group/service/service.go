package service

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidGroupName = errors.New("invalid group name: must not be empty")
	ErrNotMember        = errors.New("user is not a member of this group")
	ErrInvalidToken     = errors.New("invalid invite token")
)

// データベースの操作をこのインターフェースから行う。
// テスト容易性のために定義する。
type Repository interface {
	// グループを作成した後、ユーザを追加するため、userIDを引数に取る。
	CreateGroup(grp Group, userID int64) (id int64, err error)

	IsMember(groupID, userID int64) (bool, error)
	AddMember(groupID, userID int64) error
	RemoveMember(groupID, userID int64) error
}

// グループの招待にトークンを用いるため。
// トークンの生成はランダムであり、テストがしづらいためインターフェースを定義する。
type TokenStore interface {
	CreateToken(groupID int64) (token string, err error)
	GetGroupIDByToken(token string) *int64
}

type Group struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

type Service struct {
	repo       Repository
	tokenStore TokenStore
}

func NewService(repo Repository, tokenStore TokenStore) *Service {
	return &Service{repo: repo, tokenStore: tokenStore}
}

// 空文字やスペースのみの文字列は無効とする。
func validateGroupName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidGroupName
	}
	return nil
}

// グループを作成し、招待tokenを発行して返す。
func (s *Service) Create(name string, userID int64) (string, error) {
	if err := validateGroupName(name); err != nil {
		return "", err
	}
	groupID, err := s.repo.CreateGroup(Group{Name: name}, userID)
	if err != nil {
		return "", err
	}
	return s.tokenStore.CreateToken(groupID)
}

// 既存グループの招待tokenを発行する。
// リクエスト者がグループに所属していることが必要。
func (s *Service) IssueToken(groupID int64, requesterID int64) (string, error) {
	ok, err := s.repo.IsMember(groupID, requesterID)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrNotMember
	}
	return s.tokenStore.CreateToken(groupID)
}
