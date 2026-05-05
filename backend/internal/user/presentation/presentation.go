// httpリクエストを処理してサービス層を呼び出し、httpレスポンスを返す。
// httpに関する処理を行う責務がある。

package presentation

import (
	"backend/internal/user/service"
	"encoding/json"
	"errors"
	"net/http"
)

// テストでサービス層に依存しないためにインターフェースを利用する
type UserService interface {
	Create(username string, password string) error
	UpdateUsername(id int64, newUsername string) error
	UpdatePassword(id int64, oldPassword, newPassword string) error
}

// httpで返すエラーを1か所にまとめて、管理しやすくするため
type httpError struct {
	code    int
	message string
}

func (e httpError) write(w http.ResponseWriter) {
	http.Error(w, e.message, e.code)
}

var (
	httpErrInvalidBody     = httpError{http.StatusBadRequest, "invalid request body"}
	httpErrUnauthorized    = httpError{http.StatusUnauthorized, "unauthorized"}
	httpErrInternal        = httpError{http.StatusInternalServerError, "internal server error"}
	httpErrUsernameInvalid = httpError{http.StatusBadRequest, "invalid username: must not be empty and must not contain spaces"}
	httpErrInvalidPassword = httpError{http.StatusBadRequest, "invalid password: must be at least 8 characters"}
	httpErrWrongPassword   = httpError{http.StatusUnauthorized, "wrong password"}
)

type Presentation struct {
	service UserService
	store   SessionStore
}

func NewPresentation(service UserService, store SessionStore) *Presentation {
	return &Presentation{service: service, store: store}
}

func (p *Presentation) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/users", p.Create)
	mux.HandleFunc("PATCH /api/users/username", p.UpdateUsername)
	mux.HandleFunc("PATCH /api/users/password", p.UpdatePassword)
}

func (p *Presentation) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpErrInvalidBody.write(w)
		return
	}

	if err := p.service.Create(req.Username, req.Password); err != nil {
		convertServiceErrToHttpErr(err).write(w)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func convertServiceErrToHttpErr(srvErr error) httpError {
	switch {
	case errors.Is(srvErr, service.ErrUsernameInvalid):
		return httpErrUsernameInvalid
	case errors.Is(srvErr, service.ErrInvalidPassword):
		return httpErrInvalidPassword
	case errors.Is(srvErr, service.ErrWrongPassword):
		return httpErrWrongPassword
	default:
		return httpErrInternal
	}
}
