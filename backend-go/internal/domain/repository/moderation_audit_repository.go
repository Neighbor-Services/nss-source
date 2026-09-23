package repository

import (
	"context"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
	List(ctx context.Context, userID *uuid.UUID, limit int, offset int) ([]entity.AuditLog, error)
}

type ReportRepository interface {
	Create(ctx context.Context, report *entity.Report) error
	List(ctx context.Context, status string, limit int, offset int) ([]entity.Report, error)
}

type BackgroundCheckRepository interface {
	Create(ctx context.Context, check *entity.BackgroundCheck) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.BackgroundCheck, error)
	GetByCandidateID(ctx context.Context, candidateID string) (*entity.BackgroundCheck, error)
	GetByReportID(ctx context.Context, reportID string) (*entity.BackgroundCheck, error)
	GetPendingByProvider(ctx context.Context, providerID uuid.UUID) (*entity.BackgroundCheck, error)
	List(ctx context.Context, providerID *uuid.UUID) ([]entity.BackgroundCheck, error)
	Update(ctx context.Context, check *entity.BackgroundCheck) error
	GetModerationSetting(ctx context.Context) (*entity.ModerationSetting, error)
}

type ProviderVerificationRepository interface {
	Create(ctx context.Context, v *entity.ProviderVerification) error
	List(ctx context.Context, providerID *uuid.UUID) ([]entity.ProviderVerification, error)
	Update(ctx context.Context, v *entity.ProviderVerification) error
}
