package main

import (
	"database/sql"
	"log"

	"github.com/fitraditya/useria/internal/config"
	"github.com/fitraditya/useria/internal/handler"
	"github.com/fitraditya/useria/internal/repository"
	"github.com/fitraditya/useria/internal/service"
)

// deps bundles everything either router needs. Built once per process;
// the tenant and admin binaries each wire only the handlers their own
// router mounts — see router_tenant.go and router_admin.go.
type deps struct {
	cfg *config.Config

	authHandler       *handler.AuthHandler
	profileHandler    *handler.ProfileHandler
	companyHandler    *handler.CompanyHandler
	memberHandler     *handler.MemberHandler
	invitationHandler *handler.InvitationHandler
	roleHandler       *handler.RoleHandler
	adminHandler      *handler.AdminHandler
	viewHandler       *handler.ViewHandler
}

func buildDeps(cfg *config.Config, db *sql.DB) *deps {
	userRepo := repository.NewUserRepository(db)
	memberRepo := repository.NewMemberRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	resetRepo := repository.NewPasswordResetRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// audit_logs only exists under the mysql migration — sqlite (dev) just
	// logs these events to stdout instead of persisting them.
	auditService := service.NewAuditService(cfg.DBDriver, auditRepo)

	authService := service.NewAuthService(userRepo, memberRepo, roleRepo, companyRepo, resetRepo, cfg.JWTSecret, cfg.JWTExpiration)
	companyService := service.NewCompanyService(companyRepo, auditService)
	memberService := service.NewMemberService(memberRepo, userRepo, roleRepo, auditService)
	adminService := service.NewAdminService(companyRepo, userRepo, memberRepo)

	viewHandler, err := handler.NewViewHandler()
	if err != nil {
		log.Fatalf("view handler: %v", err)
	}

	return &deps{
		cfg: cfg,

		authHandler:       handler.NewAuthHandler(authService),
		profileHandler:    handler.NewProfileHandler(userRepo),
		companyHandler:    handler.NewCompanyHandler(companyService),
		memberHandler:     handler.NewMemberHandler(memberService),
		invitationHandler: handler.NewInvitationHandler(memberService),
		roleHandler:       handler.NewRoleHandler(roleRepo),
		adminHandler:      handler.NewAdminHandler(adminService),
		viewHandler:       viewHandler,
	}
}
