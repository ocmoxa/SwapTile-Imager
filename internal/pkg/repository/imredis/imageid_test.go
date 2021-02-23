// +build integration

package imredis_test

import (
	"context"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository/imredis"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"

	"github.com/google/uuid"
)

func TestImageIDRepository(t *testing.T) {
	kvp := test.InitKVP(t)
	defer test.DisposeKVP(t, kvp)

	ctx := context.Background()
	id := uuid.NewString()

	var imageIDRepo repository.ImageIDRepository = imredis.NewImageIDRepository(kvp)
	ok, err := imageIDRepo.Set(ctx, id)
	test.AssertErrNil(t, err)
	if !ok {
		t.Fatal(ok)
	}

	ok, err = imageIDRepo.Set(ctx, id)
	test.AssertErrNil(t, err)
	if ok {
		t.Fatal(ok)
	}

	err = imageIDRepo.Delete(ctx, id)
	test.AssertErrNil(t, err)

	ok, err = imageIDRepo.Set(ctx, id)
	test.AssertErrNil(t, err)
	if !ok {
		t.Fatal(ok)
	}
}
