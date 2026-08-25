package service

import (
	"context"
	"log"

	"github.com/fitraditya/useria/internal/models"
	"github.com/fitraditya/useria/internal/repository"
	"github.com/fitraditya/useria/internal/utils"
)

// AuditService records sensitive mutations (company create/update/suspend/
// delete, member invite/role-change/remove) for accountability. Under
// MySQL it persists to audit_logs; under sqlite (dev) it just logs to
// stdout, since sqlite dev databases get thrown away constantly and don't
// warrant a real audit trail.
type AuditService interface {
	Log(ctx context.Context, actorUserID, action, resourceType, resourceID, companyID, metadata string)
}

// NewAuditService picks the implementation based on the configured DB
// driver. companyID and metadata may be empty.
func NewAuditService(driver string, repo *repository.AuditRepository) AuditService {
	if driver == "mysql" {
		return &dbAuditService{repo: repo}
	}
	return &logAuditService{}
}

type dbAuditService struct {
	repo *repository.AuditRepository
}

func (s *dbAuditService) Log(ctx context.Context, actorUserID, action, resourceType, resourceID, companyID, metadata string) {
	entry := &models.AuditLog{
		ID:           utils.NewUUID(),
		ActorUserID:  actorUserID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}
	if companyID != "" {
		entry.CompanyID = &companyID
	}
	if metadata != "" {
		entry.Metadata = &metadata
	}
	// Audit logging must never break the request it's observing — log and
	// move on if the write fails.
	if err := s.repo.Create(ctx, entry); err != nil {
		log.Printf("audit log write failed (%s %s/%s by %s): %v", action, resourceType, resourceID, actorUserID, err)
	}
}

type logAuditService struct{}

func (s *logAuditService) Log(ctx context.Context, actorUserID, action, resourceType, resourceID, companyID, metadata string) {
	log.Printf("[audit] actor=%s action=%s resource=%s/%s company=%s metadata=%s", actorUserID, action, resourceType, resourceID, companyID, metadata)
}
