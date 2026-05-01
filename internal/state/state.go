package state

import (
	"log/slog"

	"github.com/Dzar87/gator/internal/config"
	"github.com/Dzar87/gator/internal/database"
)

type State struct {
	Cfg    *config.Config
	DB     *database.Queries
	Logger *slog.Logger
}
