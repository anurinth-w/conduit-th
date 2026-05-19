package main

import (
"context"
"log"
"net/http"
"os"
"os/signal"
"syscall"
"time"

"github.com/anurinth-w/conduit-th/services/job/config"
"github.com/anurinth-w/conduit-th/services/job/handler"
"github.com/anurinth-w/conduit-th/services/job/repository"
"github.com/anurinth-w/conduit-th/services/job/service"
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

jobRepo := repository.NewJobRepository(db)
	matRepo := repository.NewMaterialRepository(db)
	matHandler := handler.NewMaterialHandler(matRepo)
jobSvc := service.NewJobService(jobRepo)
jobHandler := handler.NewJobHandler(jobSvc)

if cfg.Env == "production" {
gin.SetMode(gin.ReleaseMode)
}

r := gin.Default()
r.GET("/health", jobHandler.Health)

v1 := r.Group("/v1")
{
v1.POST("/jobs", jobHandler.Create)
v1.GET("/jobs/:id", jobHandler.GetByID)
v1.PATCH("/jobs/:id/status", jobHandler.UpdateStatus)
v1.POST("/jobs/:id/assign", jobHandler.Assign)
v1.GET("/jobs/:id/assignments", jobHandler.GetAssignments)
v1.GET("/companies/:company_id/jobs", jobHandler.ListByCompany)
		v1.POST("/jobs/:id/materials", matHandler.Add)
		v1.GET("/jobs/:id/materials", matHandler.List)
		v1.DELETE("/jobs/:id/materials/:material_id", matHandler.Delete)
}

srv := &http.Server{
Addr:    ":8003",
Handler: r,
}

go func() {
log.Println("job service starting on :8003")
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
