package http

import (
	"encoding/json"
	"net/http"

	"github.com/fiqrikm18/quran-app/internal/shared"
	"github.com/fiqrikm18/quran-app/internal/shared/errors"
)

func Decode[T any](r *http.Request, destination *T) error {
	if err := json.NewDecoder(r.Body).Decode(destination); err != nil {
		return errors.BadRequest(
			"We couldn't understand your request. Please check the format and try again.",
			errors.ActionDetails(errors.ActionFixRequest),
		)
	}

	if details := shared.ValidateWithDetails(destination); details != nil {
		return errors.BadRequest(
			"Please fix the highlighted fields and try again.",
			details,
		)
	}

	return nil
}
