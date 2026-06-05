package app

import (
	"log/slog"

	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
)

// App is the top-level container for all application use cases.
// It wires together domain services, storage, and platform utilities.
type App struct {
	Config  *config.Config
	Storage storage.Storage
	Clock   platform.Clock
	IDGen   *platform.IDGen
	Logger  *platform.Logger
}

func New(cfg *config.Config, st storage.Storage) *App {
	return &App{
		Config:  cfg,
		Storage: st,
		Clock:   platform.RealClock{},
		IDGen:   platform.NewIDGen(),
		Logger:  platform.NewLogger(slog.Default()),
	}
}
