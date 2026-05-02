package cli

import (
	"context"
	"fmt"

	"github.com/Dzar87/gator/internal/state"
)

func handlerListFeeds(ctx context.Context, s *state.State, cmd Command) error {
	rows, err := s.Queries.GetFeedsWithUser(ctx)
	if err != nil {
		return fmt.Errorf("listfeeds: couldn't list feeds: %w", err)
	}
	if len(rows) == 0 {
		fmt.Println("No feeds found.")
		return nil
	}
	tmpl := `* Name: %s
  Url: %s
  Added By: %s
`
	for _, row := range rows {
		fmt.Printf(tmpl, row.Feed.Name, row.Feed.Url, row.User.Name)
		s.Logger.Debug("feed row", "feed", row.Feed, "user", row.User)
	}
	return nil
}
