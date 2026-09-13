package http

// RegisterError mirrors the error object inside the shared response envelope.
type RegisterError struct {
	Code    string `json:"code" example:"EMAIL_TAKEN"`
	Details any    `json:"details,omitempty"`
}

// RegisterResponse mirrors the shared response envelope for API docs.
// (Declared locally because swag cannot resolve the aliased shared package.)
type RegisterResponse struct {
	Success bool           `json:"success" example:"true"`
	Message string         `json:"message" example:"user register successful"`
	Data    any            `json:"data,omitempty"`
	Error   *RegisterError `json:"error,omitempty"`
}
