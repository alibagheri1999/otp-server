package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"otp-server/internal/application"
	"otp-server/internal/infrastructure/circuitbreaker"
	"otp-server/internal/infrastructure/config"
	"otp-server/internal/infrastructure/database"
	"otp-server/internal/infrastructure/events"
	"otp-server/internal/infrastructure/logger"
	"otp-server/internal/infrastructure/metrics"
	"otp-server/internal/infrastructure/redis"
	"otp-server/internal/infrastructure/retry"
	"otp-server/internal/infrastructure/shutdown"
	"otp-server/internal/interfaces/http/handlers"
	"otp-server/internal/interfaces/http/middleware"
	"otp-server/internal/interfaces/http/router"

	"github.com/joho/godotenv"
)

// @title OTP Server API
// @version 1.0
// @description A secure OTP-based authentication service with user management
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@otpserver.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(err.Error())
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log := logger.New(cfg.Log)
	ctx := context.Background()
	log.Info(ctx, "Starting OTP Server Backend", logger.F("version", "1.0.0"), logger.F("environment", cfg.Server.Environment))

	shutdownManager := shutdown.NewShutdownManager(log, 30*time.Second)
	shutdownManager.Start()

	circuitBreakerManager := circuitbreaker.NewManager(log)

	var postgresPool *database.PostgresPool

	switch cfg.Infrastructure.DatabaseProvider {
	case "postgres":
		postgresPool, err = initializePostgresPoolWithRetry(ctx, cfg, log, circuitBreakerManager)
		if err != nil {
			log.Fatal(ctx, "Failed to connect to PostgreSQL", logger.F("error", err))
		}
		defer postgresPool.Close()

		shutdownManager.AddHandler(shutdown.NewDatabaseShutdownHandler("postgres", func(ctx context.Context) error {
			return postgresPool.Close()
		}))

		log.Info(ctx, "Connected to PostgreSQL database")

	default:
		log.Fatal(ctx, "Unsupported database provider", logger.F("provider", cfg.Infrastructure.DatabaseProvider))
	}

	var redisClient *redis.Client
	switch cfg.Infrastructure.CacheProvider {
	case "redis":
		redisClient, err = initializeRedisWithRetry(ctx, cfg, log, circuitBreakerManager)
		if err != nil {
			log.Fatal(ctx, "Failed to connect to Redis", logger.F("error", err))
		}
		defer redisClient.Close()

		shutdownManager.AddHandler(shutdown.NewCacheShutdownHandler("redis", func(ctx context.Context) error {
			return redisClient.Close()
		}))

		log.Info(ctx, "Connected to Redis cache")

	default:
		log.Fatal(ctx, "Unsupported cache provider", logger.F("provider", cfg.Infrastructure.CacheProvider))
	}

	log.Info(ctx, "Initializing metrics service")
	metricsService := metrics.NewMetricsService(log)
	log.Info(ctx, "Metrics service initialized")

	repositories := database.NewRepositories(postgresPool, redisClient)

	services := application.NewServices(repositories, cfg, redisClient, metricsService)

	ctx = context.WithValue(ctx, "metrics", metricsService)

	log.Info(ctx, "Initializing event listener")
	eventListener := events.NewEventListener(&cfg.Events, log)
	log.Info(ctx, "Event listener initialized")

	go func() {
		log.Info(ctx, "Starting event listener")
		if err := eventListener.StartEventListener(ctx, services.GetEventService()); err != nil {
			log.Error(ctx, "Failed to start event listener", logger.F("error", err))
		}
		log.Info(ctx, "Event listener started successfully")
	}()

	handlers := handlers.NewHandlers(services, log)

	middleware := middleware.NewMiddleware(cfg, log, redisClient)

	middleware.SetAuthService(services.AuthService)
	middleware.SetMetricsService(metricsService)

	log.Info(ctx, "Creating Fiber router")
	fiberApp := router.NewRouter(handlers, middleware, cfg)
	log.Info(ctx, "Fiber router created successfully")

	shutdownManager.AddHandler(shutdown.NewServerShutdownHandler("http", func(ctx context.Context) error {
		return fiberApp.Shutdown()
	}))

	go func() {
		log.Info(ctx, "Starting HTTP server",
			logger.F("port", cfg.Server.Port),
			logger.F("environment", cfg.Server.Environment),
			logger.F("database_provider", cfg.Infrastructure.DatabaseProvider),
			logger.F("cache_provider", cfg.Infrastructure.CacheProvider),
			logger.F("storage_provider", cfg.Infrastructure.StorageProvider))

		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		log.Info(ctx, "Attempting to listen on address", logger.F("address", addr))

		if err := fiberApp.Listen(addr); err != nil {
			log.Fatal(ctx, "Failed to start server", logger.F("error", err))
		}
	}()

	shutdownManager.Wait()
	log.Info(ctx, "Server exited gracefully")
}

// initializePostgresPoolWithRetry initializes PostgreSQL pool with retry and circuit breaker
func initializePostgresPoolWithRetry(ctx context.Context, cfg *config.Config, logger logger.Logger, cbManager *circuitbreaker.CircuitBreakerManager) (*database.PostgresPool, error) {
	cb := cbManager.GetOrCreate("postgres", circuitbreaker.DefaultConfig())

	retryConfig := retry.RetryWithLogger(logger, "postgres_pool_connection")

	var postgresPool *database.PostgresPool
	err := retry.Retry(ctx, retryConfig, func() error {
		return cb.Execute(ctx, func() error {
			var err error
			postgresPool, err = database.NewPostgresPool(&cfg.Database, logger)
			return err
		})
	})

	return postgresPool, err
}

// initializeRedisWithRetry initializes Redis with retry and circuit breaker
func initializeRedisWithRetry(ctx context.Context, cfg *config.Config, logger logger.Logger, cbManager *circuitbreaker.CircuitBreakerManager) (*redis.Client, error) {
	cb := cbManager.GetOrCreate("redis", circuitbreaker.DefaultConfig())

	retryConfig := retry.RetryWithLogger(logger, "redis_connection")

	var redisClient *redis.Client
	err := retry.Retry(ctx, retryConfig, func() error {
		return cb.Execute(ctx, func() error {
			var err error
			redisClient, err = redis.NewClient(cfg.Redis)
			return err
		})
	})

	return redisClient, err
}
