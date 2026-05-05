package presentation_test

import (
	"backend/internal/auth/session"
	"backend/internal/group/presentation"
	"backend/internal/group/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runTest(
	t *testing.T,
	setupMocks func(*ServiceMock, *SessionStoreMock),
	buildRequest func() *http.Request,
	expectedStatus int,
	expectedBody string,
) {
	t.Helper()

	svc := &ServiceMock{}
	svc.Test(t)
	store := &SessionStoreMock{}
	store.Test(t)

	if setupMocks != nil {
		setupMocks(svc, store)
	}

	p := presentation.NewPresentation(svc, store)
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, buildRequest())

	assert.Equal(t, expectedStatus, rr.Code)
	if expectedBody != "" {
		assert.Contains(t, rr.Body.String(), expectedBody)
	}

	svc.AssertExpectations(t)
	store.AssertExpectations(t)
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		withCookie     bool
		setupMocks     func(*ServiceMock, *SessionStoreMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Error: no cookie",
			body:           `{"name":"mygroup"}`,
			withCookie:     false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   presentation.ErrUnauthorized.Msg,
		},
		{
			name:       "Error: invalid session",
			body:       `{"name":"mygroup"}`,
			withCookie: true,
			setupMocks: func(_ *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   presentation.ErrUnauthorized.Msg,
		},
		{
			name:       "Error: malformed JSON",
			body:       `{bad json`,
			withCookie: true,
			setupMocks: func(_ *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(&session.Session{UserID: 1})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   presentation.ErrInvalidBody.Msg,
		},
		{
			name:       "Error: invalid group name",
			body:       `{"name":""}`,
			withCookie: true,
			setupMocks: func(svc *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(&session.Session{UserID: 1})
				svc.On("Create", "", int64(1)).Return("", service.ErrInvalidGroupName)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   presentation.ErrInvalidGroupName.Msg,
		},
		{
			name:       "Error: internal service error",
			body:       `{"name":"mygroup"}`,
			withCookie: true,
			setupMocks: func(svc *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(&session.Session{UserID: 1})
				svc.On("Create", "mygroup", int64(1)).Return("", serviceErr)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   presentation.ErrInternal.Msg,
		},
		{
			name:       "Success",
			body:       `{"name":"mygroup"}`,
			withCookie: true,
			setupMocks: func(svc *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(&session.Session{UserID: 1})
				svc.On("Create", "mygroup", int64(1)).Return("invite-token", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"token":"invite-token"`,
		},
	}

	for _, tc := range tests {
		buildRequest := func() *http.Request {
			req := httptest.NewRequest(http.MethodPost, "/api/groups", strings.NewReader(tc.body))
			if tc.withCookie {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "sess-abc"})
			}
			return req
		}

		run := func(t *testing.T) {
			runTest(t, tc.setupMocks, buildRequest, tc.expectedStatus, tc.expectedBody)
		}

		t.Run(tc.name, run)
	}
}

func TestIssueToken(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		withCookie     bool
		setupMocks     func(*ServiceMock, *SessionStoreMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Error: no cookie",
			path:           "/api/groups/1/invite-token",
			withCookie:     false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   presentation.ErrUnauthorized.Msg,
		},
		{
			name:       "Error: invalid session",
			path:       "/api/groups/1/invite-token",
			withCookie: true,
			setupMocks: func(_ *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   presentation.ErrUnauthorized.Msg,
		},
		{
			name:       "Error: invalid groupID",
			path:       "/api/groups/abc/invite-token",
			withCookie: true,
			setupMocks: func(_ *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(&session.Session{UserID: 1})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   presentation.ErrInvalidBody.Msg,
		},
		{
			name:       "Error: requester is not a member",
			path:       "/api/groups/1/invite-token",
			withCookie: true,
			setupMocks: func(svc *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(&session.Session{UserID: 1})
				svc.On("IssueToken", int64(1), int64(1)).Return("", service.ErrNotMember)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   presentation.ErrForbidden.Msg,
		},
		{
			name:       "Error: internal service error",
			path:       "/api/groups/1/invite-token",
			withCookie: true,
			setupMocks: func(svc *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(&session.Session{UserID: 1})
				svc.On("IssueToken", int64(1), int64(1)).Return("", serviceErr)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   presentation.ErrInternal.Msg,
		},
		{
			name:       "Success",
			path:       "/api/groups/1/invite-token",
			withCookie: true,
			setupMocks: func(svc *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(&session.Session{UserID: 1})
				svc.On("IssueToken", int64(1), int64(1)).Return("invite-token", nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"token":"invite-token"`,
		},
	}

	for _, tc := range tests {
		buildRequest := func() *http.Request {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.withCookie {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "sess-abc"})
			}
			return req
		}

		run := func(t *testing.T) {
			runTest(t, tc.setupMocks, buildRequest, tc.expectedStatus, tc.expectedBody)
		}

		t.Run(tc.name, run)
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		withCookie     bool
		setupMocks     func(*ServiceMock, *SessionStoreMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Error: no cookie",
			body:           `{"token":"invite-token"}`,
			withCookie:     false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   presentation.ErrUnauthorized.Msg,
		},
		{
			name:       "Error: invalid session",
			body:       `{"token":"invite-token"}`,
			withCookie: true,
			setupMocks: func(_ *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   presentation.ErrUnauthorized.Msg,
		},
		{
			name:       "Error: malformed JSON",
			body:       `{bad json`,
			withCookie: true,
			setupMocks: func(_ *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(&session.Session{UserID: 1})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   presentation.ErrInvalidBody.Msg,
		},
		{
			name:       "Error: invalid token",
			body:       `{"inviteToken":"invalid-token"}`,
			withCookie: true,
			setupMocks: func(svc *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(&session.Session{UserID: 1})
				svc.On("AddUser", "invalid-token", int64(1)).Return(service.ErrInvalidToken)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   presentation.ErrInvalidToken.Msg,
		},
		{
			name:       "Error: internal service error",
			body:       `{"inviteToken":"invite-token"}`,
			withCookie: true,
			setupMocks: func(svc *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(&session.Session{UserID: 1})
				svc.On("AddUser", "invite-token", int64(1)).Return(serviceErr)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   presentation.ErrInternal.Msg,
		},
		{
			name:       "Success",
			body:       `{"inviteToken":"invite-token"}`,
			withCookie: true,
			setupMocks: func(svc *ServiceMock, store *SessionStoreMock) {
				store.On("Get", "sess-abc").Return(&session.Session{UserID: 1})
				svc.On("AddUser", "invite-token", int64(1)).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tc := range tests {
		buildRequest := func() *http.Request {
			req := httptest.NewRequest(http.MethodPost, "/api/groups/join", strings.NewReader(tc.body))
			if tc.withCookie {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "sess-abc"})
			}
			return req
		}

		run := func(t *testing.T) {
			runTest(t, tc.setupMocks, buildRequest, tc.expectedStatus, tc.expectedBody)
		}

		t.Run(tc.name, run)
	}
}
