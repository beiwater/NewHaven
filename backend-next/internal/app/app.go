package app

import (
	"log/slog"

	"github.com/newhaven/backend-next/internal/app/auth"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
)

type App struct {
	Config      *config.Config
	Storage     storage.Storage
	Clock       platform.Clock
	IDGen       *platform.IDGen
	Logger      *platform.Logger
	AuthService *auth.Service
}

func New(cfg *config.Config, st storage.Storage) *App {
	clock := platform.RealClock{}
	idgen := platform.NewIDGen()
	logger := platform.NewLogger(slog.Default())

	authSvc := auth.NewService(st, st, clock, idgen, logger, cfg.JWTSigningKey)

	return &App{
		Config:      cfg,
		Storage:     st,
		Clock:       clock,
		IDGen:       idgen,
		Logger:      logger,
		AuthService: authSvc,
	}
}
