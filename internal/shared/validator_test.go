package shared

import (
	"strings"
	"testing"

	"github.com/fiqrikm18/quran-app/internal/modules/user/application"
)

func TestValidateWithDetailsFriendlyMessages(t *testing.T) {
	details := ValidateWithDetails(&application.RegisterRequest{
		Email:       "not-an-email",
		Password:    "short",
		DisplayName: "A",
	})
	if details == nil {
		t.Fatal("expected validation details, got none")
	}

	// Keys use the JSON names clients send, not Go struct names.
	if msg, ok := details["email"]; !ok || msg != "Email must be a valid email address." {
		t.Fatalf("expected friendly email message, got %q", msg)
	}
	if msg, ok := details["password"]; !ok || !strings.Contains(msg, "at least 8 characters") {
		t.Fatalf("expected friendly password message, got %q", msg)
	}
	if msg, ok := details["display_name"]; !ok || !strings.Contains(msg, "at least 2 characters") {
		t.Fatalf("expected friendly display_name message, got %q", msg)
	}
}

func TestValidateWithDetailsRequiredMessages(t *testing.T) {
	details := ValidateWithDetails(&application.LoginRequest{})
	if details == nil {
		t.Fatal("expected validation details, got none")
	}

	if msg := details["email"]; !strings.Contains(msg, "required") {
		t.Fatalf("expected required message for email, got %q", msg)
	}
	if msg := details["password"]; !strings.Contains(msg, "required") {
		t.Fatalf("expected required message for password, got %q", msg)
	}
}
