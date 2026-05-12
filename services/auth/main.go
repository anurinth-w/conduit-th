package main

import (
"context"
"log"
"net/http"
"os"
"os/signal"
"syscall"
"time"

"github.com/anurinth-w/conduit-th/services/auth/config"
"github.com/anurinth-w/conduit-th/services/auth/handler"
"github.com/anurinth-w/conduit-th/services/auth/repository"
"github.com/anurinth-w/conduit-th/services/auth/service"
"github.com/gin-gonic/gin"
"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
cfg := config.Load()

// database connection
db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
if err != nil {
log.Fatalf("connect db: %v", err)
}
defer db.Close()

if err := db.Ping(context.Background()); err != nil {
log.Fatalf("ping db: %v", err)
}
log.Println("database connected")

// wire up layers
userRepo := repository.NewUserRepository(db)
authSvc := service.NewAuthService(userRepo, cfg)
authHandler := handler.NewAuthHandler(authSvc)

// router
if cfg.Env == "production" {
gin.SetMode(gin.ReleaseMode)
}

r := gin.Default()

// routes
r.GET("/health", authHandler.Health)

v1 := r.Group("/v1/auth")
{
v1.POST("/login", authHandler.Login)
v1.POST("/refresh", authHandler.Refresh)
v1.POST("/logout", authHandler.Logout)
}

// graceful shutdown
srv := &http.Server{
Addr:    ":8001",
Handler: r,
}

go func() {
log.Println("auth service starting on :8001")
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
