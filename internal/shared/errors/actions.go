package errors

// Action hints tell clients what to do next. Every error response must carry
// a non-empty details payload built from these, so clients can react
// programmatically instead of parsing human-readable messages.
const (
	// ActionLogin asks the client to show the login form and authenticate.
	ActionLogin = "login"
	// ActionRefreshToken asks the client to try the refresh-token flow
	// before forcing a full login.
	ActionRefreshToken = "refresh_token"
	// ActionVerifyEmail asks the client to direct the user to email verification.
	ActionVerifyEmail = "verify_email"
	// ActionFixRequest asks the client to correct the request payload.
	ActionFixRequest = "fix_request"
	// ActionRetryLater asks the client to retry the same request later.
	ActionRetryLater = "retry_later"
)

// ActionDetails builds the details payload for an error response.
// It is never empty, so error.details is always present for clients.
func ActionDetails(action string) map[string]string {
	return map[string]string{"action": action}
}
