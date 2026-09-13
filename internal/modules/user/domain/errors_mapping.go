package domain

import (
	"net/http"

	appErrors "github.com/fiqrikm18/quran-app/internal/shared/errors"
)

// Mappings are registered here, next to the sentinels they describe, so the
// generic error handler never needs to import this package.
func init() {
	appErrors.RegisterError(
		ErrInvalidCredentials,
		http.StatusUnauthorized,
		"INVALID_CREDENTIALS",
		"Incorrect email or password. Please try again.",
		appErrors.ActionDetails(appErrors.ActionLogin),
	)
	appErrors.RegisterError(
		ErrEmailAlreadyUsed,
		http.StatusConflict,
		"EMAIL_TAKEN",
		"This email is already registered. Try logging in instead.",
		appErrors.ActionDetails(appErrors.ActionLogin),
	)
	appErrors.RegisterError(
		ErrEmailNotVerified,
		http.StatusForbidden,
		"EMAIL_NOT_VERIFIED",
		"Please verify your email before continuing. Check your inbox for the verification link.",
		appErrors.ActionDetails(appErrors.ActionVerifyEmail),
	)
	appErrors.RegisterError(
		ErrEmailInvalid,
		http.StatusBadRequest,
		"INVALID_EMAIL",
		"Please enter a valid email address.",
		appErrors.ActionDetails(appErrors.ActionFixRequest),
	)
	appErrors.RegisterError(
		ErrEmailEmpty,
		http.StatusBadRequest,
		"INVALID_EMAIL",
		"Please enter your email address.",
		appErrors.ActionDetails(appErrors.ActionFixRequest),
	)
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
	appErrors.RegisterError(
		ErrUserNotFound,
		http.StatusNotFound,
		"USER_NOT_FOUND",
		"We couldn't find your account.",
		appErrors.ActionDetails(appErrors.ActionLogin),
	)
}
