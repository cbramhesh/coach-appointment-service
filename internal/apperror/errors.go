package apperror

import "fmt"

type AppError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func BadRequest(code, message string) *AppError {
	return &AppError{
		StatusCode: 400,
		Code:       code,
		Message:    message,
	}
}

func NotFound(code, message string) *AppError {
	return &AppError{
		StatusCode: 404,
		Code:       code,
		Message:    message,
	}
}

func Internal(code, message string) *AppError {
	return &AppError{
		StatusCode: 500,
		Code:       code,
		Message:    message,
	}
}
