package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/config"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager/improto"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imerrors"

	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
)

// Migrate runs database migration by version.
func Migrate(ctx context.Context, configFile string, version int) {
	l := zerolog.New(os.Stdout)

	cfg, err := config.Load(configFile)
	if err != nil {
		l.Fatal().Err(err).Msg("loading config")
	}

	kvp := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(cfg.Redis.Endpoint)
		},
	}

	err = applyMigration(kvp, version)
	if err != nil {
		l.Fatal().Err(err).Msg("applying migration")
	}

	l.Info().Msgf("migration #%d applied", version)
}

const errUnknownVersion imerrors.Error = "unknown version"

func applyMigration(kvp *redis.Pool, version int) (err error) {
	migrations := getMigrations()
	if version < 0 || version >= len(migrations) {
		return errUnknownVersion
	}

	kv := kvp.Get()
	defer func() { err = imerrors.ErrorPair(err, kv.Close()) }()

	err = migrations[version](kv)
	if err != nil {
		return fmt.Errorf("applying migration: %d: %w", version, err)
	}

	return nil
}

type migration func(kv redis.Conn) (err error)

func getMigrations() [1]migration {
	return [...]migration{
		migrateV1ProtoBufToJSON,
	}
}

func migrateV1ProtoBufToJSON(kv redis.Conn) (err error) {
	const keyImageIDs = "ocmoxa:image_id"
	const keyImageMeta = "ocmoxa:image_meta"
	const categoryAll = "all"

	categories, err := redis.Strings(kv.Do("KEYS", keyImageMeta+":*"))
	if err != nil {
		return fmt.Errorf("getting categories: %w", err)
	}

	im := new(improto.ImageMeta)
	for _, category := range categories {
		category = strings.TrimPrefix(category, keyImageMeta+":")

		if category == categoryAll {
			continue
		}

		imRange, err := redis.ByteSlices(kv.Do(
			"LRANGE",
			"ocmoxa:image_meta:"+category,
			0,  // Start.
			-1, // Stop.
		))
		if err != nil {
			return fmt.Errorf("doing lrange: %w", err)
		}

		for _, imBytes := range imRange {
			err := proto.Unmarshal(imBytes, im)
			if err != nil {
				return fmt.Errorf("decoding image meta: %w", err)
			}

			imID := im.GetId()
			imMap := map[string]interface{}{
				"id":       imID,
				"author":   im.GetAuthor(),
				"source":   im.GetWebSource(),
				"mimetype": im.GetMimeType(),
				"category": category,
			}
			imBytes, err := json.Marshal(&imMap)
			if err != nil {
				return fmt.Errorf("encoding image meta: %w", err)
			}

			_, err = kv.Do("HSET", keyImageMeta, imID, imBytes)
			if err != nil {
				return fmt.Errorf("saving image meta: %w", err)
			}

			_, err = kv.Do("RPUSH", keyImageIDs+":"+category, imID)
			if err != nil {
				return fmt.Errorf("saving image meta: %w", err)
			}

			_, err = kv.Do("RPUSH", keyImageIDs+":"+categoryAll, imID)
			if err != nil {
				return fmt.Errorf("saving image meta: %w", err)
			}
		}
	}

	return nil
}
