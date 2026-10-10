package http

import (
	"os"

	"famloop/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type RouterConfig struct {
	AuthHandler     *AuthHandler
	GroupHandler    *GroupHandler
	TrackingHandler *TrackingHandler
	AuthUC          *usecase.AuthUseCase
	UploadDir       string
	WebDistDir      string
}

func SetupRoutes(app *fiber.App, cfg RouterConfig) {
	// Healthcheck
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "famloop-tracking",
		})
	})

	// Static serving untuk foto profil avatar
	app.Static("/uploads", cfg.UploadDir)

	api := app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", cfg.AuthHandler.Register)
	auth.Post("/login", cfg.AuthHandler.Login)
	auth.Get("/me", JWTMiddleware(cfg.AuthUC), cfg.AuthHandler.Me)
	auth.Put("/me", JWTMiddleware(cfg.AuthUC), cfg.AuthHandler.UpdateProfile)
	auth.Post("/avatar", JWTMiddleware(cfg.AuthUC), cfg.AuthHandler.UploadAvatar)

	// Family Groups routes
	groups := api.Group("/groups")
	groups.Post("/", JWTMiddleware(cfg.AuthUC), cfg.GroupHandler.CreateGroup)
	groups.Get("/", JWTMiddleware(cfg.AuthUC), cfg.GroupHandler.GetGroups)
	groups.Get("/:id", JWTMiddleware(cfg.AuthUC), cfg.GroupHandler.GetGroup)
	groups.Delete("/:id", JWTMiddleware(cfg.AuthUC), cfg.GroupHandler.DeleteGroup)
	groups.Post("/:id/pair-device", JWTMiddleware(cfg.AuthUC), cfg.GroupHandler.PairDevice)
	groups.Delete("/:id/members/:device_id", JWTMiddleware(cfg.AuthUC), cfg.GroupHandler.RemoveMember)

	// Live Locations & SSE streaming (dapat diakses oleh hp orang tua dan hp anak)
	groups.Get("/:id/locations/latest", cfg.TrackingHandler.GetLatestLocations)
	groups.Get("/:id/members/:device_id/history", cfg.TrackingHandler.GetDeviceHistory)
	groups.Get("/:id/sse", cfg.TrackingHandler.GroupSSE)

	// Member Devices routes
	devices := api.Group("/devices")
	devices.Get("/:id", cfg.GroupHandler.GetDevice)
	devices.Post("/:id/avatar", cfg.GroupHandler.UploadDeviceAvatar)

	// Direct High-throughput Ingestion & Offline Sync routes
	tracking := api.Group("/tracking")
	tracking.Post("/ping", cfg.TrackingHandler.Ping)
	tracking.Post("/sync/batch", cfg.TrackingHandler.SyncBatch)

	// Static serving untuk Astro Landing Page & Asset SEO
	if cfg.WebDistDir != "" {
		stat, err := os.Stat(cfg.WebDistDir)
		_ = err
		if err == nil && stat.IsDir() {
			app.Static("/", cfg.WebDistDir, fiber.Static{
				Index: "index.html",
			})
		}
	}
}
