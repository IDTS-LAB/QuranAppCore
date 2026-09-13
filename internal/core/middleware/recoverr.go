package middleware

import (
	"net/http"
	"runtime/debug"

	appErrors "github.com/fiqrikm18/quran-app/internal/shared/errors"
	httpResponse "github.com/fiqrikm18/quran-app/internal/shared/http"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

func Recoverer(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				if err := recover(); err != nil {
					logger.Error().
						Interface("panic", err).
						Bytes("stack", debug.Stack()).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Msg("panic recovered")

					if ww.Status() == 0 {
						httpResponse.ErrorResponse(
							ww,
							http.StatusInternalServerError,
							"INTERNAL_ERROR",
							"Something went wrong on our end. Please try again later.",
							appErrors.ActionDetails(appErrors.ActionRetryLater),
						)
					}
				}
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
