package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Dzar87/gator/internal/state"
)

var ErrUserNotFound = errors.New("user does not exist")

func classifyLoginErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrUserNotFound
	}
	return fmt.Errorf("login: getting user: %w", err)
}

func handlerLogin(ctx context.Context, s *state.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <user_name>", cmd.Name)
	}
	if _, err := s.Queries.GetUserByName(ctx, cmd.Args[0]); err != nil {
		return classifyLoginErr(err)
	}
	if err := s.Cfg.SetUser(cmd.Args[0]); err != nil {
		return fmt.Errorf("login: setting user: %w", err)
	}
	s.Logger.Info("user set", "username", cmd.Args[0])
	return nil
}
