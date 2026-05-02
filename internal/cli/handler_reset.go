package cli

import (
	"context"
	"fmt"

	"github.com/Dzar87/gator/internal/state"
)

func handlerReset(ctx context.Context, s *state.State, cmd Command) error {
	if err := s.Queries.DeleteUsers(ctx); err != nil {
		return fmt.Errorf("reset: delete users: %w", err)
	}
	s.Logger.Info("users reset")
	return nil
}
