package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anurinth-w/conduit-th/services/document/config"
	"github.com/anurinth-w/conduit-th/services/document/handler"
	"github.com/anurinth-w/conduit-th/services/document/pdf"
	"github.com/anurinth-w/conduit-th/services/document/repository"
	"github.com/anurinth-w/conduit-th/services/document/service"
	"github.com/anurinth-w/conduit-th/services/document/storage"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	// Database
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("database connected")

	// MinIO
	minioStorage, err := storage.NewMinIOStorage(
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.MinIOBucket,
		cfg.MinIOUseSSL,
		cfg.PresignExpirySec,
	)
	if err != nil {
		log.Fatalf("connect minio: %v", err)
	}
	log.Println("minio connected")

	// Gotenberg
	pdfClient := pdf.NewGotenbergClient(cfg.GotenbergURL)
	log.Printf("gotenberg client ready: %s", cfg.GotenbergURL)

	docRepo := repository.NewDocumentRepository(db)
	docSvc := service.NewDocumentService(docRepo, pdfClient, minioStorage)
	docHandler := handler.NewDocumentHandler(docSvc)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.GET("/health", docHandler.Health)

	v1 := r.Group("/v1")
	{
		// Templates (Dev จัดการ)
		v1.POST("/companies/:company_id/templates", docHandler.CreateTemplate)
		v1.GET("/companies/:company_id/templates", docHandler.ListTemplates)
		v1.GET("/templates/:id", docHandler.GetTemplate)

		// Bundles (กำหนดชุดเอกสารต่อบริษัท × job_type)
		v1.POST("/companies/:company_id/bundles", docHandler.CreateBundle)

		// Generate PDF
		v1.POST("/jobs/:job_id/documents/generate", docHandler.Generate)
		v1.GET("/jobs/:job_id/documents", docHandler.ListByJob)
	}

	srv := &http.Server{
		Addr:    ":8006",
		Handler: r,
	}

	go func() {
		log.Println("document service starting on :8006")
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
