package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anurinth-w/conduit-th/services/notify/config"
	"github.com/anurinth-w/conduit-th/services/notify/handler"
	"github.com/anurinth-w/conduit-th/services/notify/line"
	"github.com/anurinth-w/conduit-th/services/notify/repository"
	"github.com/anurinth-w/conduit-th/services/notify/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("database connected")

	lineClient := line.NewClient()
	notifyRepo := repository.NewNotifyRepository(db)
	notifySvc := service.NewNotifyService(notifyRepo, lineClient)
	notifyHandler := handler.NewNotifyHandler(notifySvc)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.GET("/health", notifyHandler.Health)

	v1 := r.Group("/v1")
	{
		v1.POST("/notify", notifyHandler.Send)
	}

	srv := &http.Server{
		Addr:    ":8007",
		Handler: r,
	}

	go func() {
		log.Println("notify service starting on :8007")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
