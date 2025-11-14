package app

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"os"
	"time"
	"log"
	"fmt"

	"github.com/gorilla/mux"
)

type App struct {
	config *Config
}

func NewApp(config *Config) *App {
	app := &App{
		config: config,
	}

	return app
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := mux.NewRouter()

	serverAddr := fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port)
	srv := http.Server{
		Addr: serverAddr,
		Handler: r,
	}

	log.Printf("Starting server on %s", serverAddr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\\n", err)
		}
	}()
		
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(
		ctx,
		time.Duration(a.config.Server.ShutdownTimeout)*time.Second,
	)
	defer shutdownCancel()

	return srv.Shutdown(shutdownCtx)
}
