package middleware

import (
	"net/http"

	appErrors "github.com/fiqrikm18/quran-app/internal/shared/errors"
	httpResponse "github.com/fiqrikm18/quran-app/internal/shared/http"
	"github.com/rs/zerolog"
)

type Handler func(http.ResponseWriter, *http.Request) error

func Handle(logger zerolog.Logger, handler Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			handleError(logger, w, r, err)
		}
	})
}

func handleError(logger zerolog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	if appErr, ok := err.(*appErrors.AppError); ok {
		if appErr.Status >= http.StatusInternalServerError {
			logServerError(logger, appErr)
		}

		respondError(w, appErr.Status, appErr.Code, appErr.Message, appErr.Details)
		return
	}

	// Known errors resolve through the shared registry, which feature
	// packages populate themselves. This keeps the handler generic: it
	// never imports a domain package, yet still returns friendly,
	// correctly-coded responses instead of a generic 500.
	if mapped := appErrors.FromError(err); mapped != nil {
		if mapped.Status >= http.StatusInternalServerError {
			logServerError(logger, mapped)
		}

		respondError(w, mapped.Status, mapped.Code, mapped.Message, mapped.Details)
		return
	}

	logUnhandled(logger, r, err)
	respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong on our end. Please try again later.", appErrors.ActionDetails(appErrors.ActionRetryLater))
}

func logUnhandled(logger zerolog.Logger, r *http.Request, err error) {
	logger.Error().
		Err(err).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg("unhandled application error")
}

func logServerError(logger zerolog.Logger, err *appErrors.AppError) {
	logger.Error().
		Err(err).
		Str("code", err.Code).
		Msg("application error")
}

func respondError(w http.ResponseWriter, status int, code, message string, details interface{}) {
	httpResponse.ErrorResponse(w, status, code, message, details)
}
