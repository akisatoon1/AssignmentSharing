# DBの設計

## 機能整理
- ユーザログイン機能
- ユーザのグループ化機能
- 回答データ見る機能

前2つはよくある機能なので割愛。

### 回答データ見る機能について
あるユーザAがユーザBの回答を見ることができる条件は次の2つ。
- **Aも同じ課題に対して回答を投稿**
- A, Bは同じグループに所属

## テーブル設計の概要

### テーブル
- users
- groups
- user groups
- assignments
- answers

### 各テーブルの詳細
#### users
- id
- username (一意)
- password hash

#### groups
- id
- group name

#### user groups
userが複数のグループに所属するための、多対多の中間テーブル。

- user id
- group id

#### assignments
- id
- assginement name
- group id
- deadline (締め切りを過ぎた課題には解答できない)

#### answers
- id
- content
- user id
- assignment id
- created at
- updated at

### 懸念点
#### あるユーザが同一名の2つ以上のグループに所属しているときが存在する。
これには対処しない。所属メンバーによっユーザはグループを判別できるだろう。

#### 解答の更新履歴を保存したい
ただ乗りを防ぐため。てきとうな解答をしてから、周りの解答を見て更新することができる。

=> 

アプリケーション側でanswersに対してUPDATE操作をしないようにする。

解答を更新したいときはCREATEする。

### パフォーマンスは
アプリ開発時のsqlを見て決める。
explainで中でどうやって実行しているのかを見る。
sqlのアンチパターンをチェックする。

#### 自動でインデックスの条件(Postgresの場合)
Gemini情報だから、ソースはない
- primary key
- unique
- exclude
**外部参照は作成されない**

## テーブルのsql
**参照する**
