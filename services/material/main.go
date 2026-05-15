package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anurinth-w/conduit-th/services/material/config"
	"github.com/anurinth-w/conduit-th/services/material/handler"
	"github.com/anurinth-w/conduit-th/services/material/repository"
	"github.com/anurinth-w/conduit-th/services/material/service"
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

	materialRepo := repository.NewMaterialRepository(db)
	materialSvc := service.NewMaterialService(materialRepo)
	materialHandler := handler.NewMaterialHandler(materialSvc)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.GET("/health", materialHandler.Health)

	v1 := r.Group("/v1")
	{
		// รายการวัสดุต่อบริษัท
		v1.POST("/companies/:company_id/materials", materialHandler.Create)
		v1.GET("/companies/:company_id/materials", materialHandler.ListByCompany)

		// จัดการวัสดุรายชิ้น
		v1.GET("/materials/:id", materialHandler.GetByID)
		v1.PATCH("/materials/:id", materialHandler.Update)
		v1.DELETE("/materials/:id", materialHandler.Delete)

		// จัดการราคา
		v1.PATCH("/materials/:id/price", materialHandler.UpdatePrice)
		v1.GET("/materials/:id/price-history", materialHandler.GetPriceHistory)
	}

	srv := &http.Server{
		Addr:    ":8004",
		Handler: r,
	}

	go func() {
		log.Println("material service starting on :8004")
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
