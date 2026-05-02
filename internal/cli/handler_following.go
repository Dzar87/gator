package cli

import (
	"context"
	"fmt"

	"github.com/Dzar87/gator/internal/database"
	"github.com/Dzar87/gator/internal/state"
)

func handlerFollowing(
	ctx context.Context, s *state.State, cmd Command, user database.User,
) error {
	rows, err := s.Queries.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("following: couldn't list feeds: %w", err)
	}
	if len(rows) == 0 {
		fmt.Println("No followed feeds found.")
		return nil
	}
	for _, row := range rows {
		fmt.Println("*", row.FeedName)
	}
	return nil
}
