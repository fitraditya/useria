package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appmw "github.com/fitraditya/useria/internal/middleware"
)

// newTenantRouter serves the customer-facing app: self-signup, company
// team management, own profile. It never mounts /api/admin/* — that only
// exists on the admin process (see router_admin.go), so a bug in a scope
// check here can't expose platform-wide data even in principle.
func newTenantRouter(d *deps) chi.Router {
	r := newBaseRouter()
	jwtSecret := d.cfg.JWTSecret

	r.Route("/api", func(api chi.Router) {
		api.Post("/auth/register", d.authHandler.Register)
		api.Post("/auth/login", d.authHandler.Login)
		api.Post("/auth/forgot-password", d.authHandler.ForgotPassword)
		api.Post("/auth/reset-password", d.authHandler.ResetPassword)

		api.Group(func(pr chi.Router) {
			pr.Use(appmw.Auth(jwtSecret))
			pr.Get("/auth/select-company", d.authHandler.ListCompanies)
			pr.Post("/auth/select-company", d.authHandler.SelectCompany)
			pr.Post("/auth/refresh", d.authHandler.Refresh)
			pr.Get("/profile", d.profileHandler.Get)
			pr.Put("/profile", d.profileHandler.Update)
			pr.Get("/invitations", d.invitationHandler.List)
			pr.Post("/invitations/accept", d.invitationHandler.Accept)
		})

		// Company admin: team member management, scoped to the token's selected company.
		api.Group(func(pr chi.Router) {
			pr.Use(appmw.Auth(jwtSecret))
			pr.Use(appmw.RequireCompany)
			pr.With(appmw.RequireScope("members:read")).Get("/company/members", d.memberHandler.List)
			pr.With(appmw.RequireScope("members:create")).Post("/company/members/invite", d.memberHandler.Invite)
			pr.With(appmw.RequireScope("members:read")).Get("/company/members/{id}", d.memberHandler.Get)
			pr.With(appmw.RequireScope("members:write")).Put("/company/members/{id}", d.memberHandler.UpdateRole)
			pr.With(appmw.RequireScope("members:delete")).Delete("/company/members/{id}", d.memberHandler.Remove)
			pr.With(appmw.RequireScope("members:read")).Get("/company/roles", d.roleHandler.ListAssignable)
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusFound)
	})
	r.Get("/login", d.viewHandler.Login("tenant"))
	r.Get("/register", d.viewHandler.Register())
	r.Get("/select-company", d.viewHandler.SelectCompany())
	r.Get("/forgot-password", d.viewHandler.ForgotPassword("tenant"))
	r.Get("/reset-password", d.viewHandler.ResetPassword("tenant"))
	r.Get("/dashboard", d.viewHandler.Dashboard())
	r.Get("/profile", d.viewHandler.Profile("tenant"))
	r.Get("/company/members", d.viewHandler.CompanyMembers())

	return r
}
