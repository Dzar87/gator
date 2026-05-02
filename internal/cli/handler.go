package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Dzar87/gator/internal/database"
	"github.com/Dzar87/gator/internal/rss"
	"github.com/Dzar87/gator/internal/state"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrUserExists       = errors.New("user already exists")
	ErrFeedExists       = errors.New("feed already exists")
	ErrFeedFollowExists = errors.New("feed follow already exists")
	ErrUserNotFound     = errors.New("user does not exist")
	ErrFeedNotFound     = errors.New("feed does not exist")
	ErrBadArgs          = errors.New("bad command arguments")
)

func requireCurrentUser(ctx context.Context, s *state.State) (database.User, error) {
	user, err := s.Queries.GetUser(ctx, s.Cfg.CurrentUserName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return database.User{}, ErrUserNotFound
		}
		return database.User{}, err
	}
	return user, nil
}

func classifyLoginErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrUserNotFound
	}
	return fmt.Errorf("login: getting user: %w", err)
}

func handlerLogin(ctx context.Context, s *state.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("login: %w", ErrBadArgs)
	}
	if _, err := s.Queries.GetUser(ctx, cmd.Args[0]); err != nil {
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
		return fmt.Errorf("register: %w", ErrBadArgs)
	}
	now := time.Now().UTC()
	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      cmd.Args[0],
	}
	user, err := s.Queries.CreateUser(ctx, userParams)
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
	url := "https://www.wagslane.dev/index.xml"
	feed, err := rss.FetchFeed(ctx, url)
	if err != nil {
		return fmt.Errorf("agg: failed to fetch feed: %w", err)
	}
	fmt.Printf("Feed: %+v\n", feed)
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

func handlerAddFeed(ctx context.Context, s *state.State, cmd Command) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("addfeed: %w", ErrBadArgs)
	}
	user, err := requireCurrentUser(ctx, s)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
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
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
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

func handlerFollowFeed(ctx context.Context, s *state.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("followfeed: %w", ErrBadArgs)
	}
	user, err := requireCurrentUser(ctx, s)
	if err != nil {
		return err
	}
	feed, err := s.Queries.GetFeed(ctx, cmd.Args[0])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrFeedNotFound
		}
		return err
	}
	now := time.Now().UTC()
	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	if _, err := s.Queries.CreateFeedFollow(ctx, feedFollowParams); err != nil {
		return classifyFollowFeedErr(err)
	}
	fmt.Println("Now following:", feed.Name)
	return nil
}

func handlerFollowing(ctx context.Context, s *state.State, cmd Command) error {
	user, err := requireCurrentUser(ctx, s)
	if err != nil {
		return err
	}
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
