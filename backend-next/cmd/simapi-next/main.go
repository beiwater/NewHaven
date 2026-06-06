package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/newhaven/backend-next/internal/app"
	"github.com/newhaven/backend-next/internal/catalog"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/httpapi"
	"github.com/newhaven/backend-next/internal/scheduler"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	st := memory.New()
	// Load static game data catalogs (best-effort in dev mode).
	projectRoot := config.FindProjectRoot()
	resources, err := catalog.LoadResources(projectRoot)
	if err != nil {
		slog.Warn("failed to load resources catalog, production start will fail", "error", err)
		resources = make(map[int]*catalog.ResourceEntry)
	}
	buildings, err := catalog.LoadBuildings(projectRoot)
	if err != nil {
		slog.Warn("failed to load buildings catalog, production start will fail", "error", err)
		buildings = make(map[int]*catalog.BuildingEntry)
	}

	application := app.New(cfg, st, resources, buildings)

	// Scheduler for bot economy and background tasks including bond interest
	sched := scheduler.New(application.MarketService, func(ctx context.Context) error {
		_, err := application.FinanceService.SettleBondInterest(ctx)
		return err
	}, application.SaveAll)
	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{
		Auth: application.AuthHandler, Company: application.CompanyHandler, Warehouse: application.WarehouseHandler,
		Building: application.BuildingHandler, Production: application.ProductionHandler, Market: application.MarketHandler,
		Finance: application.FinanceHandler, Bond: application.BondHandler, Player: application.PlayerHandler,
		Social: application.SocialHandler, Contract: application.ContractHandler, Research: application.ResearchHandler,
		Executive: application.ExecutiveHandler, Leaderboard: application.LeaderboardHandler,
	})
	if cfg.DevMode {
		slog.Info("dev mode enabled, bootstrapping dev user")
		if err := application.AuthService.DevBootstrap(context.Background()); err != nil {
			slog.Warn("dev bootstrap skipped", "error", err)
		}
	}

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start background scheduler
	sched.Start()

	go func() {
		slog.Info("simapi-next starting", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown failed", "error", err)
	}

	if application.PostgresStore != nil {
		application.PostgresStore.Close()
	}
}
