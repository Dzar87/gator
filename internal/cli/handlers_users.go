package cli

import (
	"context"
	"fmt"

	"github.com/Dzar87/gator/internal/state"
)

func handlerUsers(ctx context.Context, s *state.State, cmd Command) error {
	users, err := s.Queries.GetUsers(ctx)
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
