package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dzar87/gator/internal/state"
	"github.com/lib/pq"
)

var ErrUserExists = errors.New("user already exists")

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

func handlerRegister(ctx context.Context, s *state.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <user_name>", cmd.Name)
	}
	user, err := s.Queries.CreateUser(ctx, cmd.Args[0])
	if err != nil {
		return classifyRegisterErr(err)
	}
	s.Logger.Info("register", "name", cmd.Args[0])
	s.Logger.Debug("register db", "user", user)
	if err := s.Cfg.SetUser(cmd.Args[0]); err != nil {
		return fmt.Errorf("register: setting user: %w", err)
	}
	s.Logger.Info("user set", "username", cmd.Args[0])
	return nil
}
