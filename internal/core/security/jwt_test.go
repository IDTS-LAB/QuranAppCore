package security

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fiqrikm18/quran-app/internal/core/config"
	"github.com/golang-jwt/jwt/v5"
)

func testConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-that-is-long-enough",
			Issuer:     "quran-app-test",
			AccessTTL:  "15m",
			RefreshTTL: "168h",
		},
	}
}

func TestGenerateTokenSignsWithConfiguredValues(t *testing.T) {
	provider := NewJwtProvider(testConfig())

	before := time.Now()
	tokenString, err := provider.GenerateToken(42, "user@example.com", "admin")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	var claims JwtClaim
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-that-is-long-enough"), nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("generated token does not verify with configured secret: %v", err)
	}

	if claims.Issuer != "quran-app-test" {
		t.Fatalf("expected issuer quran-app-test, got %q", claims.Issuer)
	}
	if claims.UserID != 42 || claims.Email != "user@example.com" || claims.Role != "admin" {
		t.Fatalf("claims not carried through: %+v", claims)
	}
	wantExpiry := before.Add(15 * time.Minute)
	if claims.ExpiresAt.Time.Before(wantExpiry.Add(-time.Minute)) ||
		claims.ExpiresAt.Time.After(wantExpiry.Add(time.Minute)) {
		t.Fatalf("expiry not ~15m from now: %v", claims.ExpiresAt.Time)
	}
}

func TestGenerateTokenRejectsEmptySecret(t *testing.T) {
	cfg := testConfig()
	cfg.JWT.Secret = ""

	if _, err := NewJwtProvider(cfg).GenerateToken(1, "a@b.c", "user"); err == nil {
		t.Fatal("expected error for empty jwt secret, got nil")
	}
}

func TestGenerateTokenRejectsInvalidTTL(t *testing.T) {
	cfg := testConfig()
	cfg.JWT.AccessTTL = "not-a-duration"

	if _, err := NewJwtProvider(cfg).GenerateToken(1, "a@b.c", "user"); err == nil {
		t.Fatal("expected error for invalid access_ttl, got nil")
	}
}

func signWith(t *testing.T, secret string, claims JwtClaim) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return signed
}

func validClaims() JwtClaim {
	now := time.Now()
	return JwtClaim{
		UserID: 7,
		Email:  "user@example.com",
		Role:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "quran-app-test",
		},
	}
}

func TestValidateTokenRoundTrip(t *testing.T) {
	provider := NewJwtProvider(testConfig())

	tokenString, err := provider.GenerateToken(7, "user@example.com", "user")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	claims, err := provider.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("ValidateToken rejected a fresh token: %v", err)
	}
	if claims.UserID != 7 || claims.Email != "user@example.com" || claims.Role != "user" {
		t.Fatalf("claims not carried through: %+v", claims)
	}
}

func TestValidateTokenRejectsEmpty(t *testing.T) {
	provider := NewJwtProvider(testConfig())
	for _, input := range []string{"", "   "} {
		if _, err := provider.ValidateToken(input); !errors.Is(err, ErrTokenInvalid) {
			t.Fatalf("expected ErrTokenInvalid for %q, got %v", input, err)
		}
	}
}

func TestValidateTokenRejectsTamperedPayload(t *testing.T) {
	provider := NewJwtProvider(testConfig())

	tokenString, err := provider.GenerateToken(7, "user@example.com", "user")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 token segments, got %d", len(parts))
	}
	payload := parts[1]
	flip := "A"
	if strings.HasSuffix(payload, "A") {
		flip = "B"
	}
	parts[1] = payload[:len(payload)-1] + flip

	if _, err := provider.ValidateToken(strings.Join(parts, ".")); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid for tampered token, got %v", err)
	}
}

func TestValidateTokenRejectsNoneAlgorithm(t *testing.T) {
	provider := NewJwtProvider(testConfig())

	// Header {"alg":"none"}, payload {"user_id":7}, empty signature.
	unsigned := "eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjo3fQ."
	if _, err := provider.ValidateToken(unsigned); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid for alg=none token, got %v", err)
	}
}

func TestValidateTokenRejectsWrongAlgorithm(t *testing.T) {
	provider := NewJwtProvider(testConfig())

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"user_id":7}`))
	confused := header + "." + payload + ".junk-signature"

	if _, err := provider.ValidateToken(confused); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid for RS256 token, got %v", err)
	}
}

func TestValidateTokenRejectsWrongSecret(t *testing.T) {
	provider := NewJwtProvider(testConfig())

	forged := signWith(t, "attacker-secret", validClaims())
	if _, err := provider.ValidateToken(forged); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid for foreign secret, got %v", err)
	}
}

func TestValidateTokenRejectsExpired(t *testing.T) {
	provider := NewJwtProvider(testConfig())

	claims := validClaims()
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))
	stale := signWith(t, "test-secret-that-is-long-enough", claims)

	if _, err := provider.ValidateToken(stale); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired for expired token, got %v", err)
	}
}

func TestValidateTokenRejectsWrongIssuer(t *testing.T) {
	provider := NewJwtProvider(testConfig())

	claims := validClaims()
	claims.Issuer = "evil-issuer"
	foreign := signWith(t, "test-secret-that-is-long-enough", claims)

	if _, err := provider.ValidateToken(foreign); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid for wrong issuer, got %v", err)
	}
}

func TestValidateTokenRejectsMissingExpiry(t *testing.T) {
	provider := NewJwtProvider(testConfig())

	claims := validClaims()
	claims.ExpiresAt = nil
	eternal := signWith(t, "test-secret-that-is-long-enough", claims)

	if _, err := provider.ValidateToken(eternal); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid for token without expiry, got %v", err)
	}
}

func TestRefreshTokenRoundTrip(t *testing.T) {
	provider := NewJwtProvider(testConfig())

	before := time.Now()
	refresh, err := provider.GenerateRefreshToken(7)
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}

	claims, err := provider.ValidateRefreshToken(refresh)
	if err != nil {
		t.Fatalf("ValidateRefreshToken rejected a fresh token: %v", err)
	}
	if claims.UserID != 7 {
		t.Fatalf("expected user ID 7, got %d", claims.UserID)
	}
	if claims.TokenType != TokenTypeRefresh {
		t.Fatalf("expected type refresh, got %q", claims.TokenType)
	}
	if claims.ID == "" {
		t.Fatal("expected refresh token to carry a jti")
	}
	if claims.Email != "" || claims.Role != "" {
		t.Fatalf("refresh token must not carry authorizations: %+v", claims)
	}
	wantExpiry := before.Add(168 * time.Hour)
	if claims.ExpiresAt.Time.Before(wantExpiry.Add(-time.Minute)) ||
		claims.ExpiresAt.Time.After(wantExpiry.Add(time.Minute)) {
		t.Fatalf("expiry not ~168h from now: %v", claims.ExpiresAt.Time)
	}
}

func TestTokenTypesAreNotInterchangeable(t *testing.T) {
	provider := NewJwtProvider(testConfig())

	access, err := provider.GenerateToken(7, "user@example.com", "user")
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	refresh, err := provider.GenerateRefreshToken(7)
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}

	if _, err := provider.ValidateRefreshToken(access); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("access token must not validate as refresh, got %v", err)
	}
	if _, err := provider.ValidateAccessToken(refresh); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("refresh token must not validate as access, got %v", err)
	}
}

func TestTypedValidatorsRejectLegacyUntypedTokens(t *testing.T) {
	provider := NewJwtProvider(testConfig())

	legacy := signWith(t, "test-secret-that-is-long-enough", validClaims())

	if _, err := provider.ValidateToken(legacy); err != nil {
		t.Fatalf("untyped token should still pass type-agnostic validation, got %v", err)
	}
	if _, err := provider.ValidateAccessToken(legacy); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid for untyped access, got %v", err)
	}
	if _, err := provider.ValidateRefreshToken(legacy); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expected ErrTokenInvalid for untyped refresh, got %v", err)
	}
}

func TestIssueMintsValidPairThroughGate(t *testing.T) {
	var gate TokenGate = NewJwtProvider(testConfig())

	pair, err := gate.Issue(7, "user@example.com", "user")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected both halves of the pair, got %+v", pair)
	}
	if pair.AccessToken == pair.RefreshToken {
		t.Fatal("access and refresh tokens must differ")
	}

	access, err := gate.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("pair access half rejected: %v", err)
	}
	if access.UserID != 7 || access.Email != "user@example.com" {
		t.Fatalf("access claims wrong: %+v", access)
	}
	refresh, err := gate.ValidateRefreshToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("pair refresh half rejected: %v", err)
	}
	if refresh.UserID != 7 || refresh.ID == "" {
		t.Fatalf("refresh claims wrong: %+v", refresh)
	}
}

func TestIssueReportsExpiries(t *testing.T) {
	var gate TokenGate = NewJwtProvider(testConfig())

	before := time.Now()
	pair, err := gate.Issue(7, "user@example.com", "user")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}

	wantAccess := before.Add(15 * time.Minute)
	if pair.AccessExpiresAt.Before(wantAccess.Add(-time.Minute)) ||
		pair.AccessExpiresAt.After(wantAccess.Add(time.Minute)) {
		t.Fatalf("access expiry not ~15m from now: %v", pair.AccessExpiresAt)
	}
	wantRefresh := before.Add(168 * time.Hour)
	if pair.RefreshExpiresAt.Before(wantRefresh.Add(-time.Minute)) ||
		pair.RefreshExpiresAt.After(wantRefresh.Add(time.Minute)) {
		t.Fatalf("refresh expiry not ~168h from now: %v", pair.RefreshExpiresAt)
	}
}

func TestIssueFailsClosed(t *testing.T) {
	cfg := testConfig()
	cfg.JWT.Secret = ""

	var gate TokenGate = NewJwtProvider(cfg)
	pair, err := gate.Issue(7, "user@example.com", "user")
	if err == nil {
		t.Fatal("expected Issue error with empty secret, got nil")
	}
	if pair != (TokenPair{}) {
		t.Fatalf("expected zero pair on failure, got %+v", pair)
	}
}
