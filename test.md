# テストについて
## 背景
`user.Create(username, password) error`の動作をテストしたい。
値を受け取ってハッシュ化してrepositoryに保存するだけの関数。
- repositoryに適切なオブジェクトを保存できているか
- repositoryで発生したエラーを返しているか
- 入力を検証できているか。
- hash生成関数に適切な引数を渡しているか。
- hash生成関数で発生したエラーを返しているか。

## 依存関係
### db
dbに依存しないでテストするために、Repository Interface`Save(usr User)`を用意した。目的は、テストを容易にするのと、dbに影響を与えないため。
### hash
salt, hashはどうやってセキュリティに問題がないのかをチェックするのか。手動で値を見ればわかるが、自動テストにならない。
わざわざDIするのは過剰だと思っていたが、パスワードの比較でauthのログインでも使うので、hashやsaltの値を取得するのはDIでスタブを作る。
ここでは与えられたhash, saltがちゃんとRepositoryに保存できていることを確認する。
別パッケージ(hashと仮定)で提供するAPIを考える。hash計算はログイン、新規登録両方で使うが、salt生成は新規登録のみで使う。
これは、userとhashのどちらで行うべきか。[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)にはCompareHashAndPasswordやGenerateFromPassword
といったAPIが提供されているし、authやuserは具体的な照合や生成ロジックは負わないようにするため。

## テストケース
- username, passwordがhash関数に適切に渡されていて、usernameとpasswordHashがリポジトリに適切に渡されていれば成功。
- usernameが空のときエラーで成功。
- usernameがスペースを含むときエラー
- passwordが空のときエラー
- passwordの長さが7のときエラー
- hash関数がエラーを返したときエラー
- リポジトリがエラーを返したときエラー
### 問題点
リポジトリやハッシュ関数に適切な値を渡しているかどうかをテストしようとすると、リポジトリに適切な実装の詳細もテストしてしまっている。

## モック
mock.Mockを埋め込む。o.Called()で呼ばれた引数を記録？args.Int(0)でのちに記述するReturnの値を適切な型で返す？という理解
## 期待
mock.On("Method", arg1, arg2).Return(ret1, ret2)で期待できる。引数が一致しない, Onが未定義などの、パニック条件がある？
## panic
CalledはパニックするがTest(t)を渡すことで、メッセージの見やすさ優先でパニックが起こらないようにする。
パニック条件をしりたい。
## t.Run
Test(t)を渡しても、テストケース全てを実行せずに途中で止まる。t.Runを使用して全てサブテストにすることで回避。


## 読みづらい
```
package user_test

import (
	"backend/internal/user"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	t.Run("Positive case", func(t *testing.T) {
		tests := []struct {
			name          string
			username      string
			password      string
			hash          string // hashGeneratorが返すハッシュを指定
			expectedSaved user.User
		}{
			{
				name:     "Valid user creation",
				username: "testuser",
				password: "testPassword",
				hash:     "thisisHaaash!",
				expectedSaved: user.User{
					Username:     "testuser",
					PasswordHash: "thisisHaaash!",
				},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hashGenerator := &HashGeneratorMock{}
				hashGenerator.Test(t)
				hashGenerator.On("GenerateFromPassword", test.password).Return(test.hash, nil)

				repo := &RepositoryMock{}
				repo.Test(t)
				repo.On("Save", test.expectedSaved).Return(nil)

				srv := user.NewService(repo, hashGenerator)
				err := srv.Create(test.username, test.password)

				assert.NoError(t, err)
				hashGenerator.AssertExpectations(t)
				repo.AssertExpectations(t)
			})
		}
	})

	t.Run("Negative case. Invalid Input.", func(t *testing.T) {
		tests := []struct {
			name           string
			username       string
			password       string
			expectedErrStr string
		}{
			{
				name:           "Empty username",
				username:       "",
				password:       "aspdofj1239817",
				expectedErrStr: "username is required",
			},
			{
				name:           "Weak password",
				username:       "testuser",
				password:       "123",
				expectedErrStr: "password is weak",
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				hashGenerator := &HashGeneratorMock{}
				hashGenerator.Test(t)

				repo := &RepositoryMock{}
				repo.Test(t)

				srv := user.NewService(repo, hashGenerator)
				err := srv.Create(test.username, test.password)

				assert.EqualError(t, err, test.expectedErrStr)
				hashGenerator.AssertNotCalled(t, "GenerateFromPassword", mock.Anything)
				repo.AssertNotCalled(t, "Save", mock.Anything)
			})
		}
	})

	t.Run("Negative case. Propagates hash generator error", func(t *testing.T) {
		hashGenerator := &HashGeneratorMock{}
		hashGenerator.Test(t)
		hashGenerator.On("GenerateFromPassword", mock.Anything).Return("", errors.New("hash generator error"))

		repo := &RepositoryMock{}
		repo.Test(t)

		srv := user.NewService(repo, hashGenerator)
		err := srv.Create("tt.username", "tt.password")

		assert.Error(t, err)
		hashGenerator.AssertExpectations(t)
		repo.AssertNotCalled(t, "Save", mock.Anything)
	})

	t.Run("Negative case. Propagates repository error", func(t *testing.T) {
		repo := &RepositoryMock{}
		repo.Test(t)
		repo.On("Save", mock.Anything).Return(errors.New("repository error"))

		hashGenerator := &HashGeneratorMock{}
		hashGenerator.Test(t)
		hashGenerator.On("GenerateFromPassword", mock.Anything).Return("", nil).Maybe()

		srv := user.NewService(repo, hashGenerator)
		err := srv.Create("tt.username", "tt.password")

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}
```

モックの振る舞いが違っていることでテストケースをまとめられなくて、複雑になっている。
モックの振る舞いもテストケースで指定できるようにした。
```
package user_test

import (
	"backend/internal/user"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
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
			password: "testPassword",
			setupMocks: func(repo *RepositoryMock, hash *HashGeneratorMock) {
				hash.On("GenerateFromPassword", "testPassword").Return("thisisHaaash!", nil)
				repo.On("Save", user.User{Username: "testuser", PasswordHash: "thisisHaaash!"}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "Error: Empty username",
			username:    "",
			password:    "password123",
			setupMocks:  func(repo *RepositoryMock, hash *HashGeneratorMock) {},
			expectedErr: errors.New("username is required"),
		},
		{
			name:        "Error: Empty password",
			username:    "testuser",
			password:    "",
			setupMocks:  func(repo *RepositoryMock, hash *HashGeneratorMock) {},
			expectedErr: errors.New("password is required"),
		},
		{
			name:     "Error: Hash generator failure",
			username: "testuser",
			password: "testPassword",
			setupMocks: func(repo *RepositoryMock, hash *HashGeneratorMock) {
				hash.On("GenerateFromPassword", "testPassword").Return("", errors.New("hash generator error"))
			},
			expectedErr: errors.New("hash generator error"),
		},
		{
			name:     "Error: Repository save failure",
			username: "testuser",
			password: "testPassword",
			setupMocks: func(repo *RepositoryMock, hash *HashGeneratorMock) {
				hash.On("GenerateFromPassword", "testPassword").Return("thisisHaaash!", nil)
				repo.On("Save", mock.Anything).Return(errors.New("repository error"))
			},
			expectedErr: errors.New("repository error"),
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
```

下の挙動。
Test(t)はこのmockがパニックを起こさないでテストを失敗させるため。
AssertExpectationsはmockが呼び出されるべきテストで、呼び出されなかったときに失敗させるため。
On.Returnで期待される引数と返り値を定義できる。
expected: call => actual: call, 成功
expected: call => actual: not call, AssertExpectationsで失敗
expected: not call => actual: call, モック内のm.Called()が失敗する
expected: not call => actual: not call, 成功
expected: a call => actual: b call(引数が違う): モック内のm.Called()が失敗する
よってすべてをテストできている。
```
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
```

## 参考
[golang test best practice](https://grid.gg/testing-in-go-best-practices-and-tips/)
[mock]()
[assert]()
