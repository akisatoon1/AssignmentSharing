# 環境構築

## 目的
開発環境の構築を目指す。コードやdbなどは環境構築をテストするために適当なものにする。

## ディレクトリ構成
```
/my-app (プロジェクトルート)
  ├── /backend        (Goのコードを入れる)
  │    └── Dockerfile
  ├── /frontend       (Reactのコードを入れる)
  │    └── Dockerfile
  ├── /db             (DB初期化用 - 空でOK)
  ├── .env            (環境変数を共通管理)
  └── docker-compose.yml (全体をオーケストラする設定ファイル)
```

## docker compose
```.env
POSTGRES_USER=fillme
POSTGRES_PASSWORD=fillme
POSTGRES_DB=fillme
```

```docker-compose.yaml
services:
  # データベース (PostgreSQL)
  db:
    image: postgres:17-bookworm
    restart: always
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    ports:
      - "5432:5432"
    volumes:
      # データを永続化
      - ./db/data:/var/lib/postgresql/data
      # 起動時に初期テーブル作成SQLを実行
      - ./db/init:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready"]
      interval: 10s
      retries: 5
      start_period: 30s
      timeout: 10s 

  # バックエンド (Go)
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      # postgresに接続するため
      - DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@db:5432/${POSTGRES_DB}?sslmode=disable
    depends_on:
      db:
        condition: service_healthy # DBの準備ができるまで待機

  # フロントエンド (React/Vite)
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "5173:5173"
    environment:
      - VITE_API_URL=http://localhost:8080
    depends_on:
      - backend
    # ライブリロード（コード変更の即時反映）を有効にするための設定
    volumes:
      - ./frontend:/app
      - /app/node_modules # node_modulesをホストから隔離するため
```

## db
```db/init/01_schema.sql
CREATE TABLE message (
    id SERIAL PRIMARY KEY,
    content TEXT NOT NULL
);

INSERT INTO message (content) VALUES ('Hello from PostgreSQL!');
```

## backend
```bash
cd backend
go mod init backend
go mod tidy
```

```backend/main.go
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// DB接続
	dbURL := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		// CORS設定（Viteからのアクセスを許可）
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		var msg string
		err := db.QueryRow("SELECT content FROM message LIMIT 1").Scan(&msg)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": msg})
	})

	fmt.Println("Server starting at :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

```backend/Dockerfile
FROM golang:1.25-bookworm AS builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -v -o /bin/app ./...

FROM scratch
COPY --from=builder /bin/app /bin/app

EXPOSE 8080

CMD ["/bin/app"]
```

```.dockerignore
# 自分のPCでビルドしたバイナリを除外
bin/
main

# 依存関係のキャッシュを除外
vendor/

# その他
.git
.env
```

## frontend
```bash
cd frontend
npm create vite@latest . -- --template react-ts
npm install
```

```frontend/src/App.tsx
import { useEffect, useState } from 'react'

function App() {
  const [msg, setMsg] = useState("Loading...")

  useEffect(() => {
    fetch("http://localhost:8080/api/hello")
      .then(res => res.json())
      .then(data => setMsg(data.message))
      .catch(() => setMsg("Error connecting to API"))
  }, [])

  return (
    <div>
      <h1>University Assignment App</h1>
      <p>Backend says: {msg}</p>
    </div>
  )
}

export default App
```

```frontend/Dockerfile
FROM node:22-alpine

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

EXPOSE 5173

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
```

```.dockerignore
node_modules
dist
.env
.git
npm-debug.log
```

## 起動
```bash
docker compose up --build
```

frontend: `http://localhost:5173`

backend: `http://localhost:8080/api/hello`

## チームメンバーの環境構築
1. .envファイルの追加
