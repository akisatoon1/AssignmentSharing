package presentation

import (
	"backend/internal/auth/session"
	"errors"
	"net/http"
)

var ErrSessionNotFound = errors.New("session not found")

type SessionStore interface {
	Get(id string) *session.Session
}

// TODO: 他のパッケージでも使うので、汎用的に使えるようにする。
func getUserIDFromRequest(r *http.Request, store SessionStore) (int64, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return 0, err
	}
	sess := store.Get(cookie.Value)
	if sess == nil {
		return 0, ErrSessionNotFound
	}
	return sess.UserID, nil
}
