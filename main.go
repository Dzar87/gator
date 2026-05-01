// Package main is the entry point for the gator CLI.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Dzar87/gator/internal/cli"
	"github.com/Dzar87/gator/internal/config"
	"github.com/Dzar87/gator/internal/database"
	"github.com/Dzar87/gator/internal/state"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	if err := run(logger); err != nil {
		logger.Error("Fatal error", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	if len(os.Args) < 2 {
		return errors.New("usage: gator <command> [args...]")
	}
	cfg, err := config.Read()
	if err != nil {
		return err
	}
	s := state.State{
		Cfg:    cfg,
		Logger: logger,
	}
	c := cli.NewCommands()
	cmd := cli.Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	ctx, stop := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM,
	)
	defer stop()

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	dbQueries := database.New(db)
	s.DB = dbQueries
	return c.Run(ctx, &s, cmd)
}
