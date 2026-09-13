package middleware

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fiqrikm18/quran-app/internal/core/security"
	userApplication "github.com/fiqrikm18/quran-app/internal/modules/user/application"
	"github.com/fiqrikm18/quran-app/internal/modules/user/domain"
	"github.com/rs/zerolog"
)

type errorEnvelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   struct {
		Code    string         `json:"code"`
		Details map[string]any `json:"details"`
	} `json:"error"`
}

func serveHandlerErr(t *testing.T, err error) (int, errorEnvelope) {
	t.Helper()
	logger := zerolog.New(zerolog.NewTestWriter(t))

	failing := func(http.ResponseWriter, *http.Request) error { return err }
	rec := httptest.NewRecorder()
	Handle(logger, failing).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	var body errorEnvelope
	if decErr := json.Unmarshal(rec.Body.Bytes(), &body); decErr != nil {
		t.Fatalf("expected JSON error body: %v (%s)", decErr, rec.Body.String())
	}
	return rec.Code, body
}

func TestKnownErrorsMapToFriendlyResponses(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		status     int
		code       string
		contains   string
		mustNotHas []string
	}{
		{"invalid credentials", domain.ErrInvalidCredentials, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Incorrect email or password", nil},
		{"email taken", domain.ErrEmailAlreadyUsed, http.StatusConflict, "EMAIL_TAKEN", "already registered", nil},
		{"email not verified", domain.ErrEmailNotVerified, http.StatusForbidden, "EMAIL_NOT_VERIFIED", "verify your email", nil},
		{"bad email", domain.ErrEmailInvalid, http.StatusBadRequest, "INVALID_EMAIL", "valid email address", nil},
		{"domain token expired", domain.ErrTokenExpired, http.StatusUnauthorized, "TOKEN_EXPIRED", "expired", nil},
		{"domain token invalid", domain.ErrTokenInvalid, http.StatusUnauthorized, "INVALID_TOKEN", "log in again", nil},
		{"user not found", domain.ErrUserNotFound, http.StatusNotFound, "USER_NOT_FOUND", "couldn't find your account", nil},
		{"app unauthorized", userApplication.ErrUnauthorized, http.StatusUnauthorized, "UNAUTHORIZED", "log in again", nil},
		{"security token expired", security.ErrTokenExpired, http.StatusUnauthorized, "TOKEN_EXPIRED", "expired", nil},
		{"security token invalid", security.ErrTokenInvalid, http.StatusUnauthorized, "INVALID_TOKEN", "log in again", nil},
		{
			"wrapped sentinel still maps",
			fmt.Errorf("session lookup failed: %w", domain.ErrTokenInvalid),
			http.StatusUnauthorized, "INVALID_TOKEN", "log in again",
			[]string{"session lookup failed"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			status, body := serveHandlerErr(t, c.err)
			if status != c.status {
				t.Fatalf("expected status %d, got %d (%s)", c.status, status, body.Message)
			}
			if body.Success {
				t.Fatalf("expected success=false, got %+v", body)
			}
			if body.Error.Code != c.code {
				t.Fatalf("expected code %q, got %q", c.code, body.Error.Code)
			}
			if len(body.Error.Details) == 0 {
				t.Fatalf("expected non-empty error.details for %q, got none", c.code)
			}
			if !strings.Contains(body.Message, c.contains) {
				t.Fatalf("expected message to contain %q, got %q", c.contains, body.Message)
			}
			for _, leaked := range c.mustNotHas {
				if strings.Contains(body.Message, leaked) {
					t.Fatalf("message leaks internals %q: %q", leaked, body.Message)
				}
			}
		})
	}
}

func TestUnknownErrorsStayGeneric(t *testing.T) {
	for _, err := range []error{
		stderrors.New("jwt secret is not configured"),
		stderrors.New("pq: duplicate key value violates unique constraint"),
		fmt.Errorf("bcrypt compare: %w", stderrors.New("crypto/bcrypt: hashedPassword too short")),
	} {
		status, body := serveHandlerErr(t, err)
		if status != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d for %v", status, err)
		}
		if body.Error.Code != "INTERNAL_ERROR" {
			t.Fatalf("expected INTERNAL_ERROR, got %q", body.Error.Code)
		}
		if len(body.Error.Details) == 0 {
			t.Fatalf("expected non-empty error.details for fallback, got none")
		}
		if !strings.Contains(body.Message, "Something went wrong") {
			t.Fatalf("expected friendly fallback, got %q", body.Message)
		}
		for _, leaked := range []string{"jwt secret", "pq:", "bcrypt"} {
			if strings.Contains(body.Message, leaked) {
				t.Fatalf("message leaks internals %q: %q", leaked, body.Message)
			}
		}
	}
}
