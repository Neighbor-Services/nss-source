package usecase

import (
	"context"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type ModerationUseCase interface {
	SubmitReport(ctx context.Context, reporterID uuid.UUID, reportedUserID *uuid.UUID, contentType, objectID, reason, description string) (*entity.Report, error)
	GetReports(ctx context.Context, status string, limit, offset int) ([]entity.Report, error)
	GetVerifications(ctx context.Context, providerID *uuid.UUID) ([]entity.ProviderVerification, error)
	SubmitVerification(ctx context.Context, providerID uuid.UUID, frontURL, backURL string) (*entity.ProviderVerification, error)
	GetBackgroundChecks(ctx context.Context, providerID *uuid.UUID) ([]entity.BackgroundCheck, error)
	GetBackgroundCheckConfig(ctx context.Context) (map[string]interface{}, error)
	InitiateBackgroundCheck(ctx context.Context, userID uuid.UUID, paymentIntentID string) (*entity.BackgroundCheck, error)
	ResyncBackgroundCheck(ctx context.Context, id uuid.UUID) (*entity.BackgroundCheck, error)
	ProcessCheckrWebhook(ctx context.Context, signature string, rawBody []byte, payload map[string]interface{}) error
}

type AuditUseCase interface {
	LogAction(ctx context.Context, userID *uuid.UUID, action, resourceType, resourceID, ip string, details map[string]interface{}) error
	GetLogs(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]entity.AuditLog, error)
}
