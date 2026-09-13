package application

import (
	"net/http"

	appErrors "github.com/fiqrikm18/quran-app/internal/shared/errors"
)

// Register the application-level sentinels next to their owner so the
// generic error handler never needs to import this package.
func init() {
	appErrors.RegisterError(
		ErrUnauthorized,
		http.StatusUnauthorized,
		"UNAUTHORIZED",
		"You're not authorized to do that. Please log in again.",
		appErrors.ActionDetails(appErrors.ActionLogin),
	)
}
