package errors

import (
	stderrors "errors"
	"sync"
)

// mapping binds one sentinel error to its user-facing HTTP representation.
// Matching uses errors.Is, so wrapped sentinels resolve too.
type mapping struct {
	target  error
	status  int
	code    string
	message string
	details interface{}
}

var registry = struct {
	sync.RWMutex
	mappings []mapping
}{}

// RegisterError binds a sentinel error to a user-friendly HTTP response.
// Each package registers the sentinels it owns (typically in init), so the
// error handler stays generic and never imports feature packages.
// Message and details are user-facing: never put internals here.
func RegisterError(target error, status int, code, message string, details interface{}) {
	registry.Lock()
	defer registry.Unlock()
	registry.mappings = append(registry.mappings, mapping{
		target:  target,
		status:  status,
		code:    code,
		message: message,
		details: details,
	})
}

// FromError translates a known sentinel error into its registered AppError.
// It returns nil for unknown errors so callers fall back to a generic
// response that never leaks internals.
func FromError(err error) *AppError {
	registry.RLock()
	defer registry.RUnlock()
	for _, m := range registry.mappings {
		if m.target != nil && stderrors.Is(err, m.target) {
			return New(m.status, m.code, m.message, m.details)
		}
	}
	return nil
}
