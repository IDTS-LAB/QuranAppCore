package middleware

import (
	"context"
	stderrors "errors"
	"net/http"
	"strings"

	"github.com/fiqrikm18/quran-app/internal/core/security"
	appErrors "github.com/fiqrikm18/quran-app/internal/shared/errors"
	httpResponse "github.com/fiqrikm18/quran-app/internal/shared/http"
)

// claimsKey is unexported so only ClaimsFrom can read it.
type claimsKey struct{}

// Shared wording so both auth middlewares speak the same language.
const (
	msgAuthRequired = "Please log in to continue."
	msgTokenExpired = "Your session has expired. Please log in again."
)

// ClaimsFrom returns the authenticated claims stored by RequireAuth.
func ClaimsFrom(ctx context.Context) (*security.JwtClaim, bool) {
	claims, ok := ctx.Value(claimsKey{}).(*security.JwtClaim)
	return claims, ok && claims != nil
}

// RequireAuth wraps a Handler so it only runs with a valid access token.
// The verified claims are stored in the request context (see ClaimsFrom).
// Expired tokens fail with TOKEN_EXPIRED so clients know to use the refresh
// flow; everything else fails with UNAUTHORIZED and no further detail.
func RequireAuth(gate security.TokenGate) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || strings.TrimSpace(token) == "" {
				return appErrors.UnauthorizedWithDetails(msgAuthRequired, appErrors.ActionDetails(appErrors.ActionLogin))
			}

			claims, err := gate.ValidateAccessToken(token)
			if err != nil {
				if stderrors.Is(err, security.ErrTokenExpired) {
					return appErrors.New(
						http.StatusUnauthorized,
						"TOKEN_EXPIRED",
						msgTokenExpired,
						appErrors.ActionDetails(appErrors.ActionRefreshToken),
					)
				}
				return appErrors.UnauthorizedWithDetails(msgAuthRequired, appErrors.ActionDetails(appErrors.ActionLogin))
			}

			return next(w, r.WithContext(context.WithValue(r.Context(), claimsKey{}, claims)))
		}
	}
}

// RequireAuthMiddleware returns a chi-compatible middleware that validates access tokens.
func RequireAuthMiddleware(gate security.TokenGate) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || strings.TrimSpace(token) == "" {
				httpResponse.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", msgAuthRequired, appErrors.ActionDetails(appErrors.ActionLogin))
				return
			}

			claims, err := gate.ValidateAccessToken(token)
			if err != nil {
				if stderrors.Is(err, security.ErrTokenExpired) {
					httpResponse.ErrorResponse(w, http.StatusUnauthorized, "TOKEN_EXPIRED", msgTokenExpired, appErrors.ActionDetails(appErrors.ActionRefreshToken))
					return
				}
				httpResponse.ErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", msgAuthRequired, appErrors.ActionDetails(appErrors.ActionLogin))
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey{}, claims)))
		})
	}
}
