package command

import (
	"context"
	"testing"
	"time"

	"github.com/fiqrikm18/quran-app/internal/core/security"
	"github.com/fiqrikm18/quran-app/internal/modules/user/domain"
)

type stubUserRepo struct {
	user *domain.User
}

func (s *stubUserRepo) Create(context.Context, *domain.User, string) error { return nil }
func (s *stubUserRepo) FindByEmail(_ context.Context, _ string) (*domain.User, error) {
	return s.user, nil
}
func (s *stubUserRepo) FindByID(_ context.Context, _ uint64) (*domain.User, error) {
	return s.user, nil
}
func (s *stubUserRepo) MarkVerified(context.Context, uint64) error { return nil }

type stubSessionRepo struct {
	savedExpiresAt time.Time
}

func (s *stubSessionRepo) Save(_ context.Context, _ uint64, _ string, expiresAt time.Time) error {
	s.savedExpiresAt = expiresAt
	return nil
}
func (s *stubSessionRepo) FindByJTIHash(_ context.Context, _ string) (*domain.UserSession, error) {
	return nil, nil
}
func (s *stubSessionRepo) Revoke(context.Context, string) error { return nil }

type stubHasher struct{}

func (stubHasher) Hash(string) (string, error)  { return "hash", nil }
func (stubHasher) Compare(string, string) error { return nil }

type stubGate struct {
	pair security.TokenPair
}

func (s *stubGate) Issue(uint, string, string) (security.TokenPair, error) {
	return s.pair, nil
}
func (s *stubGate) ValidateAccessToken(string) (*security.JwtClaim, error)      { return nil, nil }
func (s *stubGate) ValidateRefreshToken(string) (*security.JwtClaim, error)     { return nil, nil }
func (s *stubGate) ValidateVerifyEmailToken(string) (*security.JwtClaim, error) { return nil, nil }
func (s *stubGate) GenerateVerifyEmailToken(uint, string) (string, error)       { return "", nil }

func TestLoginResponseCarriesExpiry(t *testing.T) {
	accessExpiresAt := time.Now().Add(15 * time.Minute).Truncate(time.Second)
	refreshExpiresAt := time.Now().Add(168 * time.Hour).Truncate(time.Second)
	pair := security.TokenPair{
		AccessToken:      "access",
		RefreshToken:     "refresh",
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}
	sessions := &stubSessionRepo{}
	handler := NewLoginCommandHandler(
		&stubUserRepo{user: &domain.User{
			ID:              7,
			Email:           "user@example.com",
			PasswordHash:    "hash",
			EmailVerifiedAt: time.Now(),
		}},
		sessions,
		stubHasher{},
		&stubGate{pair: pair},
	)

	resp, err := handler.Exec(context.Background(), &LoginCommand{
		Email:    "user@example.com",
		Password: "s3cret-pass",
	})
	if err != nil {
		t.Fatalf("Exec returned error: %v", err)
	}
	if resp.AccessToken != "access" || resp.RefreshToken != "refresh" {
		t.Fatalf("tokens not carried through: %+v", resp)
	}
	if !resp.ExpiresAt.Equal(accessExpiresAt) {
		t.Fatalf("expected expires_at %v, got %v", accessExpiresAt, resp.ExpiresAt)
	}
	if !sessions.savedExpiresAt.Equal(refreshExpiresAt) {
		t.Fatalf("expected session expiry %v, got %v", refreshExpiresAt, sessions.savedExpiresAt)
	}
}
