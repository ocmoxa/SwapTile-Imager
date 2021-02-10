package imerrors

import (
	"errors"
)

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
