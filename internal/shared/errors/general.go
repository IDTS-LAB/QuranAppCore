package errors

type AppError struct {
	Code    string
	Message string
	Status  int
	Details interface{}
}

func (e *AppError) Error() string {
	return e.Message
}

func New(
	status int,
	code string,
	message string,
	details interface{},
) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
		Details: details,
	}
}
