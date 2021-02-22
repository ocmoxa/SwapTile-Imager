package imerrors

import (
	"errors"
)

// Error is a constant-like error.
type Error string

func (err Error) Error() string {
	return string(err)
}

// WrappedError wraps error for adding meaning to it.
type WrappedError struct {
	Err error
}

func (err WrappedError) Unwrap() error {
	return err.Err
}

func (err WrappedError) Error() string {
	return err.Err.Error()
}

// OversizeError means that given err is about something too large.
type OversizeError struct {
	WrappedError
}

// NewOversizeError wraps err and creates OversizeError.
func NewOversizeError(err error) error {
	return OversizeError{
		WrappedError: WrappedError{
			Err: err,
		},
	}
}

// NotFoundError means that given err is about something not found.
type NotFoundError struct {
	WrappedError
}

// NewNotFoundError wraps err and creates NotFoundError.
func NewNotFoundError(err error) error {
	return NotFoundError{
		WrappedError: WrappedError{
			Err: err,
		},
	}
}

// UserError means that given err is about bad request or invalid
// argument.
type UserError struct {
	WrappedError
}

// NewUserError wraps err and creates UserError.
func NewUserError(err error) error {
	return UserError{
		WrappedError: WrappedError{
			Err: err,
		},
	}
}

// MediaTypeError means that given err is about invalid content type.
type MediaTypeError struct {
	WrappedError
}

// NewMediaTypeError wraps err and creates MediaTypeError.
func NewMediaTypeError(err error) error {
	return MediaTypeError{
		WrappedError: WrappedError{
			Err: err,
		},
	}
}

// ConflictError means that given err is about entity conflict.
type ConflictError struct {
	WrappedError
}

// NewConflictError wraps err and creates ConflictError.
func NewConflictError(err error) error {
	return ConflictError{
		WrappedError: WrappedError{
			Err: err,
		},
	}
}

type errorPair struct {
	main      error
	secondary error
}

func (err errorPair) Unwrap() error {
	return err.main
}

func (err errorPair) Error() string {
	return err.main.Error() + "; " + err.secondary.Error()
}

func (err errorPair) Is(original error) bool {
	return errors.Is(original, err.main) || errors.Is(original, err.secondary)
}

// ErrorPair handles deferred errors. If all error are not nil, it
// creates a new error pair that holds information about both errors.
// If one of the errors is nil, non-nil error will be returned. If bot
//  errors are nil, nil will be returned.
func ErrorPair(main, secondary error) error {
	switch {
	case main == nil:
		return secondary
	case secondary == nil:
		return main
	default:
		return &errorPair{
			main:      main,
			secondary: secondary,
		}
	}
}

type temporaryError interface {
	Temporary() bool
}

// IsTemporaryError checks that error has Temporary method and it
// returns true.
func IsTemporaryError(err error) bool {
	var errTmp temporaryError

	if errors.As(err, &errTmp) {
		return errTmp.Temporary()
	}

	return false
}
