// Package main is the entry point for the gator CLI.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/Dzar87/gator/internal/config"
)

type state struct {
	cfg    *config.Config
	logger *slog.Logger
}

type command struct {
	Name string
	Args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return errors.New("login: bad command arguments")
	}
	if err := s.cfg.SetUser(cmd.Args[0]); err != nil {
		return fmt.Errorf("login: setting user: %w", err)
	}
	s.logger.Info("user set", "username", cmd.Args[0])
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	f, ok := c.handlers[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
	return f(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func newCommands() *commands {
	c := &commands{
		handlers: make(map[string]func(*state, command) error),
	}
	c.register("login", handlerLogin)
	return c
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
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
	return c.run(&s, cmd)
}
