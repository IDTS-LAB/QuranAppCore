package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fiqrikm18/quran-app/internal/core/config"
	"github.com/fiqrikm18/quran-app/internal/core/security"
	"github.com/rs/zerolog"
)

func authTestGate() security.TokenGate {
	return security.NewJwtProvider(&config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-that-is-long-enough",
			Issuer:     "quran-app-test",
			AccessTTL:  "15m",
			RefreshTTL: "168h",
		},
	})
}

// protected echoes the authenticated user ID; it only runs past RequireAuth.
func protected(w http.ResponseWriter, r *http.Request) error {
	claims, ok := ClaimsFrom(r.Context())
	if !ok {
		w.WriteHeader(http.StatusTeapot)
		return nil
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"user_id": claims.UserID})
	return nil
}

func serveProtected(t *testing.T, gate security.TokenGate, authorization string) *httptest.ResponseRecorder {
	t.Helper()
	logger := zerolog.New(zerolog.NewTestWriter(t))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()

	Handle(logger, RequireAuth(gate)(protected)).ServeHTTP(rec, req)
	return rec
}

func TestRequireAuthPassesWithValidToken(t *testing.T) {
	gate := authTestGate()
	access, err := gate.Issue(7, "user@example.com", "user")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}

	rec := serveProtected(t, gate, "Bearer "+access.AccessToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON body: %v", err)
	}
	if body["user_id"] != float64(7) {
		t.Fatalf("expected claims user_id 7 in context, got %v", body)
	}
}

func TestRequireAuthRejectsMissingHeader(t *testing.T) {
	rec := serveProtected(t, authTestGate(), "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "UNAUTHORIZED") {
		t.Fatalf("expected UNAUTHORIZED code, got: %s", body)
	}
}

func TestRequireAuthRejectsForgedToken(t *testing.T) {
	rec := serveProtected(t, authTestGate(), "Bearer forged.token.here")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuthRejectsRefreshToken(t *testing.T) {
	gate := authTestGate()
	pair, err := gate.Issue(7, "user@example.com", "user")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}

	rec := serveProtected(t, gate, "Bearer "+pair.RefreshToken)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for refresh-as-access, got %d", rec.Code)
	}
}

func TestRequireAuthSignalsExpiredDistinctly(t *testing.T) {
	expiredGate := security.NewJwtProvider(&config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-that-is-long-enough",
			Issuer:     "quran-app-test",
			AccessTTL:  "-1m",
			RefreshTTL: "168h",
		},
	})
	pair, err := expiredGate.Issue(7, "user@example.com", "user")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}

	rec := serveProtected(t, expiredGate, "Bearer "+pair.AccessToken)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "TOKEN_EXPIRED") {
		t.Fatalf("expected TOKEN_EXPIRED code, got: %s", body)
	}
}
