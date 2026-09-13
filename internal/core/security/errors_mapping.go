package security

import (
	"net/http"

	appErrors "github.com/fiqrikm18/quran-app/internal/shared/errors"
)

// Register the token sentinels next to their owner so the generic error
// handler never needs to import this package.
func init() {
	appErrors.RegisterError(
		ErrTokenExpired,
		http.StatusUnauthorized,
		"TOKEN_EXPIRED",
		"Your session has expired. Please log in again.",
		appErrors.ActionDetails(appErrors.ActionRefreshToken),
	)
	appErrors.RegisterError(
		ErrTokenInvalid,
		http.StatusUnauthorized,
		"INVALID_TOKEN",
		"Your session is no longer valid. Please log in again.",
		appErrors.ActionDetails(appErrors.ActionLogin),
	)
}
