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
	im := imager.ImageMeta{
		ID:        "test_id",
		Author:    "test_author",
		WEBSource: "test_websource",
		MIMEType:  "test_mimetype",
		Category:  uuid.New().String(),
	}
	pagination := repository.Pagination{
		Limit:  1000,
		Offset: 0,
	}

	var imageIDRepo repository.ImageMetaRepository = imredis.NewImageMetaRepository(kvp)
	err := imageIDRepo.Insert(ctx, im)
	test.AssertErrNil(t, err)

	gotImageMetaList, err := imageIDRepo.List(ctx, im.Category, pagination)
	test.AssertErrNil(t, err)
	if len(gotImageMetaList) != 1 {
		t.Fatal(len(gotImageMetaList))
	}

	gotImageMeta := gotImageMetaList[0]
	if !reflect.DeepEqual(im, gotImageMeta.ImageMeta) {
		t.Fatal("exp", im, "got", gotImageMeta.ImageMeta)
	}

	categories, err := imageIDRepo.Categories(ctx)
	test.AssertErrNil(t, err)
	if !strings.Contains(strings.Join(categories, ";"), im.Category) {
		t.Fatal(im.Category, "not in", categories)
	}

	err = imageIDRepo.Delete(ctx, im.Category, gotImageMeta.Index)
	test.AssertErrNil(t, err)

	gotImageMetaList, err = imageIDRepo.List(ctx, im.Category, pagination)
	test.AssertErrNil(t, err)
	if len(gotImageMetaList) != 0 {
		t.Fatal(len(gotImageMetaList))
	}

	categories, err = imageIDRepo.Categories(ctx)
	test.AssertErrNil(t, err)
	if strings.Contains(strings.Join(categories, ";"), im.Category) {
		t.Fatal(im.Category, "in", categories)
	}
}
