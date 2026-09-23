package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/pkg/email"
	"github.com/google/uuid"
)

type publicUseCase struct {
	publicRepo   repository.PublicRepository
	catalogRepo  repository.CatalogServiceRepository
	categoryRepo repository.CategoryRepository
	cfg          *config.Config
}

func NewPublicUseCase(
	publicRepo repository.PublicRepository,
	catalogRepo repository.CatalogServiceRepository,
	categoryRepo repository.CategoryRepository,
	cfg *config.Config,
) domainUsecase.PublicUseCase {
	return &publicUseCase{
		publicRepo:   publicRepo,
		catalogRepo:  catalogRepo,
		categoryRepo: categoryRepo,
		cfg:          cfg,
	}
}

func (u *publicUseCase) GetCMSContent(ctx context.Context) (map[string]interface{}, error) {
	hero, _ := u.publicRepo.GetHeroSection(ctx)
	features, _ := u.publicRepo.ListFeatures(ctx)
	testimonials, _ := u.publicRepo.ListTestimonials(ctx)
	faqs, _ := u.publicRepo.ListFAQs(ctx)
	steps, _ := u.publicRepo.ListHowItWorksSteps(ctx)
	stats, _ := u.publicRepo.ListSiteStats(ctx)
	about, _ := u.publicRepo.GetAboutContent(ctx)
	settings, _ := u.publicRepo.GetSiteSetting(ctx)

	categories, _ := u.categoryRepo.ListActive(ctx)
	catalogServices, _ := u.catalogRepo.List(ctx, "", "")

	return map[string]interface{}{
		"hero":             hero,
		"features":         features,
		"testimonials":     testimonials,
		"faqs":             faqs,
		"steps":            steps,
		"stats":            stats,
		"about":            about,
		"settings":         settings,
		"categories":       categories,
		"catalog_services": catalogServices,
	}, nil
}

func (u *publicUseCase) SubmitContactMessage(ctx context.Context, msg *entity.ContactMessage) error {
	if msg.FirstName == "" || msg.Email == "" || msg.Message == "" {
		return errors.New("first_name, email, and message are required")
	}

	msg.ID = uuid.New()
	msg.CreatedAt = time.Now()
	msg.IsResolved = false

	if err := u.publicRepo.CreateContactMessage(ctx, msg); err != nil {
		return err
	}

	// Send automatic confirmation email to submitter
	if u.cfg != nil && msg.Email != "" {
		go func(toEmail, senderName, inquiryType, messageText string) {
			mailCfg := &email.Config{
				Host:     u.cfg.SMTPHost,
				Port:     u.cfg.SMTPPort,
				User:     u.cfg.SMTPUser,
				Password: u.cfg.SMTPPassword,
				From:     u.cfg.EmailFrom,
			}
			snippet := messageText
			if len(snippet) > 200 {
				snippet = snippet[:197] + "..."
			}
			if err := email.SendContactConfirmationEmail(mailCfg, toEmail, senderName, inquiryType, snippet); err != nil {
				log.Printf("⚠️ Failed to send contact confirmation email to %s: %v", toEmail, err)
			} else {
				log.Printf("📧 Automated contact confirmation email sent to %s", toEmail)
			}
		}(msg.Email, msg.FirstName, msg.InquiryType, msg.Message)
	}

	return nil
}

func (u *publicUseCase) SubmitResolutionReport(ctx context.Context, report *entity.ResolutionReport) error {
	if report.Role == "" || report.IssueType == "" || report.Description == "" {
		return errors.New("role, issue_type, and description are required")
	}

	report.ID = uuid.New()
	report.CreatedAt = time.Now()
	report.IsReviewed = false

	return u.publicRepo.CreateResolutionReport(ctx, report)
}
