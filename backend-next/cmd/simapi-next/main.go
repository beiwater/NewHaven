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
	"github.com/newhaven/backend-next/internal/app/building"
	"github.com/newhaven/backend-next/internal/app/company"
	"github.com/newhaven/backend-next/internal/app/finance"
	"github.com/newhaven/backend-next/internal/app/market"
	"github.com/newhaven/backend-next/internal/app/production"
	"github.com/newhaven/backend-next/internal/app/warehouse"
	"github.com/newhaven/backend-next/internal/catalog"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/httpapi"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg := config.Load()
	st := memory.New()
	application := app.New(cfg, st)
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

	companySvc := company.NewService(st, application.Logger)
	companyHandler := httpapi.NewCompanyHandler(companySvc)
	authHandler := httpapi.NewAuthHandler(application.AuthService)
	warehouseSvc := warehouse.NewService(st, st, cfg.Game, application.Logger)
	warehouseHandler := httpapi.NewWarehouseHandler(warehouseSvc)
	buildingSvc := building.NewService(st, buildings, cfg.Game, application.Clock, application.IDGen)
	buildingHandler := httpapi.NewBuildingHandler(buildingSvc)

	productionSvc := production.NewService(st, st, st, cfg.Game, resources, buildings, application.Clock, application.IDGen)

	marketSvc := market.NewService(st, st, st, resources, cfg.Game, application.Clock, application.IDGen)
	marketHandler := httpapi.NewMarketHandler(marketSvc)

	financeSvc := finance.NewService(st, st, application.Clock, application.IDGen, cfg.Game)
	financeHandler := httpapi.NewFinanceHandler(financeSvc)
	bondHandler := httpapi.NewBondHandler(financeSvc)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	mux := httpapi.NewRouter(cfg, authHandler, companyHandler, warehouseHandler, buildingHandler, productionHandler, marketHandler, financeHandler, bondHandler)
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
}
