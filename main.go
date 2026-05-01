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
	"time"

	"github.com/Dzar87/gator/internal/config"
	"github.com/Dzar87/gator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type state struct {
	cfg    *config.Config
	db     *database.Queries
	logger *slog.Logger
}

type command struct {
	Name string
	Args []string
}

type commands struct {
	handlers map[string]func(context.Context, *state, command) error
}

var (
    ErrUserExists      = errors.New("user already exists")
    ErrUserNotFound    = errors.New("user does not exist")
    ErrBadArgs         = errors.New("bad command arguments")
)

func classifyLoginErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrUserNotFound
	}
	return fmt.Errorf("login: getting user: %w", err)
}

func handlerLogin(ctx context.Context, s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("login: %w", ErrBadArgs)
	}
	if _, err := s.db.GetUser(ctx, cmd.Args[0]); err != nil {
		return classifyLoginErr(err)
	}
	if err := s.cfg.SetUser(cmd.Args[0]); err != nil {
		return fmt.Errorf("login: setting user: %w", err)
	}
	s.logger.Info("user set", "username", cmd.Args[0])
	return nil
}

func classifyRegisterErr(err error) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return fmt.Errorf("register: creating user: %w", err)
	}
	if pqErr.Code.Name() == "unique_violation" && pqErr.Constraint == "users_name_key" {
		return ErrUserExists
	}
	return fmt.Errorf("register: db error: %w", err)
}

func handlerRegister(ctx context.Context, s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("register: %w", ErrBadArgs)
	}
	now := time.Now().UTC()
	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      cmd.Args[0],
	}
	user, err := s.db.CreateUser(ctx, userParams)
	if err != nil {
		return classifyRegisterErr(err)
	}
	s.logger.Info("register", "name", cmd.Args[0])
	s.logger.Debug("register db", "user", user)
	if err := s.cfg.SetUser(cmd.Args[0]); err != nil {
		return fmt.Errorf("register: setting user: %w", err)
	}
	s.logger.Info("user set", "username", cmd.Args[0])
	return nil
}

func handlerReset(ctx context.Context, s *state, cmd command) error {
	if err := s.db.DeleteUsers(ctx); err != nil {
		return fmt.Errorf("reset: delete users: %w", err)
	}
	s.logger.Info("users reset")
	return nil
}

func (c *commands) run(ctx context.Context, s *state, cmd command) error {
	f, ok := c.handlers[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return f(ctx, s, cmd)
}

func (c *commands) register(
	name string, f func(context.Context, *state, command) error,
) {
	c.handlers[name] = f
}

func newCommands() *commands {
	c := &commands{
		handlers: make(map[string]func(context.Context, *state, command) error),
	}
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	return c
}

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
	s := state{
		cfg:    cfg,
		logger: logger,
	}
	c := newCommands()
	cmd := command{
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
	s.db = dbQueries
	return c.run(ctx, &s, cmd)
}
