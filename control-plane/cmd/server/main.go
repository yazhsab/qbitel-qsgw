package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/quantun-opensource/qsgw/control-plane/internal/config"
	"github.com/quantun-opensource/qsgw/control-plane/internal/handler"
	"github.com/quantun-opensource/qsgw/control-plane/internal/repository"
	"github.com/quantun-opensource/qsgw/control-plane/internal/service"
	qdb "github.com/quantun-opensource/qsgw/shared/go/database"
	qmw "github.com/quantun-opensource/qsgw/shared/go/middleware"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	ctx := context.Background()
	poolCfg := qdb.DefaultPoolConfig(cfg.DatabaseURL)
	poolCfg.Logger = logger
	pool, err := qdb.NewPool(ctx, poolCfg)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	// Repositories
	gatewayRepo := repository.NewGatewayRepository(pool)
	upstreamRepo := repository.NewUpstreamRepository(pool)
	routeRepo := repository.NewRouteRepository(pool)
	threatRepo := repository.NewThreatRepository(pool)

	// Services
	gatewaySvc := service.NewGatewayService(gatewayRepo, logger)
	upstreamSvc := service.NewUpstreamService(upstreamRepo, logger)
	routeSvc := service.NewRouteService(routeRepo, logger)

	// Handlers
	healthH := handler.NewHealthHandler()
	gatewayH := handler.NewGatewayHandler(gatewaySvc, logger)
	upstreamH := handler.NewUpstreamHandler(upstreamSvc, logger)
	routeH := handler.NewRouteHandler(routeSvc, logger)
	threatH := handler.NewThreatHandler(threatRepo, logger)

	// Router
	r := chi.NewRouter()

	// --- Infrastructure middleware ---
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))

	// --- Security middleware ---
	r.Use(qmw.SecurityHeaders(qmw.DefaultSecurityHeadersConfig()))
	r.Use(qmw.MaxBodySize(cfg.MaxBodyBytes))

	// CORS (only if origins are configured)
	if len(cfg.CORSOrigins) > 0 {
		corsConfig := qmw.DefaultCORSConfig()
		corsConfig.AllowedOrigins = cfg.CORSOrigins
		r.Use(qmw.CORS(corsConfig))
	}

	// Rate limiting (100 req/min per IP)
	rateLimiter := qmw.NewRateLimiter(qmw.DefaultRateLimitConfig())
	defer rateLimiter.Stop()
	r.Use(rateLimiter.Middleware())

	// --- Auth middleware ---
	authConfig := qmw.AuthConfig{
		JWTSecret: cfg.JWTSecret,
		JWTIssuer: cfg.JWTIssuer,
		APIKeys:   qmw.ParseAPIKeyEntries(cfg.APIKeys),
		SkipPaths: []string{"/health"},
		Logger:    logger,
	}

	// Health endpoint (unauthenticated)
	r.Get("/health", healthH.Check)

	// API routes (authenticated)
	r.Route("/api/v1", func(r chi.Router) {
		// Apply auth only if JWT secret or API keys are configured
		if cfg.JWTSecret != "" || len(cfg.APIKeys) > 0 {
			r.Use(qmw.Auth(authConfig))
		}

		r.Mount("/gateways", gatewayH.Routes())
		r.Mount("/upstreams", upstreamH.Routes())
		r.Mount("/routes", routeH.Routes())
		r.Mount("/threats", threatH.Routes())
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting QSGW control-plane server",
			zap.String("addr", addr),
			zap.Bool("auth_enabled", cfg.JWTSecret != "" || len(cfg.APIKeys) > 0),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}
	logger.Info("server stopped")
}
