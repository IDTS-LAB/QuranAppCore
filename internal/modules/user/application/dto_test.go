package application

import (
	"testing"

	"github.com/fiqrikm18/quran-app/internal/shared"
)

func TestRegisterRequestValidationTags(t *testing.T) {
	valid := RegisterRequest{Email: "user@example.com", Password: "s3cret-pass", DisplayName: "Aisyah"}
	if details := shared.ValidateWithDetails(&valid); details != nil {
		t.Fatalf("expected valid request, got %v", details)
	}

	cases := []RegisterRequest{
		{Email: "not-an-email", Password: "s3cret-pass", DisplayName: "Aisyah"},
		{Email: "user@example.com", Password: "short", DisplayName: "Aisyah"},
		{Email: "user@example.com", Password: "s3cret-pass", DisplayName: "A"},
		{Email: "", Password: "", DisplayName: ""},
	}
	for i, c := range cases {
		if details := shared.ValidateWithDetails(&c); details == nil {
			t.Fatalf("case %d: expected validation failure, got none", i)
		}
	}
}

func TestLoginRequestValidationTags(t *testing.T) {
	valid := LoginRequest{Email: "user@example.com", Password: "s3cret-pass"}
	if details := shared.ValidateWithDetails(&valid); details != nil {
		t.Fatalf("expected valid request, got %v", details)
	}

	cases := []LoginRequest{
		{Email: "not-an-email", Password: "s3cret-pass"},
		{Email: "user@example.com", Password: "short"},
		{Email: "", Password: ""},
	}
	for i, c := range cases {
		if details := shared.ValidateWithDetails(&c); details == nil {
			t.Fatalf("case %d: expected validation failure, got none", i)
		}
	}
}

func TestRefreshTokenRequestValidationTags(t *testing.T) {
	valid := RefreshTokenRequest{RefreshToken: "valid-refresh-token"}
	if details := shared.ValidateWithDetails(&valid); details != nil {
		t.Fatalf("expected valid request, got %v", details)
	}

	if details := shared.ValidateWithDetails(&RefreshTokenRequest{RefreshToken: ""}); details == nil {
		t.Fatalf("expected validation failure for empty refresh token, got none")
	}
}

func TestVerifyEmailRequestValidationTags(t *testing.T) {
	valid := VerifyEmailRequest{Token: "valid-verification-token"}
	if details := shared.ValidateWithDetails(&valid); details != nil {
		t.Fatalf("expected valid request, got %v", details)
	}

	if details := shared.ValidateWithDetails(&VerifyEmailRequest{Token: ""}); details == nil {
		t.Fatalf("expected validation failure for empty token, got none")
	}
}

func TestLogoutRequestValidationTags(t *testing.T) {
	valid := LogoutRequest{RefreshToken: "valid-refresh-token"}
	if details := shared.ValidateWithDetails(&valid); details != nil {
		t.Fatalf("expected valid request, got %v", details)
	}

	if details := shared.ValidateWithDetails(&LogoutRequest{RefreshToken: ""}); details == nil {
		t.Fatalf("expected validation failure for empty refresh token, got none")
	}
}
