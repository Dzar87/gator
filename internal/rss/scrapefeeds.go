package rss

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Dzar87/gator/internal/database"
	"github.com/Dzar87/gator/internal/state"
	"github.com/lib/pq"
)

func classifyScrapeFeedDBErr(err error) error {
	if err == nil {
		return nil
	}
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return fmt.Errorf("scrapefeeds: creating post: %w", err)
	}
	if pqErr.Code.Name() == "unique_violation" &&
		pqErr.Constraint == "posts_url_key" {
		return nil
	}
	return fmt.Errorf("scrapefeeds: db error: %w", err)
}

func ScrapeFeeds(ctx context.Context, s *state.State) error {
	feed, err := s.Queries.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("scrapefeeds: failed to get feed: %w", err)
	}
	feed, err = s.Queries.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		return fmt.Errorf("scrapefeeds: failed to mark feed: %w", err)
	}
	rssFeed, err := FetchFeed(ctx, feed.Url)
	if err != nil {
		return fmt.Errorf("scrapefeeds: failed to fetch feed: %w", err)
	}
	fmt.Println(rssFeed.Channel.Title, ":")
	var qParams database.CreatePostParams
	for _, item := range rssFeed.Channel.Item {
		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			s.Logger.Warn("Unable to parse pubDate", "rawPubDate", item.PubDate)
		}
		qParams = database.CreatePostParams{
			Title: sql.NullString{
				String: item.Title,
				Valid:  true,
			},
			Url: item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid:  true,
			},
			PublishedAt: sql.NullTime{
				Time:  pubDate,
				Valid: (pubDate != time.Time{}),
			},
			FeedID: feed.ID,
		}

		_, err = s.Queries.CreatePost(ctx, qParams)
		if err := classifyScrapeFeedDBErr(err); err != nil {
			s.Logger.Error("Failed to save post", "err", err)
		}
	}
	return nil
}
