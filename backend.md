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

また、ソルトも追加した。user.Createのテストより、ハッシュやソルト生成は別パッケージとして切り分けるので、[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
に提供されているAPIを参考に、saltをAPI利用側で管理しないようにする。よってdbにはsaltカラムは用意しない。
[参考記事](https://qiita.com/ockeghem/items/d7324d383fb7c104af58)
[同じく](https://developer.mozilla.org/ja/docs/Glossary/Salt)
[ソルトの長さ](https://github.com/alexedwards/argon2id#changing-the-parameters)

### user Create()
user.Createをテストしたい。
#### 何をテストしたいのか
入力値の検証ができているのか、ソルトとハッシュ値で保存しているか、
データベースに保存できているのか、データベースのエラーは返すのか。
#### テスト設計
dbに依存しないでテストするために、Repository Interface`Save(usr User)`を用意した。
目的は、テストを容易にするのと、dbに影響を与えないため。
- Saveに渡す引数の値で、Username, PasswordHash, PasswordSaltが適切な値か。(正常系)
- 重要: saltがランダムになっており、passwordHashは平文を推測できないものであるか。
- usernameの長さが0
- passwordの強度が弱い(チェック方法はあとで考える), 長さ制限
- データベース由来のエラーを返すのか。(そもそもそれを定義できていない)
#### 実装
t.Runで分けた。理由はSaveに渡される引数をテストする必要がないものもあったりして、汎用的に実装(tests)すると分かりづらくなるから。
モックを作って、Saveへ渡す引数をチェック。
チェック項目は、err, Saved(username, hash, salt), でusername passwordの指定。
[golang test best practice](https://grid.gg/testing-in-go-best-practices-and-tips/)
車輪の再発明を避けるためにassertとmock(testify)を使った。
mockの使い方は別の記事でやろう。
##### salt hashのテスト
salt, hashはどうやってセキュリティに問題がないのかをチェックするのか。手動で値を見ればわかるが、自動テストにならない。
わざわざDIするのは過剰だと思っていたが、パスワードの比較でauthのログインでも使うので、hashやsaltの値を取得するのはDIでスタブを作る。
ここでは与えられたhash, saltがちゃんとRepositoryに保存できていることを確認する。
別パッケージ(hashと仮定)で提供するAPIを考える。hash計算はログイン、新規登録両方で使うが、salt生成は新規登録のみで使う。
これは、userとhashのどちらで行うべきか。[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)にはCompareHashAndPasswordやGenerateFromPassword
といったAPIが提供されているし、authやuserは具体的な照合や生成ロジックは負わないようにするため。

#### 依存先のエラーにはどんな種類があるのか
Repositoryは、ユーザ側(usernameが一意ではない)とdb側に責任があるもので別れる。
全てのエラーを区別なく返しているのでそれが分からない。Repository側でエラーを抽象化して定義して、
伝搬させる。Create側でも抽象化する必要はあるのか？