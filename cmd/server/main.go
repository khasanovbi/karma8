package main

import (
	"context"
	"fmt"
	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
	"github.com/heetch/confita/backend/file"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"karma8/internal/app/server"
	"os"
	"os/signal"
	"syscall"
)

func makeLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

func main() {
	configPath := "./configs/config.yaml"
	loader := confita.NewLoader(env.NewBackend(), file.NewBackend(configPath))

	config := &server.Config{}
	if err := loader.Load(context.Background(), config); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "can't load config: %s\n", err)
		os.Exit(1)
	}

	logger, err := makeLogger()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "can't create logger: %s\n", err)
		os.Exit(1)
	}

	app, err := server.NewApplication(config, logger)
	if err != nil {
		logger.Error("can't create app", zap.Error(err))
		os.Exit(1)
	}

	group, ctx := errgroup.WithContext(context.Background())
	group.Go(func() error {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case s := <-sig:
			logger.Info("got signal", zap.Stringer("signal", s))
			return fmt.Errorf("signal stop: %s", s)
		}
	})
	group.Go(func() error {
		return app.Run(ctx)
	})

	if err := group.Wait(); err != nil {
		logger.Error("app error", zap.Error(err))
		os.Exit(1)
	}
}
