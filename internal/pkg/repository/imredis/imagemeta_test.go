// +build integration

package imredis_test

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository/imredis"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"

	"github.com/google/uuid"
)

func TestImageIDRepository(t *testing.T) {
	t.Parallel()

	kvp := test.InitKVP(t)
	defer test.DisposeKVP(t, kvp)

	ctx := context.Background()
	testImageMeta := imager.ImageMeta{
		ID:        "test_id",
		Author:    "test_author",
		WEBSource: "test_websource",
		MIMEType:  "test_mimetype",
	}
	testCategory := uuid.New().String()
	pagination := repository.Pagination{
		Limit:  1000,
		Offset: 0,
	}

	var imageIDRepo repository.ImageMetaRepository = imredis.NewImageMetaRepository(kvp)
	err := imageIDRepo.Insert(ctx, testCategory, testImageMeta)
	test.AssertErrNil(t, err)

	gotImageMetaList, err := imageIDRepo.List(ctx, testCategory, pagination)
	test.AssertErrNil(t, err)
	if len(gotImageMetaList) != 1 {
		t.Fatal(len(gotImageMetaList))
	}

	gotImageMeta := gotImageMetaList[0]
	if !reflect.DeepEqual(testImageMeta, gotImageMeta.ImageMeta) {
		t.Fatal("exp", testImageMeta, "got", gotImageMeta.ImageMeta)
	}

	categories, err := imageIDRepo.Categories(ctx)
	test.AssertErrNil(t, err)
	if !strings.Contains(strings.Join(categories, ";"), testCategory) {
		t.Fatal(testCategory, "not in", categories)
	}

	err = imageIDRepo.Delete(ctx, testCategory, gotImageMeta.Index)
	test.AssertErrNil(t, err)

	gotImageMetaList, err = imageIDRepo.List(ctx, testCategory, pagination)
	test.AssertErrNil(t, err)
	if len(gotImageMetaList) != 0 {
		t.Fatal(len(gotImageMetaList))
	}

	categories, err = imageIDRepo.Categories(ctx)
	test.AssertErrNil(t, err)
	if strings.Contains(strings.Join(categories, ";"), testCategory) {
		t.Fatal(testCategory, "in", categories)
	}
}
