package state

import (
	"database/sql"
	"log/slog"

	"github.com/Dzar87/gator/internal/config"
	"github.com/Dzar87/gator/internal/database"
)

type State struct {
	Cfg     *config.Config
	DB      *sql.DB
	Queries *database.Queries
	Logger  *slog.Logger
}
