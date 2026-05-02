package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Dzar87/gator/internal/database"
	"github.com/Dzar87/gator/internal/rss"
	"github.com/Dzar87/gator/internal/state"
	"github.com/lib/pq"
)

var (
	ErrUserExists       = errors.New("user already exists")
	ErrFeedExists       = errors.New("feed already exists")
	ErrFeedFollowExists = errors.New("feed follow already exists")
	ErrUserNotFound     = errors.New("user does not exist")
	ErrFeedNotFound     = errors.New("feed does not exist")
)

func requireCurrentUser(ctx context.Context, s *state.State) (database.User, error) {
	user, err := s.Queries.GetUserByName(ctx, s.Cfg.CurrentUserName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return database.User{}, ErrUserNotFound
		}
		return database.User{}, err
	}
	return user, nil
}

func loggedIn(h func(
	ctx context.Context, s *state.State, cmd Command, user database.User,
) error) func(context.Context, *state.State, Command) error {
	return func(ctx context.Context, s *state.State, cmd Command) error {
		user, err := requireCurrentUser(ctx, s)
		if err != nil {
			return err
		}
		return h(ctx, s, cmd, user)
	}
}

func classifyLoginErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrUserNotFound
	}
	return fmt.Errorf("login: getting user: %w", err)
}

func handlerLogin(ctx context.Context, s *state.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <user_name>", cmd.Name)
	}
	if _, err := s.Queries.GetUserByName(ctx, cmd.Args[0]); err != nil {
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
		return fmt.Errorf("usage: %s <user_name>", cmd.Name)
	}
	user, err := s.Queries.CreateUser(ctx, cmd.Args[0])
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
	if err := s.Queries.DeleteUsers(ctx); err != nil {
		return fmt.Errorf("reset: delete users: %w", err)
	}
	s.Logger.Info("users reset")
	return nil
}

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
		if err := scrapeFeeds(ctx, s); err != nil {
			s.Logger.Error("Failed to scrape feed", "err", err)
		}
		select {
			case <- ctx.Done():
				ticker.Stop()
				return nil
			case <-ticker.C:
				continue
		}
	}
	return nil
}

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
		Name:      cmd.Args[0],
		Url:       cmd.Args[1],
		UserID:    user.ID,
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
		UserID:    user.ID,
		FeedID:    feed.ID,
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
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	if _, err := s.Queries.CreateFeedFollow(ctx, feedFollowParams); err != nil {
		return classifyFollowFeedErr(err)
	}
	fmt.Println("Now following:", feed.Name)
	return nil
}

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

func scrapeFeeds(ctx context.Context, s *state.State) error {
	feed, err := s.Queries.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("scrapefeeds: failed to get feed: %w", err)
	}
	feed, err = s.Queries.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		return fmt.Errorf("scrapefeeds: failed to mark feed: %w", err)
	}
	rssFeed, err := rss.FetchFeed(ctx, feed.Url)
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
				Valid: true,
			},
			Url: item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid: true,
			},
			PublishedAt: sql.NullTime{
				Time: pubDate,
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

func handlerBrowse(
	ctx context.Context, s *state.State, cmd Command, user database.User,
) error {
	if len(cmd.Args) > 1 {
		return fmt.Errorf("usage: %s <limit>", cmd.Name)
	}
	var limit int32 = 2
	if len(cmd.Args) == 1 {
		parsedLimit, err := strconv.ParseInt(cmd.Args[0], 10, 32)
		if err != nil {
			s.Logger.Warn("Invalid limit, defaulting", "err", err)
			parsedLimit = 2
		} else if parsedLimit < 1 {
			s.Logger.Warn("Invalid limit, defaulting", "err", "limit less than 1")
			parsedLimit = 2
		}
		limit = int32(parsedLimit)
	}
	qParams := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit: limit,
	}
	rows, err := s.Queries.GetPostsForUser(ctx, qParams)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		fmt.Println("No posts found.")
		return nil
	}
	tmpl := `========
Title: %s
Published At: %s
Description: %s
Link: %s
Feed: %s
`
	for _, post := range rows {
		fmt.Printf(
			tmpl,
			post.Title.String, post.PublishedAt.Time.Format(time.RFC1123Z), post.Description.String, post.Url, post.FeedName,
		)
	}
	return nil
}
