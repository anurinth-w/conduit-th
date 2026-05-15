package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anurinth-w/conduit-th/services/media/config"
	"github.com/anurinth-w/conduit-th/services/media/handler"
	"github.com/anurinth-w/conduit-th/services/media/repository"
	"github.com/anurinth-w/conduit-th/services/media/service"
	"github.com/anurinth-w/conduit-th/services/media/storage"
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

	photoRepo := repository.NewPhotoRepository(db)
	mediaSvc := service.NewMediaService(photoRepo, minioStorage)
	mediaHandler := handler.NewMediaHandler(mediaSvc)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// จำกัดขนาดไฟล์ที่อัปโหลด 20MB
	r.MaxMultipartMemory = 20 << 20

	r.GET("/health", mediaHandler.Health)

	v1 := r.Group("/v1")
	{
		// รูปต่องาน
		v1.POST("/jobs/:job_id/photos", mediaHandler.Upload)
		v1.GET("/jobs/:job_id/photos", mediaHandler.ListByJob)
		v1.POST("/jobs/:job_id/photos/refresh-urls", mediaHandler.RefreshURLs)

		// จัดการรูปรายชิ้น
		v1.GET("/photos/:id/url", mediaHandler.GetPresignURL)
		v1.PATCH("/photos/:id/select", mediaHandler.SetSelected)
		v1.DELETE("/photos/:id", mediaHandler.Delete)
	}

	srv := &http.Server{
		Addr:    ":8005",
		Handler: r,
	}

	go func() {
		log.Println("media service starting on :8005")
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
