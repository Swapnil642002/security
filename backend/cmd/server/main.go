package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"firewall-manager/internal/auth"
	"firewall-manager/internal/config"
	"firewall-manager/internal/database"
	"firewall-manager/internal/http/handlers"
	"firewall-manager/internal/http/router"
	"firewall-manager/internal/repository"
	"firewall-manager/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepository(pool)
	dashboardRepo := repository.NewDashboardRepository(pool)
	policyRepo := repository.NewPolicyRepository(pool)
	fleetRepo := repository.NewFleetRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)
	firewallProvider, providerTag, err := service.BuildProvider(
		cfg.FirewallProvider,
		cfg.FirewallDryRun,
		cfg.NFTablesBin,
		cfg.OPNsenseBaseURL,
		cfg.OPNsenseAPIKey,
		cfg.OPNsenseAPISecret,
	)
	if err != nil {
		log.Fatalf("firewall provider config error: %v", err)
	}
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTTTL)
	authService := service.NewAuthService(userRepo, jwtManager)
	dashboardService := service.NewDashboardService(dashboardRepo)
	policyService := service.NewPolicyService(policyRepo, userRepo)
	fleetService := service.NewFleetService(fleetRepo, userRepo)
	enrollmentService := service.NewEnrollmentService(enrollmentRepo, userRepo, fleetRepo, cfg.AppBaseURL)
	firewallSyncService := service.NewFirewallSyncService(userRepo, policyRepo, firewallProvider, providerTag)
	authHandler := handlers.NewAuthHandler(authService, cfg.BootstrapToken)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	policyHandler := handlers.NewPolicyHandler(policyService)
	fleetHandler := handlers.NewFleetHandler(fleetService)
	enrollmentHandler := handlers.NewEnrollmentHandler(enrollmentService)
	firewallHandler := handlers.NewFirewallHandler(firewallSyncService)

	handler := router.New(authHandler, dashboardHandler, policyHandler, fleetHandler, enrollmentHandler, firewallHandler, jwtManager)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("server listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
