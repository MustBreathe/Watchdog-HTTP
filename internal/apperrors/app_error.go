package apperrors

import "fmt"

type Kind string

const (
	VALIDATION_ERROR Kind = "VALIDATION_ERROR"
	NOT_FOUND        Kind = "NOT_FOUND"
	UNAUTHORIZED     Kind = "UNAUTHORIZED"
	INTERNAL_ERROR   Kind = "INTERNAL_ERROR"
	CONFLICT         Kind = "CONFLICT"
)

type AppError struct {
	Kind    Kind
	Message string
	Err     error
}

func New(kind Kind, message string, err error) *AppError {
	return &AppError{
		Kind:    kind,
		Message: message,
		Err:     err,
	}
}

func (e *AppError) Error() string {

	if e == nil {
		return ""
	}

	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}
