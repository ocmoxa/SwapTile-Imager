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
	configFile := flag.String("config", "", "path to config")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	done := app.Start(ctx, *configFile)

	go func() {
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
	}()

	<-done
}
