// +build integration

package imredis_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository/imredis"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"

	"github.com/google/uuid"
)

func TestCategoryRepository(t *testing.T) {
	t.Parallel()

	kvp := test.InitKVP(t)
	defer test.DisposeKVP(t, kvp)

	ctx := context.Background()
	testName := uuid.New().String()

	categoryRepo := imredis.NewCategoryRepository(kvp)
	err := categoryRepo.Upsert(ctx, testName)
	test.AssertErrNil(t, err)

	names, err := categoryRepo.List(ctx)
	test.AssertErrNil(t, err)
	if !strings.Contains(strings.Join(names, ";"), testName) {
		t.Fatal(testName, "not in", names)
	}

	err = categoryRepo.Delete(ctx, testName)
	test.AssertErrNil(t, err)

	names, err = categoryRepo.List(ctx)
	test.AssertErrNil(t, err)
	if strings.Contains(strings.Join(names, ";"), testName) {
		t.Fatal(testName, "is in", names)
	}
}
