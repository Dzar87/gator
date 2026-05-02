package cli

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Dzar87/gator/internal/database"
	"github.com/Dzar87/gator/internal/state"
)

func requireCurrentUser(ctx context.Context, s *state.State) (database.User, error) {
	user, err := s.Queries.GetUserByName(ctx, s.Cfg.CurrentUserName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return database.User{}, ErrUserNotFound
		}
		return database.User{}, err
	}
	return user, nil
}

func loggedIn(h func(
	ctx context.Context, s *state.State, cmd Command, user database.User,
) error) func(context.Context, *state.State, Command) error {
	return func(ctx context.Context, s *state.State, cmd Command) error {
		user, err := requireCurrentUser(ctx, s)
		if err != nil {
			return err
		}
		return h(ctx, s, cmd, user)
	}
}
