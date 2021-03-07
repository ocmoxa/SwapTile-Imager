// +build integration

package imredis_test

import (
	"context"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository/imredis"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"

	"github.com/google/uuid"
)

func TestImageMetaRepository(t *testing.T) {
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

func TestImageMetaRepository_Shuffle(t *testing.T) {
	const count = 10
	const depth = 100
	category := uuid.New().String()

	kvp := test.InitKVP(t)
	defer test.DisposeKVP(t, kvp)

	ctx := context.Background()
	pagination := repository.Pagination{
		Limit:  1000,
		Offset: 0,
	}

	var imageMetaRepo repository.ImageMetaRepository = imredis.NewImageMetaRepository(kvp)
	for i := 0; i < count; i++ {
		err := imageMetaRepo.Insert(ctx, imager.ImageMeta{
			ID:        "test_id_" + strconv.Itoa(i),
			Author:    "test_author",
			WEBSource: "test_websource",
			MIMEType:  "test_mimetype",
			Category:  category,
		})
		test.AssertErrNil(t, err)
	}

	initialMeta, err := imageMetaRepo.List(ctx, category, pagination)
	test.AssertErrNil(t, err)
	if len(initialMeta) == 0 {
		t.Fatal(len(initialMeta))
	}

	rand.Seed(1)

	err = imageMetaRepo.Shuffle(ctx, category, depth)
	test.AssertErrNil(t, err)

	shuffledMeta, err := imageMetaRepo.List(ctx, category, pagination)
	test.AssertErrNil(t, err)
	if len(shuffledMeta) != len(initialMeta) {
		t.Fatal("exp", len(initialMeta), "got", len(shuffledMeta))
	}

	var equalCount int
	for i, im := range initialMeta {
		if im.ID == shuffledMeta[i].ID {
			equalCount++
		}
	}

	t.Log("coincided", equalCount, "elements out of", len(initialMeta))

	if equalCount == len(initialMeta) {
		t.Fatal()
	}
}

func TestImageMetaRepository_Shuffle_empty(t *testing.T) {
	const depth = 100
	category := uuid.New().String()

	kvp := test.InitKVP(t)
	defer test.DisposeKVP(t, kvp)

	ctx := context.Background()
	var imageMetaRepo repository.ImageMetaRepository = imredis.NewImageMetaRepository(kvp)

	rand.Seed(1)

	t.Run("empty_list", func(t *testing.T) {
		err := imageMetaRepo.Shuffle(ctx, category, depth)
		test.AssertErrNil(t, err)
	})

	t.Run("zero_depth", func(t *testing.T) {
		err := imageMetaRepo.Shuffle(ctx, category, 0)
		test.AssertErrNil(t, err)
	})
}
