package gorm

import (
	"context"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) repository.AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditLogRepository) List(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]entity.AuditLog, error) {
	var list []entity.AuditLog
	query := r.db.WithContext(ctx).Order("created_at DESC")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&list).Error
	return list, err
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) repository.ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) Create(ctx context.Context, report *entity.Report) error {
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *reportRepository) List(ctx context.Context, status string, limit, offset int) ([]entity.Report, error) {
	var list []entity.Report
	query := r.db.WithContext(ctx).Order("created_at DESC")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&list).Error
	return list, err
}

type bgCheckRepository struct {
	db *gorm.DB
}

func NewBackgroundCheckRepository(db *gorm.DB) repository.BackgroundCheckRepository {
	return &bgCheckRepository{db: db}
}

func (r *bgCheckRepository) Create(ctx context.Context, check *entity.BackgroundCheck) error {
	return r.db.WithContext(ctx).Create(check).Error
}

func (r *bgCheckRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.BackgroundCheck, error) {
	var check entity.BackgroundCheck
	err := r.db.WithContext(ctx).Preload("Provider").First(&check, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &check, nil
}

func (r *bgCheckRepository) GetByCandidateID(ctx context.Context, candidateID string) (*entity.BackgroundCheck, error) {
	var check entity.BackgroundCheck
	err := r.db.WithContext(ctx).Preload("Provider").First(&check, "checkr_candidate_id = ?", candidateID).Error
	if err != nil {
		return nil, err
	}
	return &check, nil
}

func (r *bgCheckRepository) GetByReportID(ctx context.Context, reportID string) (*entity.BackgroundCheck, error) {
	var check entity.BackgroundCheck
	err := r.db.WithContext(ctx).Preload("Provider").First(&check, "checkr_report_id = ?", reportID).Error
	if err != nil {
		return nil, err
	}
	return &check, nil
}

func (r *bgCheckRepository) GetPendingByProvider(ctx context.Context, providerID uuid.UUID) (*entity.BackgroundCheck, error) {
	var check entity.BackgroundCheck
	err := r.db.WithContext(ctx).Preload("Provider").First(&check, "provider_id = ? AND status = ?", providerID, "pending").Error
	if err != nil {
		return nil, err
	}
	return &check, nil
}

func (r *bgCheckRepository) List(ctx context.Context, providerID *uuid.UUID) ([]entity.BackgroundCheck, error) {
	var list []entity.BackgroundCheck
	query := r.db.WithContext(ctx).Preload("Provider").Order("created_at desc")
	if providerID != nil {
		query = query.Where("provider_id = ?", *providerID)
	}
	err := query.Find(&list).Error
	return list, err
}

func (r *bgCheckRepository) Update(ctx context.Context, check *entity.BackgroundCheck) error {
	return r.db.WithContext(ctx).Save(check).Error
}

func (r *bgCheckRepository) GetModerationSetting(ctx context.Context) (*entity.ModerationSetting, error) {
	var setting entity.ModerationSetting
	err := r.db.WithContext(ctx).First(&setting).Error
	if err != nil {
		defaultSetting := entity.ModerationSetting{
			ID:                         uuid.New(),
			BackgroundCheckPaymentMode: "IN_APP_STRIPE",
			BackgroundCheckFee:         29.99,
		}
		_ = r.db.WithContext(ctx).Create(&defaultSetting)
		return &defaultSetting, nil
	}
	return &setting, nil
}

type providerVerificationRepository struct {
	db *gorm.DB
}

func NewProviderVerificationRepository(db *gorm.DB) repository.ProviderVerificationRepository {
	return &providerVerificationRepository{db: db}
}

func (r *providerVerificationRepository) Create(ctx context.Context, v *entity.ProviderVerification) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *providerVerificationRepository) List(ctx context.Context, providerID *uuid.UUID) ([]entity.ProviderVerification, error) {
	var list []entity.ProviderVerification
	query := r.db.WithContext(ctx).Order("created_at desc")
	if providerID != nil {
		query = query.Where("provider_id = ?", *providerID)
	}
	err := query.Find(&list).Error
	return list, err
}

func (r *providerVerificationRepository) Update(ctx context.Context, v *entity.ProviderVerification) error {
	return r.db.WithContext(ctx).Save(v).Error
}
