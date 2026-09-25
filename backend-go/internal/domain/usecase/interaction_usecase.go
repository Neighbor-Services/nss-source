package usecase

import (
	"context"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type InteractionUseCase interface {
	GetFavorites(ctx context.Context, userID uuid.UUID) ([]entity.Favorite, error)
	CreateFavorite(ctx context.Context, userID, providerID uuid.UUID) (*entity.Favorite, error)
	DeleteFavorite(ctx context.Context, userID, providerID uuid.UUID) error

	GetReviews(ctx context.Context, providerID uuid.UUID) ([]entity.Review, error)
	CreateReview(ctx context.Context, authorID uuid.UUID, review *entity.Review) (*entity.Review, error)

	GetAppointments(ctx context.Context, userID uuid.UUID, userType, status string) ([]entity.Appointment, error)
	CreateAppointment(ctx context.Context, customerID uuid.UUID, apt *entity.Appointment) (*entity.Appointment, error)
	VerifyArrivalCode(ctx context.Context, providerID, appointmentID uuid.UUID, code string) (*entity.Appointment, error)
	NotifyOnTheWay(ctx context.Context, providerID, appointmentID uuid.UUID) (*entity.Appointment, error)
	CompleteAppointment(ctx context.Context, userID, appointmentID uuid.UUID, amount float64) (fundsReleased float64, commission float64, err error)
	CancelAppointment(ctx context.Context, userID, appointmentID uuid.UUID) error

	GetDisputes(ctx context.Context, userID uuid.UUID) ([]entity.Dispute, error)
	CreateDispute(ctx context.Context, initiatorID uuid.UUID, dispute *entity.Dispute) (*entity.Dispute, error)
	UploadDisputeEvidence(ctx context.Context, userID, disputeID uuid.UUID, evidenceURL string) error
}
