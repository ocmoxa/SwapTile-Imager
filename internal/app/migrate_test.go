package app

import (
	"errors"
	"testing"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/test"

	"github.com/gomodule/redigo/redis"
)

func TestMigration_applyMigration_invalid(t *testing.T) {
	kvp := test.InitKVP(t)
	t.Cleanup(func() { test.DisposeKVP(t, kvp) })

	t.Run("negative", func(t *testing.T) {
		err := applyMigration(kvp, -1)
		if !errors.Is(err, errUnknownVersion) {
			t.Fatal(err)
		}
	})

	t.Run("oversized", func(t *testing.T) {
		err := applyMigration(kvp, len(getMigrations()))
		if !errors.Is(err, errUnknownVersion) {
			t.Fatal(err)
		}
	})
}

func TestMigration_applyMigration_0(t *testing.T) {
	const imageID = "unsplash_SVwOposMxHY"

	kvp := test.InitKVP(t)
	t.Cleanup(func() { test.DisposeKVP(t, kvp) })

	kv := kvp.Get()
	t.Cleanup(func() { test.AssertErrNil(t, kv.Close()) })

	_, err := kv.Do(
		"LPUSH",
		"ocmoxa:image_meta:color",
		"\x0A\x14"+imageID+"\x12\x14Celine Sayuri Tagami\x1A'https://unsplash.com/photos/E48Dz9_54DY\"\x0Aimage/jpeg",
	)
	test.AssertErrNil(t, err)

	err = applyMigration(kvp, 0)
	test.AssertErrNil(t, err)

	_, err = kv.Do("HGET", "ocmoxa:image_meta", imageID)
	test.AssertErrNil(t, err)

	const expRemoveCount = 1
	var removed int
	removed, err = redis.Int(kv.Do("LREM", "ocmoxa:image_id:all", expRemoveCount, imageID))
	test.AssertErrNil(t, err)

	if removed != expRemoveCount {
		t.Fatal("got", removed, "exp", expRemoveCount)
	}

	removed, err = redis.Int(kv.Do("LREM", "ocmoxa:image_id:color", expRemoveCount, imageID))
	test.AssertErrNil(t, err)

	if removed != expRemoveCount {
		t.Fatal("got", removed, "exp", expRemoveCount)
	}
}
