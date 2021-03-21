package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/ocmoxa/SwapTile-Imager/internal/app"
)

func main() {
	const noMigrate = -1

	configFile := flag.String("config", "", "path to config")
	migrateVersion := flag.Int("migrate-version", noMigrate, "id of migration to run")
	flag.Parse()

	ctx := context.Background()

	if *migrateVersion != noMigrate {
		app.Migrate(ctx, *configFile, *migrateVersion)

		return
	}

	ctx, cancel := context.WithCancel(ctx)
	done := app.Start(ctx, *configFile)

	go func() {
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
	}()

	<-done
}
