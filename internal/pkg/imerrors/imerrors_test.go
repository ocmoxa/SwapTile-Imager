package imerrors_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"
)

func TestError(t *testing.T) {
	const expMessage = "test error"
	const err imerrors.Error = expMessage

	if err.Error() != expMessage {
		t.Fatal("exp", expMessage, "got", err.Error())
	}
}

func TestWrappedError(t *testing.T) {
	const expErrUnwrap = imerrors.Error("test error")
	err := imerrors.WrappedError{
		Err: expErrUnwrap,
	}

	switch {
	case err.Error() != expErrUnwrap.Error():
		t.Fatal("exp", expErrUnwrap.Error(), "got", err.Error())
	case errors.Unwrap(err) != expErrUnwrap:
		t.Fatal("exp", expErrUnwrap, "got", errors.Unwrap(err))
	}
}

func TestSubWrappedErrors(t *testing.T) {
	const errOriginal = imerrors.Error("test error")

	testCases := []struct {
		Err error
		Exp interface{}
	}{{
		Err: imerrors.NewUnprocessableEntity(errOriginal),
		Exp: &imerrors.UnprocessableEntity{},
	}, {
		Err: imerrors.NewMediaTypeError(errOriginal),
		Exp: &imerrors.MediaTypeError{},
	}, {
		Err: imerrors.NewNotFoundError(errOriginal),
		Exp: &imerrors.NotFoundError{},
	}, {
		Err: imerrors.NewOversizeError(errOriginal),
		Exp: &imerrors.OversizeError{},
	}, {
		Err: imerrors.NewConflictError(errOriginal),
		Exp: &imerrors.ConflictError{},
	}, {
		Err: imerrors.NewBadRequestError(errOriginal),
		Exp: &imerrors.BadRequestError{},
	}}

	for _, tc := range testCases {
		tc := tc
		t.Run(reflect.TypeOf(tc.Err).Name(), func(t *testing.T) {
			switch {
			case tc.Err.Error() != errOriginal.Error():
				t.Fatal("exp", tc.Err.Error(), "got", errOriginal.Error())
			case errors.Unwrap(tc.Err) != errOriginal:
				t.Fatal("exp", errOriginal, "got", errors.Unwrap(tc.Err))
			case !errors.As(tc.Err, tc.Exp):
				t.Fatal("exp", tc.Exp)
			}
		})
	}
}

func TestErrorPair(t *testing.T) {
	const errFirst imerrors.Error = "first"
	const errSecond imerrors.Error = "second"

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

func TestIsTemporaryError(t *testing.T) {
	const errPermanent imerrors.Error = "permanent test"

	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(-1),
	)
	defer cancel()

	switch {
	case imerrors.IsTemporaryError(nil):
		t.Fatal()
	case !imerrors.IsTemporaryError(ctx.Err()):
		t.Fatal(ctx.Err())
	case imerrors.IsTemporaryError(errPermanent):
		t.Fatal(ctx.Err())
	}
}
