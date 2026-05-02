package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/Dzar87/gator/internal/rss"
	"github.com/Dzar87/gator/internal/state"
)

func handlerAggregate(ctx context.Context, s *state.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <time_between_reqs>", cmd.Name)
	}
	interval, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("agg: failed to parse interval: %w", err)
	}
	ticker := time.NewTicker(interval)
	fmt.Println("Collecting feeds every", interval)
	for {
		if err := rss.ScrapeFeeds(ctx, s); err != nil {
			s.Logger.Error("Failed to scrape feed", "err", err)
		}
		select {
		case <-ctx.Done():
			ticker.Stop()
			return nil
		case <-ticker.C:
			continue
		}
	}
	return nil
}
