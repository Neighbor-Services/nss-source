package repository

import (
	"context"

	"backend-go/internal/domain/entity"
)

type PublicRepository interface {
	GetHeroSection(ctx context.Context) (*entity.HeroSection, error)
	ListFeatures(ctx context.Context) ([]entity.Feature, error)
	ListTestimonials(ctx context.Context) ([]entity.Testimonial, error)
	ListFAQs(ctx context.Context) ([]entity.FAQ, error)
	ListHowItWorksSteps(ctx context.Context) ([]entity.HowItWorksStep, error)
	ListSiteStats(ctx context.Context) ([]entity.SiteStat, error)
	GetAboutContent(ctx context.Context) (*entity.AboutContent, error)
	GetSiteSetting(ctx context.Context) (*entity.SiteSetting, error)
	CreateContactMessage(ctx context.Context, msg *entity.ContactMessage) error
	CreateResolutionReport(ctx context.Context, report *entity.ResolutionReport) error
}
