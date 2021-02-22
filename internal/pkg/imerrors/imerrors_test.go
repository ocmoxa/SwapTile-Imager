package imerrors_test

import (
	"context"
	"errors"
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

func TestWrappedErrorDefinitions(t *testing.T) {
	const expErr = imerrors.Error("test error")

	errNotFound := imerrors.NewNotFoundError(expErr)
	errUser := imerrors.NewUserError(expErr)
	errOversize := imerrors.NewOversizeError(expErr)
	errMediaType := imerrors.NewMediaTypeError(expErr)
	errConfict := imerrors.NewConflictError(expErr)

	switch {
	// NotFoundError.
	case errNotFound.Error() != expErr.Error():
		t.Fatal("exp", expErr.Error(), "got", errNotFound.Error())
	case errors.Unwrap(errNotFound) != expErr:
		t.Fatal("exp", expErr, "got", errors.Unwrap(errNotFound))
	case !errors.As(errNotFound, &imerrors.NotFoundError{}):
		t.Fatal("exp", errNotFound)
	case errors.As(errNotFound, &imerrors.UserError{}):
		t.Fatal("exp", errNotFound)
	case errors.As(errNotFound, &imerrors.OversizeError{}):
		t.Fatal("exp", errNotFound)
	case errors.As(errNotFound, &imerrors.MediaTypeError{}):
		t.Fatal("exp", errNotFound)
	case errors.As(errNotFound, &imerrors.ConflictError{}):
		t.Fatal("exp", errNotFound)

	// UserError.
	case errUser.Error() != expErr.Error():
		t.Fatal("exp", expErr.Error(), "got", errUser.Error())
	case errors.Unwrap(errUser) != expErr:
		t.Fatal("exp", expErr, "got", errors.Unwrap(errUser))
	case !errors.As(errUser, &imerrors.UserError{}):
		t.Fatal("exp", errUser)
	case errors.As(errUser, &imerrors.NotFoundError{}):
		t.Fatal("exp", errUser)
	case errors.As(errUser, &imerrors.OversizeError{}):
		t.Fatal("exp", errUser)
	case errors.As(errUser, &imerrors.MediaTypeError{}):
		t.Fatal("exp", errUser)
	case errors.As(errUser, &imerrors.ConflictError{}):
		t.Fatal("exp", errUser)

	// OversizeError.
	case errOversize.Error() != expErr.Error():
		t.Fatal("exp", expErr.Error(), "got", errOversize.Error())
	case errors.Unwrap(errOversize) != expErr:
		t.Fatal("exp", expErr, "got", errors.Unwrap(errOversize))
	case !errors.As(errOversize, &imerrors.OversizeError{}):
		t.Fatal("exp", errOversize)
	case errors.As(errOversize, &imerrors.NotFoundError{}):
		t.Fatal("exp", errOversize)
	case errors.As(errOversize, &imerrors.UserError{}):
		t.Fatal("exp", errOversize)
	case errors.As(errOversize, &imerrors.MediaTypeError{}):
		t.Fatal("exp", errOversize)
	case errors.As(errOversize, &imerrors.ConflictError{}):
		t.Fatal("exp", errOversize)

	// MediaTypeError.
	case errMediaType.Error() != expErr.Error():
		t.Fatal("exp", expErr.Error(), "got", errMediaType.Error())
	case errors.Unwrap(errMediaType) != expErr:
		t.Fatal("exp", expErr, "got", errors.Unwrap(errMediaType))
	case !errors.As(errMediaType, &imerrors.MediaTypeError{}):
		t.Fatal("exp", errMediaType)
	case errors.As(errMediaType, &imerrors.NotFoundError{}):
		t.Fatal("exp", errMediaType)
	case errors.As(errMediaType, &imerrors.UserError{}):
		t.Fatal("exp", errMediaType)
	case errors.As(errMediaType, &imerrors.OversizeError{}):
		t.Fatal("exp", errMediaType)
	case errors.As(errMediaType, &imerrors.ConflictError{}):
		t.Fatal("exp", errMediaType)

	// ConflictError.
	case errConfict.Error() != expErr.Error():
		t.Fatal("exp", expErr.Error(), "got", errConfict.Error())
	case errors.Unwrap(errConfict) != expErr:
		t.Fatal("exp", expErr, "got", errors.Unwrap(errConfict))
	case !errors.As(errConfict, &imerrors.ConflictError{}):
		t.Fatal("exp", errConfict)
	case errors.As(errConfict, &imerrors.NotFoundError{}):
		t.Fatal("exp", errConfict)
	case errors.As(errConfict, &imerrors.UserError{}):
		t.Fatal("exp", errConfict)
	case errors.As(errConfict, &imerrors.OversizeError{}):
		t.Fatal("exp", errConfict)
	case errors.As(errConfict, &imerrors.MediaTypeError{}):
		t.Fatal("exp", errConfict)
	}

}
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
