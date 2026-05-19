package router

import (
	"net/http"

	"firewall-manager/internal/auth"
	"firewall-manager/internal/http/handlers"
	"firewall-manager/internal/http/middleware"
	"firewall-manager/internal/web"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func New(authHandler *handlers.AuthHandler, dashboardHandler *handlers.DashboardHandler, policyHandler *handlers.PolicyHandler, fleetHandler *handlers.FleetHandler, enrollmentHandler *handlers.EnrollmentHandler, firewallHandler *handlers.FirewallHandler, jwtManager *auth.JWTManager) http.Handler {
	r := chi.NewRouter()
	webApp := web.New()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusFound)
	})
	r.Get("/login", webApp.LoginPage)
	r.Get("/admin", webApp.AdminPage)
	r.Get("/enroll", webApp.EnrollPage)
	r.Get("/static/*", webApp.Static)

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/public/enroll/accept", enrollmentHandler.AcceptPublic)
		api.Post("/public/agent/commands/next", fleetHandler.AgentClaimNextCommand)
		api.Post("/public/agent/commands/{commandID}/result", fleetHandler.AgentReportCommandResult)

		api.Route("/auth", func(authRoutes chi.Router) {
			authRoutes.Post("/bootstrap-admin", authHandler.BootstrapAdmin)
			authRoutes.Post("/login", authHandler.Login)

			authRoutes.Group(func(protected chi.Router) {
				protected.Use(middleware.RequireAuth(jwtManager))
				protected.Get("/me", authHandler.Me)
			})
		})

		api.Group(func(protected chi.Router) {
			protected.Use(middleware.RequireAuth(jwtManager))
			protected.Get("/dashboard/summary", dashboardHandler.Summary)
			protected.Post("/firewall/sync", firewallHandler.Sync)
			protected.Post("/enrollment-links", enrollmentHandler.CreateLink)
			protected.Get("/enrollments", enrollmentHandler.ListEnrollments)
			protected.Post("/enrollments/{enrollmentID}/approve", enrollmentHandler.Approve)
			protected.Post("/enrollments/{enrollmentID}/disable", enrollmentHandler.Disable)
			protected.Get("/notifications", enrollmentHandler.ListNotifications)
			protected.Get("/departments", fleetHandler.ListDepartments)
			protected.Post("/departments", fleetHandler.CreateDepartment)
			protected.Get("/laptops", fleetHandler.ListLaptops)
			protected.Post("/laptops", fleetHandler.CreateLaptop)
			protected.Post("/laptops/{laptopID}/usb/block", fleetHandler.QueueUSBBlock)
			protected.Post("/laptops/{laptopID}/usb/unblock", fleetHandler.QueueUSBUnblock)
			protected.Post("/policy-assignments", fleetHandler.CreatePolicyAssignment)
			protected.Get("/policies/{policyID}/assignments", fleetHandler.ListPolicyAssignments)
			protected.Get("/policies", policyHandler.List)
			protected.Post("/policies", policyHandler.Create)
			protected.Put("/policies/{policyID}", policyHandler.Update)
			protected.Delete("/policies/{policyID}", policyHandler.Delete)
		})
	})

	return r
}
