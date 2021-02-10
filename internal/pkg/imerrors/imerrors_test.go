package imerrors_test

import (
	"errors"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"
)

func TestErrorPair(t *testing.T) {
	errFirst := errors.New("first")
	errSecond := errors.New("second")

	t.Run("pair", func(t *testing.T) {
		err := imerrors.ErrorPair(errFirst, errSecond)
		switch {
		case !errors.Is(err, errFirst):
			t.Fatal(err, "is not", errFirst)
		case !errors.Is(err, errSecond):
			t.Fatal(err, "is not", errSecond)
		case errors.Unwrap(err) != errFirst:
			t.Fatal(errors.Unwrap(err), "is not", errFirst)
		}
	})

	t.Run("first", func(t *testing.T) {
		err := imerrors.ErrorPair(errFirst, nil)
		if err != errFirst {
			t.Fatal(err, "is not", errFirst)
		}
	})

	t.Run("second", func(t *testing.T) {
		err := imerrors.ErrorPair(nil, errSecond)
		if err != errSecond {
			t.Fatal(err, "is not", errSecond)
		}
	})

	t.Run("nil", func(t *testing.T) {
		err := imerrors.ErrorPair(nil, nil)
		if err != nil {
			t.Fatal(err, "is not nil")
		}
	})
}
