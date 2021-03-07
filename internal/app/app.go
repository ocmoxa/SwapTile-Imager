package app

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/api/imhttp"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/config"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/imager/core"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/repository/imredis"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/storage/s3"
	"github.com/ocmoxa/SwapTile-Imager/internal/pkg/validate"

	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog"
)

func Start(ctx context.Context, configFile string) (done chan struct{}) {
	done = make(chan struct{})

	go func() {
		defer close(done)

		l := zerolog.New(os.Stdout)

		err := runApp(ctx, l, configFile)
		if err != nil {
			l.Fatal().Err(err).Msg("running app")
		}
	}()

	return done
}

func runApp(ctx context.Context, l zerolog.Logger, configFile string) (err error) {
	rand.Seed(time.Now().Unix())

	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	lvl, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		l.Fatal().Err(err).Msg("parsing level")

		return
	}

	l = l.Level(lvl)

	l.Debug().Interface("config", cfg).Msg("loaded config")

	kvp := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(cfg.Redis.Endpoint)
		},
	}

	repoImageMeta := imredis.NewImageMetaRepository(kvp)
	repoImageID := imredis.NewImageIDRepository(kvp)

	fileStorage, err := s3.NewS3Storage(cfg.S3)
	if err != nil {
		return fmt.Errorf("initializing file storage: %w", err)
	}

	validate := validate.New()

	c := core.NewCore(core.Essentials{
		ImageMetaRepository: repoImageMeta,
		ImageIDRepository:   repoImageID,
		FileStorage:         fileStorage,
		Validate:            validate,
	}, cfg.Core)

	srv, err := imhttp.NewServer(imhttp.Essentials{
		Logger: l,
		Core:   c,
	}, cfg.Server)
	if err != nil {
		return fmt.Errorf("creating server: %w", err)
	}

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Duration(cfg.Server.ShutdownTimeout),
		)
		defer cancel()

		serr := srv.Shutdown(ctx)
		if serr != nil {
			l.Warn().Err(serr).Msg("shutdowing server")
		}
	}()

	l.Info().Msgf("starting server on %s", cfg.Server.Address)

	err = srv.ListenAndServe()
	if err != nil {
		return fmt.Errorf("lister and server: %w", err)
	}

	return nil
}
