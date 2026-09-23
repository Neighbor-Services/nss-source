package gorm

import (
	"context"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type publicRepository struct {
	db *gorm.DB
}

func NewPublicRepository(db *gorm.DB) repository.PublicRepository {
	return &publicRepository{db: db}
}

func (r *publicRepository) GetHeroSection(ctx context.Context) (*entity.HeroSection, error) {
	var hero entity.HeroSection
	err := r.db.WithContext(ctx).Where("is_active = ?", true).First(&hero).Error
	if err != nil {
		return nil, err
	}
	return &hero, nil
}

func (r *publicRepository) ListFeatures(ctx context.Context) ([]entity.Feature, error) {
	var list []entity.Feature
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("order ASC").Find(&list).Error
	return list, err
}

func (r *publicRepository) ListTestimonials(ctx context.Context) ([]entity.Testimonial, error) {
	var list []entity.Testimonial
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *publicRepository) ListFAQs(ctx context.Context) ([]entity.FAQ, error) {
	var list []entity.FAQ
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("order ASC").Find(&list).Error
	return list, err
}

func (r *publicRepository) ListHowItWorksSteps(ctx context.Context) ([]entity.HowItWorksStep, error) {
	var list []entity.HowItWorksStep
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("order ASC").Find(&list).Error
	return list, err
}

func (r *publicRepository) ListSiteStats(ctx context.Context) ([]entity.SiteStat, error) {
	var list []entity.SiteStat
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("order ASC").Find(&list).Error
	return list, err
}

func (r *publicRepository) GetAboutContent(ctx context.Context) (*entity.AboutContent, error) {
	var about entity.AboutContent
	err := r.db.WithContext(ctx).Where("is_active = ?", true).First(&about).Error
	if err != nil {
		return nil, err
	}
	return &about, nil
}

func (r *publicRepository) GetSiteSetting(ctx context.Context) (*entity.SiteSetting, error) {
	var setting entity.SiteSetting
	err := r.db.WithContext(ctx).First(&setting).Error
	if err != nil {
		defaultSetting := entity.SiteSetting{
			ID:             uuid.New(),
			SiteName:       "Neighbor Service",
			ContactEmail:   "support@neighborservice.com",
			ContactPhone:   "+1 (555) 000-0000",
			ContactAddress: "123 Community City, CC 12345",
		}
		_ = r.db.WithContext(ctx).Create(&defaultSetting)
		return &defaultSetting, nil
	}
	return &setting, nil
}

func (r *publicRepository) CreateContactMessage(ctx context.Context, msg *entity.ContactMessage) error {
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *publicRepository) CreateResolutionReport(ctx context.Context, report *entity.ResolutionReport) error {
	if report.ID == uuid.Nil {
		report.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(report).Error
}
