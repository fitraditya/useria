package handler

import (
	"html/template"
	"net/http"
)

// ViewHandler serves the static HTML shell for each frontend page. Pages
// carry no server-rendered data — all of it is fetched client-side via the
// JSON API using the JWT stored in localStorage, so these templates only
// need a page title and the active nav item. The same ViewHandler backs
// both the tenant and admin processes; App picks which sidebar/breadcrumb
// nav set a dashboard-family page renders (they run on separate ports and
// never share a router, but there's no reason to fork the templating code).
type ViewHandler struct {
	authBase      *template.Template
	dashboardBase *template.Template
}

func NewViewHandler() (*ViewHandler, error) {
	authBase, err := template.ParseFiles(
		"templates/partials/head.html",
		"templates/partials/preloader.html",
		"templates/base_auth.html",
	)
	if err != nil {
		return nil, err
	}
	dashboardBase, err := template.ParseFiles(
		"templates/partials/head.html",
		"templates/partials/preloader.html",
		"templates/partials/overlay.html",
		"templates/partials/sidebar.html",
		"templates/partials/header.html",
		"templates/partials/breadcrumb.html",
		"templates/base_dashboard.html",
	)
	if err != nil {
		return nil, err
	}
	return &ViewHandler{authBase: authBase, dashboardBase: dashboardBase}, nil
}

type pageData struct {
	Title   string
	PageKey string
	App     string // "tenant" or "admin"; picks the sidebar/breadcrumb nav set
}

func (v *ViewHandler) renderAuth(pageFile, title, app string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := v.authBase.Clone()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if t, err = t.ParseFiles(pageFile); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := t.ExecuteTemplate(w, "base", pageData{Title: title, App: app}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (v *ViewHandler) renderDashboard(pageFile, title, pageKey, app string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := v.dashboardBase.Clone()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if t, err = t.ParseFiles(pageFile); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := t.ExecuteTemplate(w, "base", pageData{Title: title, PageKey: pageKey, App: app}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (v *ViewHandler) Login(app string) http.HandlerFunc {
	return v.renderAuth("templates/auth/login.html", "Sign in", app)
}

func (v *ViewHandler) Register() http.HandlerFunc {
	return v.renderAuth("templates/auth/register.html", "Register", "tenant")
}

func (v *ViewHandler) SelectCompany() http.HandlerFunc {
	return v.renderAuth("templates/auth/select-company.html", "Select Company", "tenant")
}

func (v *ViewHandler) ForgotPassword(app string) http.HandlerFunc {
	return v.renderAuth("templates/auth/forgot-password.html", "Forgot Password", app)
}

func (v *ViewHandler) ResetPassword(app string) http.HandlerFunc {
	return v.renderAuth("templates/auth/reset-password.html", "Reset Password", app)
}

func (v *ViewHandler) Dashboard() http.HandlerFunc {
	return v.renderDashboard("templates/dashboard/index.html", "Dashboard", "dashboard", "tenant")
}

func (v *ViewHandler) Profile(app string) http.HandlerFunc {
	return v.renderDashboard("templates/profile/edit.html", "Profile", "profile", app)
}

func (v *ViewHandler) CompanyMembers() http.HandlerFunc {
	return v.renderDashboard("templates/company/members.html", "Team Members", "members", "tenant")
}

func (v *ViewHandler) Companies() http.HandlerFunc {
	return v.renderDashboard("templates/admin/companies.html", "Companies", "companies", "admin")
}
