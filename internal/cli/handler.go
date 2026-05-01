package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Dzar87/gator/internal/database"
	"github.com/Dzar87/gator/internal/rss"
	"github.com/Dzar87/gator/internal/state"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user does not exist")
	ErrBadArgs      = errors.New("bad command arguments")
)

func classifyLoginErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrUserNotFound
	}
	return fmt.Errorf("login: getting user: %w", err)
}

func handlerLogin(ctx context.Context, s *state.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("login: %w", ErrBadArgs)
	}
	if _, err := s.DB.GetUser(ctx, cmd.Args[0]); err != nil {
		return classifyLoginErr(err)
	}
	if err := s.Cfg.SetUser(cmd.Args[0]); err != nil {
		return fmt.Errorf("login: setting user: %w", err)
	}
	s.Logger.Info("user set", "username", cmd.Args[0])
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

func handlerRegister(ctx context.Context, s *state.State, cmd Command) error {
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
	user, err := s.DB.CreateUser(ctx, userParams)
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

func handlerReset(ctx context.Context, s *state.State, cmd Command) error {
	if err := s.DB.DeleteUsers(ctx); err != nil {
		return fmt.Errorf("reset: delete users: %w", err)
	}
	s.Logger.Info("users reset")
	return nil
}

func handlerUsers(ctx context.Context, s *state.State, cmd Command) error {
	users, err := s.DB.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("users: couldn't list users: %w", err)
	}
	for _, user := range users {
		if s.Cfg.CurrentUserName == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
			continue
		}
		fmt.Println("*", user.Name)
	}
	return nil
}

func handlerAggregate(ctx context.Context, s *state.State, cmd Command) error {
	url := "https://www.wagslane.dev/index.xml"
	feed, err := rss.FetchFeed(ctx, url)
	if err != nil {
		return fmt.Errorf("agg: failed to fetch feed: %w", err)
	}
	fmt.Printf("Feed: %+v\n", feed)
	return nil
}
