package repository

import (
	"context"
	"time"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type FavoriteRepository interface {
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Favorite, error)
	Create(ctx context.Context, favorite *entity.Favorite) error
	Delete(ctx context.Context, userID uuid.UUID, providerID uuid.UUID) error
}

type ReviewRepository interface {
	ListByProvider(ctx context.Context, providerID uuid.UUID) ([]entity.Review, error)
	Create(ctx context.Context, review *entity.Review) error
	CalculateProviderRating(ctx context.Context, providerID uuid.UUID) (avgRating float64, totalReviews int, err error)
}

type AppointmentRepository interface {
	List(ctx context.Context, seekerID *uuid.UUID, providerID *uuid.UUID, status string) ([]entity.Appointment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Appointment, error)
	Create(ctx context.Context, apt *entity.Appointment) error
	Update(ctx context.Context, apt *entity.Appointment) error
	Delete(ctx context.Context, id uuid.UUID) error
	CheckConflict(ctx context.Context, providerID uuid.UUID, scheduledTime time.Time) (bool, error)
}

type DisputeRepository interface {
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Dispute, error)
	Create(ctx context.Context, dispute *entity.Dispute) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Dispute, error)
	Update(ctx context.Context, dispute *entity.Dispute) error
}
