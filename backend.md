# バックエンド

## 背景
学習のためなるべく標準ライブラリを使う。
postgresを操作するドライバのみ外部。

## 機能を整理
- 解答を見る
- ユーザとグループ
- 認証

### 解答について
#### 課題 assignment
課題の作成。グループに対して行う。
課題の削除。
課題の更新。名前、説明、締め切りの更新。グループは変更できない。
全ての課題の読み取り。ユーザが所属しているグループに所属する課題のリスト。

#### 解答 answer
解答の作成。課題に対して作成。何回でも作成でき、履歴はすべて記録される。
全ての解答の読み取り。解答の所属する課題に対して、課題を作成しているときのみ、読み取りできる。
ユーザが所属するグループに所属する課題に所属する解答のリスト。

### ユーザとグループ
グループの作成。
グループは所属ユーザの数が0人になったら、削除。ライングループと同じ。
グループにユーザを追加、削除。
招待URLを持っているユーザがグループに追加される。

ユーザの作成、削除、更新。

### 認証
ユーザネームとパスワードでログイン。クッキーを配布してセッション開始。
クッキーでセッション管理。クッキーからユーザを特定する。
ログアウト。セッションを廃棄して、フロントのクッキーも消す。
アカウント作成。ユーザを新規追加するだけ。

## パッケージを考える
### assignment
Create, Delete, Update
, List
### answer
Create, List
### user
Create, Delete, Update
### group
Create, Delete(ユーザがいないとき)
, AddUser, DeleteUser(user側につくってもよいのかな)
### auth
Login, Logout
### session
Create, Get, Destroy

## アーキテクチャ
```
backend/
├── internal/
│   ├── answer/
│   ├── assignment/
│   ├── auth/
│   │   └── session/
│   ├── db/
│   ├── group/
│   └── user/
└── main.go
```
機能ごとにパッケージを分けた。
それぞれのpackageにservice.go, repo.go, controller.goのようなファイルを作る。

## 実装のこだわり
### Password Hash
Argon2を使用した。下のサイトを参考にパラメータを設定した。
最小構成としてメモリ19Mib, 反復回数2, 並列度1。
[OWASP](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#argon2id)
```
実際のコード
```
> これらの構成設定は同等のレベルの防御を提供し、唯一の違いはCPUとRAMの使用量のトレードオフです。
  より、パフォーマンスに問題があれば、パラメータを変更するかも。

また、ソルトも追加した。
[参考記事](https://qiita.com/ockeghem/items/d7324d383fb7c104af58)
[同じく](https://developer.mozilla.org/ja/docs/Glossary/Salt)
[ソルトの長さ](https://github.com/alexedwards/argon2id#changing-the-parameters)
