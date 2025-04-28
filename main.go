package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hionay/quotes/internal/api"
	"github.com/hionay/quotes/internal/cmdutil"
	"github.com/hionay/quotes/internal/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		cancel()
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	cfg := config.NewConfig()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	dbPool, err := cmdutil.NewMySQLPool(ctx, cfg)
	if err != nil {
		return fmt.Errorf("cmdutil.NewMySQLPool(): %w", err)
	}

	a := api.NewAPI(cfg, logger, dbPool)

	serveErrCh := make(chan error, 1)
	go func() {
		defer close(serveErrCh)
		serveErrCh <- a.ListenAndServe()
	}()

	<-ctx.Done()
	logger.Info("Shutting down the server")

	if err := a.Shutdown(); err != nil {
		logger.Error("Failed to shutdown server", slog.Any("err", err))
	}
	if err := dbPool.Close(); err != nil {
		logger.Error("Failed to close database pool", slog.Any("err", err))
	}
	return <-serveErrCh
}
