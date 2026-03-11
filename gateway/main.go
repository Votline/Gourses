package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gc "gateway/internal/gracefulshutdown"
	"gateway/internal/routers"

	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewDevelopment()
	srv := routers.Init(log)
	go func() {
		if err := srv.Srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("ListenAndServe", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	gracefulShutdown(srv, log)
}

func gracefulShutdown(srv *routers.Server, log *zap.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info("Shutting down server")
	if err := gc.Shutdown(srv.Srv.Close, ctx); err != nil {
		log.Error("Error shutting down server", zap.Error(err))
	}

	log.Info("Shutting down services")
	if err := srv.ShutdownServices(ctx, log); err != nil {
		log.Error("Error shutting down services", zap.Error(err))
	}
}
