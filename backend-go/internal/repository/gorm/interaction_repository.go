package gorm

import (
	"context"
	"time"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type favoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) repository.FavoriteRepository {
	return &favoriteRepository{db: db}
}

func (r *favoriteRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Favorite, error) {
	var list []entity.Favorite
	err := r.db.WithContext(ctx).
		Preload("FavoriteUser").
		Preload("FavoriteUser.Profile").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *favoriteRepository) Create(ctx context.Context, fav *entity.Favorite) error {
	var existing entity.Favorite
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND favorite_user_id = ?", fav.UserID, fav.FavoriteUserID).
		First(&existing).Error
	if err == nil {
		fav.ID = existing.ID
		return nil
	}
	return r.db.WithContext(ctx).Create(fav).Error
}

func (r *favoriteRepository) Delete(ctx context.Context, userID, favoriteUserID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND (favorite_user_id = ? OR id = ? OR favorite_user_id IN (SELECT user_id FROM profiles WHERE id = ?))", userID, favoriteUserID, favoriteUserID, favoriteUserID).
		Delete(&entity.Favorite{}).Error
}

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) repository.ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) ListByProvider(ctx context.Context, providerID uuid.UUID) ([]entity.Review, error) {
	var list []entity.Review
	err := r.db.WithContext(ctx).
		Preload("Reviewer").
		Preload("Reviewer.Profile").
		Where("provider_id = ?", providerID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *reviewRepository) Create(ctx context.Context, rev *entity.Review) error {
	return r.db.WithContext(ctx).Create(rev).Error
}

func (r *reviewRepository) CalculateProviderRating(ctx context.Context, providerID uuid.UUID) (float64, int, error) {
	type RatingAgg struct {
		AvgRating float64 `gorm:"column:avg_rating"`
		Count     int     `gorm:"column:total_count"`
	}
	var agg RatingAgg
	err := r.db.WithContext(ctx).
		Model(&entity.Review{}).
		Select("AVG(rating) as avg_rating, COUNT(id) as total_count").
		Where("provider_id = ?", providerID).
		Scan(&agg).Error
	return agg.AvgRating, agg.Count, err
}

type appointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) repository.AppointmentRepository {
	return &appointmentRepository{db: db}
}

func (r *appointmentRepository) List(ctx context.Context, seekerID *uuid.UUID, providerID *uuid.UUID, status string) ([]entity.Appointment, error) {
	var list []entity.Appointment
	query := r.db.WithContext(ctx).
		Preload("Seeker").
		Preload("Seeker.Profile").
		Preload("Provider").
		Preload("Provider.Profile").
		Preload("ServiceRequest").
		Preload("Proposal").
		Order("created_at DESC")

	if seekerID != nil && providerID != nil {
		query = query.Where("seeker_id = ? OR provider_id = ? OR seeker_id IN (SELECT id FROM accounts_profile WHERE user_id = ?) OR provider_id IN (SELECT id FROM accounts_profile WHERE user_id = ?)", *seekerID, *providerID, *seekerID, *providerID)
	} else if seekerID != nil {
		query = query.Where("seeker_id = ? OR seeker_id IN (SELECT id FROM accounts_profile WHERE user_id = ?)", *seekerID, *seekerID)
	} else if providerID != nil {
		query = query.Where("provider_id = ? OR provider_id IN (SELECT id FROM accounts_profile WHERE user_id = ?)", *providerID, *providerID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&list).Error
	return list, err
}

func (r *appointmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Appointment, error) {
	var apt entity.Appointment
	err := r.db.WithContext(ctx).
		Preload("Seeker").
		Preload("Seeker.Profile").
		Preload("Provider").
		Preload("Provider.Profile").
		Preload("ServiceRequest").
		Preload("Proposal").
		First(&apt, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &apt, nil
}

func (r *appointmentRepository) Create(ctx context.Context, apt *entity.Appointment) error {
	return r.db.WithContext(ctx).Create(apt).Error
}

func (r *appointmentRepository) Update(ctx context.Context, apt *entity.Appointment) error {
	return r.db.WithContext(ctx).Save(apt).Error
}

func (r *appointmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Appointment{}, "id = ?", id).Error
}

func (r *appointmentRepository) CheckConflict(ctx context.Context, providerID uuid.UUID, scheduledTime time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Appointment{}).
		Where("provider_id = ? AND appointment_date = ? AND status IN ?", providerID, scheduledTime, []string{"SCHEDULED", "IN_PROGRESS"}).
		Count(&count).Error
	return count > 0, err
}

type disputeRepository struct {
	db *gorm.DB
}

func NewDisputeRepository(db *gorm.DB) repository.DisputeRepository {
	return &disputeRepository{db: db}
}

func (r *disputeRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.Dispute, error) {
	var list []entity.Dispute
	err := r.db.WithContext(ctx).
		Preload("RaisedBy").
		Preload("Defendant").
		Preload("Appointment").
		Where("raised_by_id = ? OR defendant_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *disputeRepository) Create(ctx context.Context, dispute *entity.Dispute) error {
	return r.db.WithContext(ctx).Create(dispute).Error
}

func (r *disputeRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Dispute, error) {
	var d entity.Dispute
	err := r.db.WithContext(ctx).
		Preload("RaisedBy").
		Preload("Defendant").
		Preload("Appointment").
		First(&d, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *disputeRepository) Update(ctx context.Context, dispute *entity.Dispute) error {
	return r.db.WithContext(ctx).Save(dispute).Error
}
