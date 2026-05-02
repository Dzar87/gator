package cli

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Dzar87/gator/internal/database"
	"github.com/Dzar87/gator/internal/state"
)

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
		Limit:  limit,
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
