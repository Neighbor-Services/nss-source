package usecase

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"github.com/google/uuid"
)

type profileUseCase struct {
	profileRepo   repository.ProfileRepository
	aboutRepo     repository.AboutRepository
	portfolioRepo repository.PortfolioRepository
	pkgRepo       repository.ServicePackageRepository
	legalRepo     repository.LegalDocumentRepository
}

func NewProfileUseCase(
	profileRepo repository.ProfileRepository,
	aboutRepo repository.AboutRepository,
	portfolioRepo repository.PortfolioRepository,
	pkgRepo repository.ServicePackageRepository,
	legalRepo repository.LegalDocumentRepository,
) domainUsecase.ProfileUseCase {
	return &profileUseCase{
		profileRepo:   profileRepo,
		aboutRepo:     aboutRepo,
		portfolioRepo: portfolioRepo,
		pkgRepo:       pkgRepo,
		legalRepo:     legalRepo,
	}
}

func (u *profileUseCase) GetProfile(ctx context.Context, userID uuid.UUID, targetID *uuid.UUID) (*entity.Profile, error) {
	if targetID != nil {
		return u.profileRepo.GetByID(ctx, *targetID)
	}
	profile, err := u.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile != nil && profile.RecordActivity() {
		_ = u.profileRepo.Update(ctx, profile)
	}
	return profile, nil
}

func (u *profileUseCase) ListProfiles(ctx context.Context, params repository.ProfileFilterParams) ([]entity.Profile, error) {
	return u.profileRepo.List(ctx, params)
}

func (u *profileUseCase) UpdateProfile(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*entity.Profile, error) {
	profile, err := u.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		profile = &entity.Profile{
			ID:       uuid.New(),
			UserID:   userID,
			UserType: "SEEKER",
		}
		_ = u.profileRepo.Create(ctx, profile)
	}

	if fn, ok := updates["first_name"].(string); ok && fn != "" {
		profile.FirstName = strings.TrimSpace(fn)
	}
	if ln, ok := updates["last_name"].(string); ok && ln != "" {
		profile.LastName = strings.TrimSpace(ln)
	}
	if phone, ok := updates["phone"].(string); ok {
		profile.Phone = strings.TrimSpace(phone)
	}
	if dobStr, ok := updates["date_of_birth"].(string); ok && dobStr != "" {
		if parsed, err := time.Parse("2006-01-02", dobStr); err == nil {
			profile.DateOfBirth = &parsed
		} else if parsed, err := time.Parse(time.RFC3339, dobStr); err == nil {
			profile.DateOfBirth = &parsed
		}
	}
	if city, ok := updates["city"].(string); ok {
		profile.City = strings.TrimSpace(city)
	}
	if state, ok := updates["state"].(string); ok {
		profile.State = strings.TrimSpace(state)
	}
	if zip, ok := updates["zip_code"].(string); ok {
		profile.ZipCode = strings.TrimSpace(zip)
	}
	if country, ok := updates["country"].(string); ok {
		profile.Country = strings.TrimSpace(country)
	}
	if cc, ok := updates["country_code"].(string); ok {
		profile.CountryCode = strings.TrimSpace(cc)
	}
	if gender, ok := updates["gender"].(string); ok {
		profile.Gender = strings.ToUpper(strings.TrimSpace(gender))
	}
	if addr, ok := updates["address"].(string); ok {
		profile.Address = strings.TrimSpace(addr)
	}
	if bio, ok := updates["bio"].(string); ok {
		profile.Bio = strings.TrimSpace(bio)
	}
	if srv, ok := updates["service"].(string); ok {
		profile.Service = strings.TrimSpace(srv)
	}
	if ppm, ok := updates["preferred_payment_mode"].(string); ok && ppm != "" {
		profile.PreferredPaymentMode = ppm
	}
	if tier, ok := updates["subscription_tier"].(string); ok && tier != "" {
		profile.SubscriptionTier = tier
	}
	if interval, ok := updates["subscription_interval"].(string); ok && interval != "" {
		profile.SubscriptionInterval = interval
	}
	if dt, ok := updates["device_token"].(string); ok {
		profile.DeviceToken = dt
	}
	if lat, ok := updates["latitude"].(float64); ok {
		profile.Latitude = lat
	} else if latStr, ok := updates["latitude"].(string); ok && latStr != "" {
		if parsed, err := strconv.ParseFloat(latStr, 64); err == nil {
			profile.Latitude = parsed
		}
	}
	if lon, ok := updates["longitude"].(float64); ok {
		profile.Longitude = lon
	} else if lonStr, ok := updates["longitude"].(string); ok && lonStr != "" {
		if parsed, err := strconv.ParseFloat(lonStr, 64); err == nil {
			profile.Longitude = parsed
		}
	}
	if ut, ok := updates["user_type"].(string); ok && ut != "" {
		profile.UserType = strings.ToUpper(strings.TrimSpace(ut))
	}

	profile.UpdatedAt = time.Now()
	if err := u.profileRepo.Update(ctx, profile); err != nil {
		return nil, err
	}

	// Handle catalog services update if provided
	var serviceIDs []uuid.UUID
	if rawList, ok := updates["catalog_services"].([]interface{}); ok {
		for _, item := range rawList {
			if idStr, isStr := item.(string); isStr {
				if parsed, parseErr := uuid.Parse(idStr); parseErr == nil {
					serviceIDs = append(serviceIDs, parsed)
				}
			} else if idMap, isMap := item.(map[string]interface{}); isMap {
				if idVal, hasID := idMap["id"].(string); hasID {
					if parsed, parseErr := uuid.Parse(idVal); parseErr == nil {
						serviceIDs = append(serviceIDs, parsed)
					}
				}
			}
		}
		_ = u.profileRepo.UpdateCatalogServices(ctx, profile.ID, serviceIDs)
	} else if rawIDs, ok := updates["catalog_service_ids"].([]interface{}); ok {
		for _, item := range rawIDs {
			if idStr, isStr := item.(string); isStr {
				if parsed, parseErr := uuid.Parse(idStr); parseErr == nil {
					serviceIDs = append(serviceIDs, parsed)
				}
			}
		}
		_ = u.profileRepo.UpdateCatalogServices(ctx, profile.ID, serviceIDs)
	} else if rawSingle, ok := updates["catalog_service"].(string); ok && rawSingle != "" {
		if parsed, parseErr := uuid.Parse(rawSingle); parseErr == nil {
			serviceIDs = append(serviceIDs, parsed)
			_ = u.profileRepo.UpdateCatalogServices(ctx, profile.ID, serviceIDs)
		}
	}

	return u.profileRepo.GetByID(ctx, profile.ID)
}

func (u *profileUseCase) GetAbout(ctx context.Context, userID uuid.UUID) (*entity.About, error) {
	return u.aboutRepo.GetByUserID(ctx, userID)
}

func (u *profileUseCase) UpdateAbout(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*entity.About, error) {
	about, err := u.aboutRepo.GetByUserID(ctx, userID)
	if err != nil || about == nil {
		about = &entity.About{
			ID:     uuid.New(),
			UserID: userID,
		}
	}

	if name, ok := updates["name"].(string); ok && name != "" {
		about.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		about.Description = desc
	}
	if spec, ok := updates["specification"].(string); ok {
		about.Specification = spec
	}
	if exp, ok := updates["experience_years"].(float64); ok {
		about.ExperienceYears = int(exp)
	}
	if addr, ok := updates["address"].(string); ok {
		about.Address = addr
	}
	if edu, ok := updates["education"].(string); ok {
		about.Education = edu
	}
	if skills, ok := updates["skills"].([]string); ok {
		about.Skills = skills
	}
	if langs, ok := updates["languages"].([]string); ok {
		about.Languages = langs
	}

	about.UpdatedAt = time.Now()
	if err := u.aboutRepo.Upsert(ctx, about); err != nil {
		return nil, err
	}

	return about, nil
}

func (u *profileUseCase) UpdateProfilePicture(ctx context.Context, userID uuid.UUID, imageURL string) (*entity.Profile, error) {
	profile, err := u.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		profile = &entity.Profile{
			ID:       uuid.New(),
			UserID:   userID,
			UserType: "SEEKER",
		}
		_ = u.profileRepo.Create(ctx, profile)
	}

	profile.ProfilePicture = imageURL
	profile.UpdatedAt = time.Now()
	if err := u.profileRepo.Update(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (u *profileUseCase) GetPortfolios(ctx context.Context, userID uuid.UUID) ([]entity.Portfolio, error) {
	profile, err := u.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("profile not found")
	}
	return u.portfolioRepo.ListByProfileID(ctx, profile.ID)
}

func (u *profileUseCase) CreatePortfolio(ctx context.Context, userID uuid.UUID, item *entity.Portfolio) (*entity.Portfolio, error) {
	profile, err := u.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("profile not found")
	}
	item.ID = uuid.New()
	item.ProfileID = profile.ID
	item.CreatedAt = time.Now()
	if err := u.portfolioRepo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (u *profileUseCase) DeletePortfolio(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) error {
	profile, err := u.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return errors.New("profile not found")
	}
	return u.portfolioRepo.Delete(ctx, itemID, profile.ID)
}

func (u *profileUseCase) GetServicePackages(ctx context.Context, userID uuid.UUID) ([]entity.ServicePackage, error) {
	profile, err := u.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("profile not found")
	}
	return u.pkgRepo.ListByProfileID(ctx, profile.ID)
}

func (u *profileUseCase) CreateServicePackage(ctx context.Context, userID uuid.UUID, pkg *entity.ServicePackage) (*entity.ServicePackage, error) {
	profile, err := u.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("profile not found")
	}

	pkg.ID = uuid.New()
	pkg.ProfileID = profile.ID
	if err := u.pkgRepo.Create(ctx, pkg); err != nil {
		return nil, err
	}
	return pkg, nil
}

func (u *profileUseCase) DeleteServicePackage(ctx context.Context, userID uuid.UUID, pkgID uuid.UUID) error {
	profile, err := u.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return errors.New("profile not found")
	}
	return u.pkgRepo.Delete(ctx, pkgID, profile.ID)
}

func (u *profileUseCase) GetLegalDocuments(ctx context.Context, docType string) ([]entity.LegalDocument, error) {
	if docType != "" {
		return u.legalRepo.GetByType(ctx, strings.ToUpper(docType))
	}
	return u.legalRepo.ListActive(ctx)
}
