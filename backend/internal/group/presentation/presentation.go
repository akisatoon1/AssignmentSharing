package presentation

import (
	"backend/internal/group/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type GroupService interface {
	Create(name string, userID int64) (token string, err error)
	IssueToken(groupID int64, requesterID int64) (string, error)
	AddUser(token string, userID int64) error
	DeleteUser(groupID int64, requesterID int64, targetUserID int64) error
}

// TODO: ほかの機能でも使うのでミドルウェアなどで処理する。
type httpError struct {
	Code int
	Msg  string
}

func (e httpError) write(w http.ResponseWriter) {
	http.Error(w, e.Msg, e.Code)
}

var (
	ErrInvalidBody      = httpError{http.StatusBadRequest, "invalid request body"}
	ErrUnauthorized     = httpError{http.StatusUnauthorized, "unauthorized"}
	ErrForbidden        = httpError{http.StatusForbidden, "forbidden"}
	ErrInvalidToken     = httpError{http.StatusNotFound, "invite token not found"}
	ErrInvalidGroupName = httpError{http.StatusBadRequest, "invalid group name: must not be empty"}
	ErrInternal         = httpError{http.StatusInternalServerError, "internal server error"}
)

type Presentation struct {
	service GroupService
	store   SessionStore
}

func NewPresentation(svc GroupService, store SessionStore) *Presentation {
	return &Presentation{service: svc, store: store}
}

func (p *Presentation) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/groups", p.Create)
	mux.HandleFunc("GET /api/groups/{groupID}/invite-token", p.IssueToken)
	mux.HandleFunc("POST /api/groups/join", p.Join)
	mux.HandleFunc("DELETE /api/groups/{groupID}/users/{userID}", p.DeleteUser)
}

func (p *Presentation) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r, p.store)
	if err != nil {
		ErrUnauthorized.write(w)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrInvalidBody.write(w)
		return
	}

	token, err := p.service.Create(req.Name, userID)
	if err != nil {
		convertServiceErrToHttpErr(err).write(w)
		return
	}

	writeTokenResponse(w, token, http.StatusCreated)
}

// レスポンスボディにおいて、Json型でTokenを返す処理が、2つの関数で使われているため共通化。
func writeTokenResponse(w http.ResponseWriter, token string, statusCode int) {
	respBody, err := json.Marshal(struct {
		Token string `json:"token"`
	}{Token: token})
	if err != nil {
		ErrInternal.write(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(respBody); err != nil {
		// TODO: ログを追加
		return
	}
}

func convertServiceErrToHttpErr(srvErr error) httpError {
	switch {
	case errors.Is(srvErr, service.ErrInvalidGroupName):
		return ErrInvalidGroupName
	case errors.Is(srvErr, service.ErrNotMember):
		return ErrForbidden
	case errors.Is(srvErr, service.ErrInvalidToken):
		return ErrInvalidToken
	default:
		return ErrInternal
	}
}
