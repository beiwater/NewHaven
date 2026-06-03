package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go-sim-api/internal/config"
	"go-sim-api/internal/data"
	"go-sim-api/internal/formula"
	"go-sim-api/internal/handler"
	"go-sim-api/internal/middleware"
	"go-sim-api/internal/scheduler"
	"go-sim-api/internal/service"
	"go-sim-api/internal/storage"
)

func findProjectRoot(wd string) string {
	root := wd
	for {
		if _, err := os.Stat(filepath.Join(root, "decompiled", "data")); err == nil {
			return root
		}
		parent := filepath.Dir(root)
		if parent == root {
			log.Fatalf("cannot find decompiled/data from %s", wd)
		}
		root = parent
	}
}

func main() {
	cfg := config.Load()
	formula.SetBondFaceValue(cfg.Game.BondFaceValue)

	var st storage.Storage = &storage.NoopStorage{}
	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var err error
		st, err = storage.New(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("storage: %v", err)
		}
		defer st.Close()
	} else {
		log.Println("no database url, running memory-only with NoopStorage")
	}

	wd, _ := os.Getwd()
	d, err := data.Load(findProjectRoot(wd))
	if err != nil {
		log.Fatalf("load data: %v", err)
	}

	svc := service.New(d, cfg, st)

	sched := scheduler.New(svc)
	sched.Start()
	defer sched.Stop()

	h := handler.New(svc, cfg.JWTSigningKey)

	mux := http.ServeMux{}
	h.Register(&mux)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      middleware.RequestID(middleware.Logger(middleware.CORS(middleware.Recovery(&mux)))),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("sim api on http://%s", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}
