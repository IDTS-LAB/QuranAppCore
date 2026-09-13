package errors

import "net/http"

func BadRequest(message string, details interface{}) *AppError {
	return New(
		http.StatusBadRequest,
		"BAD_REQUEST",
		message,
		details,
	)
}

func Unauthorized(message string) *AppError {
	return New(
		http.StatusUnauthorized,
		"UNAUTHORIZED",
		message,
		nil,
	)
}

// UnauthorizedWithDetails is Unauthorized with a non-empty details payload.
func UnauthorizedWithDetails(message string, details interface{}) *AppError {
	return New(
		http.StatusUnauthorized,
		"UNAUTHORIZED",
		message,
		details,
	)
}

func Forbidden(message string) *AppError {
	return New(
		http.StatusForbidden,
		"FORBIDDEN",
		message,
		nil,
	)
}

func NotFound(message string) *AppError {
	return New(
		http.StatusNotFound,
		"NOT_FOUND",
		message,
		nil,
	)
}

func Internal(message string) *AppError {
	return New(
		http.StatusInternalServerError,
		"INTERNAL_ERROR",
		message,
		nil,
	)
}
