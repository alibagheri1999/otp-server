package router

import (
	"net/http"
	"time"

	_ "otp-server/docs" // Import generated Swagger docs
	"otp-server/internal/infrastructure/config"
	"otp-server/internal/infrastructure/metrics"
	"otp-server/internal/interfaces/http/handlers"
	"otp-server/internal/interfaces/http/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// NewRouter creates a new Fiber app with all routes
func NewRouter(handlers *handlers.Handlers, mw *middleware.Middleware, cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadTimeout:           15 * time.Second,
		WriteTimeout:          15 * time.Second,
		IdleTimeout:           60 * time.Second,
		AppName:               "otp-server",
	})

	app.Use(mw.ErrorHandler())
	app.Use(mw.Logging())
	app.Use(mw.SecurityHeaders())
	app.Use(mw.CORS())

	var metricsService *metrics.MetricsService
	if mw.GetMetricsService() != nil {
		metricsService = mw.GetMetricsService()
	}

	rateLimiter := middleware.NewRateLimitMiddleware(cfg, mw.GetLogger(), mw.GetRedisClient(), metricsService)
	app.Use(rateLimiter.Global())

	app.Use(rateLimiter.AddRateLimitHeaders())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(map[string]interface{}{
			"status":  "ok",
			"service": "otp-server",
			"version": "1.0.0",
		})
	})

	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	v1 := app.Group("/api/v1")

	auth := v1.Group("/auth")
	auth.Use(rateLimiter.Auth())
	auth.Post("/send-otp", rateLimiter.OTP(), handlers.AuthHandler.SendOTP)
	auth.Post("/verify-otp", handlers.AuthHandler.VerifyOTP)

	protected := v1.Group("")
	protected.Use(mw.Auth())

	users := protected.Group("/users")
	users.Use(rateLimiter.User()) // Rate limiting for user operations
	users.Get("/profile", handlers.UserHandler.GetProfile)
	users.Put("/profile", handlers.UserHandler.UpdateProfile)
	users.Get("/search", handlers.UserHandler.SearchUsers)

	if cfg.Server.Environment == "development" {
		app.Get("/swagger/*", fiberSwagger.WrapHandler)
	}

	return app
}
