package pkgs

type ApiError struct {
	Code    int
	Message string
}

func NewApiError(code int, message string) *ApiError {
	return &ApiError{
		Code:    code,
		Message: message,
	}
}

func (e *ApiError) Error() string {
	return e.Message
}
