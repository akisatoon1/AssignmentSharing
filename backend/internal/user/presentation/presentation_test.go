package presentation_test

import (
	"backend/internal/auth/session"
	"backend/internal/user/presentation"
	"backend/internal/user/service"
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

	// ルーティングをする
	p := presentation.NewPresentation(svc, store)
	mux := http.NewServeMux()
	p.RegisterRoutes(mux)

	// リクエストを飛ばしてテストする
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
		setupMocks     func(*ServiceMock, *SessionStoreMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			body: `{"username":"alice","password":"password123"}`,
			setupMocks: func(svc *ServiceMock, store *SessionStoreMock) {
				svc.On("Create", "alice", "password123").Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Error: malformed JSON",
			body:           `{bad json`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name: "Error: ErrUsernameInvalid",
			body: `{"username":"bad name","password":"password123"}`,
			setupMocks: func(svc *ServiceMock, store *SessionStoreMock) {
				svc.On("Create", "bad name", "password123").Return(service.ErrUsernameInvalid)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid username: must not be empty and must not contain spaces",
		},
		{
			name: "Error: ErrInvalidPassword",
			body: `{"username":"alice","password":"short"}`,
			setupMocks: func(svc *ServiceMock, store *SessionStoreMock) {
				svc.On("Create", "alice", "short").Return(service.ErrInvalidPassword)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid password: must be at least 8 characters",
		},
		{
			name: "Error: internal service error",
			body: `{"username":"alice","password":"password123"}`,
			setupMocks: func(svc *ServiceMock, store *SessionStoreMock) {
				svc.On("Create", "alice", "password123").Return(serviceErr)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error",
		},
	}

	for _, tc := range tests {
		buildRequest := func() *http.Request {
			return httptest.NewRequest(
				http.MethodPost,
				"/api/users",
				strings.NewReader(tc.body),
			)
		}

		run := func(t *testing.T) {
			runTest(
				t,
				tc.setupMocks,
				buildRequest,
				tc.expectedStatus,
				tc.expectedBody,
			)
		}

		t.Run(tc.name, run)
	}
}
