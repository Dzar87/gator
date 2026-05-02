package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Dzar87/gator/internal/database"
	"github.com/Dzar87/gator/internal/state"
)

func handlerUnfollow(
	ctx context.Context, s *state.State, cmd Command, user database.User,
) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <feed_url>", cmd.Name)
	}
	feed, err := s.Queries.GetFeedByURL(ctx, cmd.Args[0])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrFeedNotFound
		}
		return err
	}
	qParams := database.DeleteFeedFollowByUserIDFeedIDParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	if _, err := s.Queries.DeleteFeedFollowByUserIDFeedID(ctx, qParams); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("provided feed '%s' is not followed", feed.Url)
		}
		return fmt.Errorf("unfollow: %w", err)
	}
	fmt.Printf("Stopped following: %s\n", feed.Name)
	return nil
}
