package middleware

import (
	"net/http"
	"runtime/debug"

	appErrors "github.com/fiqrikm18/quran-app/internal/shared/errors"
	httpResponse "github.com/fiqrikm18/quran-app/internal/shared/http"
	"github.com/rs/zerolog"
)

func ErrorHandler(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error().
						Interface("panic", recovered).
						Bytes("stack", debug.Stack()).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Msg("panic recovered")

					httpResponse.ErrorResponse(
						w,
						http.StatusInternalServerError,
						"INTERNAL_ERROR",
						"Something went wrong on our end. Please try again later.",
						appErrors.ActionDetails(appErrors.ActionRetryLater),
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
