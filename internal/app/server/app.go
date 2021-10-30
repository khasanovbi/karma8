package server

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"
)

type Application struct {
	server          *http.Server
	shutdownTimeout time.Duration
	logger          *zap.Logger
}

func (m *Application) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		m.logger.Info("server start listening", zap.String("addr", m.server.Addr))
		if err := m.server.ListenAndServe(); err != nil {
			return fmt.Errorf("listen and server error: %w", err)
		}
		return nil
	})

	group.Go(func() error {
		<-ctx.Done()

		m.logger.Info("graceful shutdown of server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), m.shutdownTimeout)
		defer cancel()
		if err := m.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}

		return ctx.Err()
	})

	return group.Wait()
}

func NewApplication(conf *Config, logger *zap.Logger) (*Application, error) {
	balancer := newBalancer(&conf.Balancer)
	pg, err := newPG(&conf.PG, logger)
	if err != nil {
		return nil, err
	}

	fileMetaStorage := newFileMetaStorage(pg, logger)

	fileService := newFileService(balancer, fileMetaStorage, conf.MinChunkSize, logger)

	return &Application{
		server:          newHTTPServer(&conf.HTTP, fileService, logger),
		shutdownTimeout: conf.ShutdownTimeout,
		logger:          logger,
	}, nil
}
