package main

import (
"context"
"log"
"net/http"
"os"
"os/signal"
"syscall"
"time"

"github.com/anurinth-w/conduit-th/services/user/config"
"github.com/anurinth-w/conduit-th/services/user/handler"
"github.com/anurinth-w/conduit-th/services/user/repository"
"github.com/anurinth-w/conduit-th/services/user/service"
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

userRepo := repository.NewUserRepository(db)
userSvc := service.NewUserService(userRepo)
userHandler := handler.NewUserHandler(userSvc)

if cfg.Env == "production" {
gin.SetMode(gin.ReleaseMode)
}

r := gin.Default()
r.GET("/health", userHandler.Health)

v1 := r.Group("/v1")
{
users := v1.Group("/users")
{
users.POST("", userHandler.Create)
users.GET("/:id", userHandler.GetByID)
users.PUT("/:id", userHandler.Update)
users.DELETE("/:id", userHandler.Deactivate)
users.GET("/:id/memberships", userHandler.GetMemberships)
users.POST("/memberships", userHandler.AddMembership)
}

companies := v1.Group("/companies")
{
companies.GET("/:company_id/members", userHandler.ListByCompany)
}
}

srv := &http.Server{
Addr:    ":8002",
Handler: r,
}

go func() {
log.Println("user service starting on :8002")
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
