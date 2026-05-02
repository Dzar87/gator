package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Dzar87/gator/internal/database"
	"github.com/Dzar87/gator/internal/state"
	"github.com/lib/pq"
)

var (
	ErrFeedFollowExists = errors.New("feed follow already exists")
	ErrFeedNotFound     = errors.New("feed does not exist")
)

func classifyFollowFeedErr(err error) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return fmt.Errorf("followfeed: creating feed follow: %w", err)
	}
	if pqErr.Code.Name() == "unique_violation" &&
		pqErr.Constraint == "uniq_user_id_feed_id" {
		return ErrFeedFollowExists
	}
	return fmt.Errorf("followfeed: db error: %w", err)
}

func handlerFollowFeed(
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
	feedFollowParams := database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	if _, err := s.Queries.CreateFeedFollow(ctx, feedFollowParams); err != nil {
		return classifyFollowFeedErr(err)
	}
	fmt.Println("Now following:", feed.Name)
	return nil
}
