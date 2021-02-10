// +build integration

package imredis_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository/imredis"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"
)

func TestImageIDRepository(t *testing.T) {
	t.Parallel()

	kvp := test.InitKVP(t)
	defer test.DisposeKVP(t, kvp)

	ctx := context.Background()
	testCategory := uuid.New().String()
	testID := uuid.New().String()
	pagination := repository.Pagination{
		Limit:  1000,
		Offset: 0,
	}

	var imageIDRepo repository.ImageIDRepository = imredis.NewImageIDRepository(kvp)
	err := imageIDRepo.Insert(ctx, testCategory, testID)
	test.AssertErrNil(t, err)

	ids, err := imageIDRepo.List(ctx, testCategory, pagination)
	test.AssertErrNil(t, err)
	if !strings.Contains(strings.Join(ids, ";"), testID) {
		t.Fatal(testID, "not in", ids)
	}

	err = imageIDRepo.Delete(ctx, testCategory, testID)
	test.AssertErrNil(t, err)

	ids, err = imageIDRepo.List(ctx, testCategory, pagination)
	test.AssertErrNil(t, err)
	if strings.Contains(strings.Join(ids, ";"), testID) {
		t.Fatal(testID, "in", ids)
	}
}
