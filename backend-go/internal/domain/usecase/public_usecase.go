package usecase

import (
	"context"

	"backend-go/internal/domain/entity"
)

type PublicUseCase interface {
	GetCMSContent(ctx context.Context) (map[string]interface{}, error)
	SubmitContactMessage(ctx context.Context, msg *entity.ContactMessage) error
	SubmitResolutionReport(ctx context.Context, report *entity.ResolutionReport) error
}
