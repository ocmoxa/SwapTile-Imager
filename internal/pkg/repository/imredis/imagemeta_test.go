// +build integration

package imredis_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository/imredis"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"

	"github.com/google/uuid"
)

func TestImageMetaRepository(t *testing.T) {
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

	var imageMetaRepo repository.ImageMetaRepository = imredis.NewImageMetaRepository(kvp)
	err := imageMetaRepo.Insert(ctx, im)
	test.AssertErrNil(t, err)

	gotImageMetaList, err := imageMetaRepo.List(ctx, im.Category, pagination)
	test.AssertErrNil(t, err)
	mustExistsImageMeta(t, gotImageMetaList, im.ID)

	gotImageMetaList, err = imageMetaRepo.List(ctx, imredis.CategoryNameAll, pagination)
	test.AssertErrNil(t, err)
	mustExistsImageMeta(t, gotImageMetaList, im.ID)

	categories, err := imageMetaRepo.Categories(ctx)
	test.AssertErrNil(t, err)
	if !strings.Contains(strings.Join(categories, ";"), im.Category) {
		t.Fatal(im.Category, "not in", categories)
	}

	err = imageMetaRepo.Delete(ctx, im.Category, 0)
	test.AssertErrNil(t, err)

	gotImageMetaList, err = imageMetaRepo.List(ctx, im.Category, pagination)
	test.AssertErrNil(t, err)
	if len(gotImageMetaList) != 0 {
		t.Fatal(len(gotImageMetaList))
	}

	categories, err = imageMetaRepo.Categories(ctx)
	test.AssertErrNil(t, err)
	if strings.Contains(strings.Join(categories, ";"), im.Category) {
		t.Fatal(im.Category, "in", categories)
	}
}

func mustExistsImageMeta(
	t *testing.T,
	imageMetaList []repository.IndexedImageMeta,
	id string,
) {
	t.Helper()

	var found bool
	for _, im := range imageMetaList {
		if im.ID == id {
			found = true

			break
		}
	}

	if !found {
		t.Fatal(id, "not in", imageMetaList)
	}
}
