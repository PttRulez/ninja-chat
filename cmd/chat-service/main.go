package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/pttrulez/ninja-chat/internal/config"
	"github.com/pttrulez/ninja-chat/internal/logger"
	serverdebug "github.com/pttrulez/ninja-chat/internal/server-debug"
)

var configPath = flag.String("config", "configs/config.toml", "Path to config file")

func main() {
	if err := run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}

func run() (errReturned error) {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.ParseAndValidate(*configPath)
	if err != nil {
		return fmt.Errorf("parse and validate config %q: %v", *configPath, err)
	}

	// Setup Logger
	logger.MustInit(logger.NewOptions(cfg.Log.Level, logger.WithProductionMode(cfg.Global.Env == "prod")))
	logger.Sync()

	// Setup Debug Server
	srvOptions := serverdebug.NewOptions(cfg.Servers.Debug.Addr)
	srvDebug, err := serverdebug.New(srvOptions)
	if err != nil {
		return fmt.Errorf("init debug server: %v", err)
	}

	// Start Server
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error { return srvDebug.Run(ctx) })

	// Run services.
	// Ждут своего часа.
	// ...

	if err = eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("wait app stop: %v", err)
	}

	return nil
}
