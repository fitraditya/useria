package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appmw "github.com/fitraditya/useria/internal/middleware"
)

// newAdminRouter serves the platform (SuperAdmin) app: cross-company
// management, nothing else. No self-signup, no company team management —
// SuperAdmin accounts only come from the seed-admin CLI command, and this
// process never mounts the tenant company/members endpoints, so it has
// nothing to leak even if a scope check elsewhere were wrong.
func newAdminRouter(d *deps) chi.Router {
	r := newBaseRouter()
	jwtSecret := d.cfg.JWTSecret

	r.Route("/api", func(api chi.Router) {
		api.Post("/auth/login", d.authHandler.Login)
		api.Post("/auth/forgot-password", d.authHandler.ForgotPassword)
		api.Post("/auth/reset-password", d.authHandler.ResetPassword)

		api.Group(func(pr chi.Router) {
			pr.Use(appmw.Auth(jwtSecret))
			pr.Post("/auth/refresh", d.authHandler.Refresh)
			pr.Get("/profile", d.profileHandler.Get)
			pr.Put("/profile", d.profileHandler.Update)
		})

		api.Group(func(pr chi.Router) {
			pr.Use(appmw.Auth(jwtSecret))
			pr.With(appmw.RequireScope("companies:read")).Get("/admin/companies", d.companyHandler.List)
			pr.With(appmw.RequireScope("companies:create")).Post("/admin/companies", d.companyHandler.Create)
			pr.With(appmw.RequireScope("companies:read")).Get("/admin/companies/{id}", d.companyHandler.Get)
			pr.With(appmw.RequireScope("companies:write")).Put("/admin/companies/{id}", d.companyHandler.Update)
			pr.With(appmw.RequireScope("companies:delete")).Delete("/admin/companies/{id}", d.companyHandler.Delete)
			pr.With(appmw.RequireScope("companies:read")).Get("/admin/stats", d.adminHandler.Stats)
			pr.With(appmw.RequireScope("users:read")).Get("/admin/users", d.adminHandler.Users)
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusFound)
	})
	r.Get("/login", d.viewHandler.Login("admin"))
	r.Get("/forgot-password", d.viewHandler.ForgotPassword("admin"))
	r.Get("/reset-password", d.viewHandler.ResetPassword("admin"))
	r.Get("/profile", d.viewHandler.Profile("admin"))
	r.Get("/admin/dashboard", d.viewHandler.AdminDashboard())
	r.Get("/admin/companies", d.viewHandler.Companies())
	r.Get("/admin/users", d.viewHandler.AdminUsers())

	return r
}
