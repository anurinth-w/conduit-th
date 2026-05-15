package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anurinth-w/conduit-th/services/gateway/config"
	"github.com/anurinth-w/conduit-th/services/gateway/middleware"
	"github.com/anurinth-w/conduit-th/services/gateway/proxy"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Global middlewares
	r.Use(middleware.RateLimit(cfg.RateLimitRPS, cfg.RateLimitBurst))

	// Health check (ไม่ต้อง JWT)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "gateway"})
	})

	// ============================================================
	// Public routes — ไม่ต้อง JWT
	// ============================================================
	public := r.Group("/v1")
	{
		public.POST("/auth/login", proxy.Forward(cfg.AuthURL))
		public.POST("/auth/refresh", proxy.Forward(cfg.AuthURL))
	}

	// ============================================================
	// Protected routes — ต้อง JWT
	// ============================================================
	protected := r.Group("/v1")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Auth
		protected.POST("/auth/logout", proxy.Forward(cfg.AuthURL))

		// Users
		protected.GET("/users/:id", proxy.Forward(cfg.UserURL))
		protected.PATCH("/users/:id", proxy.Forward(cfg.UserURL))
		protected.GET("/users/:id/memberships", proxy.Forward(cfg.UserURL))
		protected.GET("/companies/:company_id/members", proxy.Forward(cfg.UserURL))

		// Jobs
		protected.POST("/jobs", proxy.Forward(cfg.JobURL))
		protected.GET("/jobs/:id", proxy.Forward(cfg.JobURL))
		protected.PATCH("/jobs/:id/status", proxy.Forward(cfg.JobURL))
		protected.POST("/jobs/:id/assign", proxy.Forward(cfg.JobURL))
		protected.GET("/jobs/:id/assignments", proxy.Forward(cfg.JobURL))
		protected.GET("/companies/:company_id/jobs", proxy.Forward(cfg.JobURL))

		// Materials
		protected.POST("/companies/:company_id/materials", proxy.Forward(cfg.MaterialURL))
		protected.GET("/companies/:company_id/materials", proxy.Forward(cfg.MaterialURL))
		protected.GET("/materials/:id", proxy.Forward(cfg.MaterialURL))
		protected.PATCH("/materials/:id", proxy.Forward(cfg.MaterialURL))
		protected.DELETE("/materials/:id", proxy.Forward(cfg.MaterialURL))
		protected.PATCH("/materials/:id/price", proxy.Forward(cfg.MaterialURL))
		protected.GET("/materials/:id/price-history", proxy.Forward(cfg.MaterialURL))

		// Media
		protected.POST("/jobs/:job_id/photos", proxy.Forward(cfg.MediaURL))
		protected.GET("/jobs/:job_id/photos", proxy.Forward(cfg.MediaURL))
		protected.POST("/jobs/:job_id/photos/refresh-urls", proxy.Forward(cfg.MediaURL))
		protected.GET("/photos/:id/url", proxy.Forward(cfg.MediaURL))
		protected.PATCH("/photos/:id/select", proxy.Forward(cfg.MediaURL))
		protected.DELETE("/photos/:id", proxy.Forward(cfg.MediaURL))

		// Documents
		protected.POST("/companies/:company_id/templates", proxy.Forward(cfg.DocumentURL))
		protected.GET("/companies/:company_id/templates", proxy.Forward(cfg.DocumentURL))
		protected.GET("/templates/:id", proxy.Forward(cfg.DocumentURL))
		protected.POST("/companies/:company_id/bundles", proxy.Forward(cfg.DocumentURL))
		protected.POST("/jobs/:job_id/documents/generate", proxy.Forward(cfg.DocumentURL))
		protected.GET("/jobs/:job_id/documents", proxy.Forward(cfg.DocumentURL))

		// Notify
		protected.POST("/notify", proxy.Forward(cfg.NotifyURL))
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("gateway starting on :%s", cfg.Port)
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
