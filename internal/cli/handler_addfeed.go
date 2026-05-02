package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dzar87/gator/internal/database"
	"github.com/Dzar87/gator/internal/state"
	"github.com/lib/pq"
)

var ErrFeedExists = errors.New("feed already exists")

func classifyAddFeedErr(err error) error {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return fmt.Errorf("addfeed: creating feed: %w", err)
	}
	if pqErr.Code.Name() == "unique_violation" && pqErr.Constraint == "feeds_url_key" {
		return ErrFeedExists
	}
	return fmt.Errorf("addfeed: db error: %w", err)
}

func handlerAddFeed(
	ctx context.Context, s *state.State, cmd Command, user database.User,
) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s <feed_name> <feed_url>", cmd.Name)
	}
	feedParams := database.CreateFeedParams{
		Name:   cmd.Args[0],
		Url:    cmd.Args[1],
		UserID: user.ID,
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.Queries.WithTx(tx)
	feed, err := qtx.CreateFeed(ctx, feedParams)
	if err != nil {
		return classifyAddFeedErr(err)
	}
	feedFollowParams := database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	if _, err := qtx.CreateFeedFollow(ctx, feedFollowParams); err != nil {
		return classifyFollowFeedErr(err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("addfeed: commit: %w", err)
	}
	fmt.Printf("Added feed %q (%s)\n", feed.Name, feed.Url)
	fmt.Println("Now following:", feed.Name)
	return nil
}
