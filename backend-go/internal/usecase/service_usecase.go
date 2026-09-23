package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/pkg/aimatcher"
	"backend-go/pkg/cache"
	"backend-go/pkg/email"
	"backend-go/pkg/fcm"
	"backend-go/pkg/geo"
	"github.com/google/uuid"
)

type serviceUseCase struct {
	categoryRepo repository.CategoryRepository
	catalogRepo  repository.CatalogServiceRepository
	requestRepo  repository.ServiceRequestRepository
	proposalRepo repository.ProposalRepository
	profileRepo  repository.ProfileRepository
	aptRepo      repository.AppointmentRepository
	notifRepo    repository.NotificationRepository
	tokenRepo    repository.DeviceTokenRepository
	fcmClient    fcm.Client
	userRepo     repository.UserRepository
	adminRepo    repository.AdminRepository
	cache        cache.Cache
	cfg          *config.Config
}

func NewServiceUseCase(
	categoryRepo repository.CategoryRepository,
	catalogRepo repository.CatalogServiceRepository,
	requestRepo repository.ServiceRequestRepository,
	proposalRepo repository.ProposalRepository,
	profileRepo repository.ProfileRepository,
	aptRepo repository.AppointmentRepository,
	notifRepo repository.NotificationRepository,
	tokenRepo repository.DeviceTokenRepository,
	fcmClient fcm.Client,
	userRepo repository.UserRepository,
	adminRepo repository.AdminRepository,
	cache cache.Cache,
	cfg *config.Config,
) domainUsecase.ServiceUseCase {
	return &serviceUseCase{
		categoryRepo: categoryRepo,
		catalogRepo:  catalogRepo,
		requestRepo:  requestRepo,
		proposalRepo: proposalRepo,
		profileRepo:  profileRepo,
		aptRepo:      aptRepo,
		notifRepo:    notifRepo,
		tokenRepo:    tokenRepo,
		fcmClient:    fcmClient,
		userRepo:     userRepo,
		adminRepo:    adminRepo,
		cache:        cache,
		cfg:          cfg,
	}
}

func (u *serviceUseCase) invalidateRequestCaches(ctx context.Context) {
	if u.cache != nil {
		_ = u.cache.DeleteByPattern(ctx, "cache:*requests*")
	}
}


func (u *serviceUseCase) sendPush(userID uuid.UUID, title, body string, data map[string]string) {
	if u.fcmClient == nil || u.tokenRepo == nil {
		return
	}
	go func() {
		tokens, err := u.tokenRepo.ListByUser(context.Background(), userID)
		if err != nil || len(tokens) == 0 {
			return
		}
		var tokenList []string
		for _, t := range tokens {
			if t.Token != "" && t.IsActive {
				tokenList = append(tokenList, t.Token)
			}
		}
		if len(tokenList) == 0 {
			return
		}
		invalidTokens, _ := u.fcmClient.SendMulticast(context.Background(), tokenList, title, body, data)
		for _, invToken := range invalidTokens {
			_ = u.tokenRepo.DeleteByToken(context.Background(), invToken)
		}
	}()
}

func (u *serviceUseCase) GetCategories(ctx context.Context) ([]entity.Category, error) {
	return u.categoryRepo.ListActive(ctx)
}

func (u *serviceUseCase) GetCatalogServices(ctx context.Context, categorySlug string, search string) ([]entity.CatalogService, error) {
	return u.catalogRepo.List(ctx, categorySlug, search)
}

func (u *serviceUseCase) MatchProviders(ctx context.Context, input domainUsecase.MatchProvidersInput) ([]entity.Profile, error) {
	radius := 25.0
	if input.MaxDistance > 0 {
		radius = input.MaxDistance
	} else if u.adminRepo != nil {
		if s, err := u.adminRepo.GetSettings(ctx); err == nil && s != nil && s.MatchRadiusKm > 0 {
			radius = s.MatchRadiusKm
		}
	}

	lat := input.Latitude
	lng := input.Longitude

	// If seeker coordinates not provided directly, lookup from user profile
	if (lat == 0 && lng == 0) && input.UserID != uuid.Nil && u.profileRepo != nil {
		if prof, err := u.profileRepo.GetByUserID(ctx, input.UserID); err == nil && prof != nil {
			if prof.Latitude != 0 || prof.Longitude != 0 {
				lat = prof.Latitude
				lng = prof.Longitude
			}
		}
	}

	params := repository.ProfileFilterParams{
		UserType:     "PROVIDER",
		CategorySlug: input.CategoryID,
		Limit:        100,
		Offset:       0,
	}

	if lat != 0 || lng != 0 {
		params.Latitude = &lat
		params.Longitude = &lng
		params.RadiusKm = &radius
	}

	candidates, err := u.profileRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	query := strings.TrimSpace(input.Description)
	if query == "" {
		query = strings.TrimSpace(input.Query)
	}

	// If no text prompt provided, return regular geo/category listing
	if query == "" {
		return candidates, nil
	}

	// Run In-Engine Vector Space Model & Semantic Ontology Matcher
	ranked := aimatcher.RankProviders(query, candidates)
	if len(ranked) == 0 {
		return []entity.Profile{}, nil
	}

	results := make([]entity.Profile, 0, len(ranked))
	for _, r := range ranked {
		p := r.Profile
		p.MatchScore = r.Score
		p.MatchPercentage = r.MatchPercentage
		p.MatchReason = r.MatchReason
		results = append(results, p)
	}

	return results, nil
}

func (u *serviceUseCase) enrichServiceRequest(ctx context.Context, req *entity.ServiceRequest) {
	if req == nil {
		return
	}

	// 1. Enrich User & UserProfile
	if req.User == nil && req.UserID != uuid.Nil && u.userRepo != nil {
		if user, err := u.userRepo.GetByID(ctx, req.UserID); err == nil && user != nil {
			req.User = user
		}
	}

	if req.User != nil {
		req.UserEmail = req.User.Email
		if req.User.Profile != nil {
			req.UserProfile = req.User.Profile
		}
	}

	if req.UserProfile == nil && req.UserID != uuid.Nil && u.profileRepo != nil {
		if prof, err := u.profileRepo.GetByUserID(ctx, req.UserID); err == nil && prof != nil {
			req.UserProfile = prof
			if req.User != nil {
				req.User.Profile = prof
			}
		}
	}

	// If profile still has empty name, use user email prefix
	if req.UserProfile == nil && req.UserID != uuid.Nil {
		firstName := "Seeker"
		if req.User != nil && req.User.Email != "" {
			parts := strings.Split(req.User.Email, "@")
			if len(parts) > 0 {
				firstName = strings.Title(parts[0])
			}
		}
		req.UserProfile = &entity.Profile{
			UserID:    req.UserID,
			FirstName: firstName,
			UserType:  "SEEKER",
		}
	}

	if req.UserProfile != nil && req.UserProfile.ProfilePictureURL == "" && req.UserProfile.ProfilePicture != "" {
		req.UserProfile.ProfilePictureURL = req.UserProfile.ProfilePicture
	}

	// 2. Enrich CatalogService
	if req.CatalogService != nil {
		req.CatalogServiceName = req.CatalogService.Name
	} else if req.CatalogServiceID != nil && u.catalogRepo != nil {
		if cs, err := u.catalogRepo.GetByID(ctx, *req.CatalogServiceID); err == nil && cs != nil {
			req.CatalogService = cs
			req.CatalogServiceName = cs.Name
		}
	}

	// 3. Enrich Proposals & Approval
	if len(req.Proposals) > 0 {
		req.ProposalsCount = len(req.Proposals)
		for _, p := range req.Proposals {
			if p.IsApproved {
				req.Approved = true
				if p.Provider != nil {
					if p.Provider.Profile != nil {
						p.Provider.Profile.UserID = p.Provider.ID
						req.ApprovedUser = p.Provider.Profile
					} else {
						req.ApprovedUser = p.Provider
					}
				} else {
					req.ApprovedUser = p.ProviderID.String()
				}
				break
			}
		}
	} else if u.proposalRepo != nil {
		if proposals, err := u.proposalRepo.ListByRequest(ctx, req.ID); err == nil {
			req.Proposals = proposals
			req.ProposalsCount = len(proposals)
			for _, p := range proposals {
				if p.IsApproved {
					req.Approved = true
					if p.Provider != nil {
						if p.Provider.Profile != nil {
							p.Provider.Profile.UserID = p.Provider.ID
							req.ApprovedUser = p.Provider.Profile
						} else {
							req.ApprovedUser = p.Provider
						}
					} else {
						req.ApprovedUser = p.ProviderID.String()
					}
					break
				}
			}
		}
	}

	// 4. Image URL
	if req.ImageURL == "" && req.Image != "" {
		req.ImageURL = req.Image
	}
}

func (u *serviceUseCase) GetRequests(ctx context.Context, userID uuid.UUID, userType string, status string) ([]entity.ServiceRequest, error) {
	normType := strings.ToUpper(strings.TrimSpace(userType))
	cacheKey := fmt.Sprintf("cache:nearby_requests:user:%s:%s:%s", userID.String(), normType, status)
	if u.cache != nil {
		var cached []entity.ServiceRequest
		if found, _ := u.cache.Get(ctx, cacheKey, &cached); found && len(cached) > 0 {
			return cached, nil
		}
	}

	var requests []entity.ServiceRequest
	var err error
	if normType == "CUSTOMER" || normType == "SEEKER" || normType == "USER" || normType == "" {
		requests, err = u.requestRepo.ListByUser(ctx, &userID, nil, status)
	} else {
		// Provider flow: Fetch provider profile to determine offered services and location
		filterParams := repository.ServiceRequestFilterParams{
			Status:           status,
			TargetProviderID: &userID,
		}

		if profile, pErr := u.profileRepo.GetByUserID(ctx, userID); pErr == nil && profile != nil {
			var offeredServices []string
			if strings.TrimSpace(profile.Service) != "" {
				offeredServices = append(offeredServices, strings.TrimSpace(profile.Service))
			}
			for _, cs := range profile.CatalogServices {
				if strings.TrimSpace(cs.Name) != "" {
					offeredServices = append(offeredServices, strings.TrimSpace(cs.Name))
				}
			}
			for _, name := range profile.CatalogServiceNames {
				if strings.TrimSpace(name) != "" {
					offeredServices = append(offeredServices, strings.TrimSpace(name))
				}
			}
			filterParams.OfferedServices = offeredServices

			if profile.Latitude != 0 || profile.Longitude != 0 {
				lat := profile.Latitude
				lng := profile.Longitude
				filterParams.Latitude = &lat
				filterParams.Longitude = &lng
			}
		}

		requests, err = u.requestRepo.List(ctx, filterParams)
	}
	if err != nil {
		return nil, err
	}
	for i := range requests {
		u.enrichServiceRequest(ctx, &requests[i])
	}

	if u.cache != nil && len(requests) > 0 {
		_ = u.cache.Set(ctx, cacheKey, requests, 5*time.Minute)
	}

	return requests, nil
}

func (u *serviceUseCase) GetRequestByID(ctx context.Context, id uuid.UUID) (*entity.ServiceRequest, error) {
	req, err := u.requestRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	u.enrichServiceRequest(ctx, req)
	return req, nil
}

func (u *serviceUseCase) CreateRequest(ctx context.Context, customerID uuid.UUID, req *entity.ServiceRequest) (*entity.ServiceRequest, error) {
	if req.Title == "" {
		return nil, errors.New("title is required")
	}

	req.ID = uuid.New()
	req.UserID = customerID
	if req.Status == "" {
		req.Status = "OPEN"
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if err := u.requestRepo.Create(ctx, req); err != nil {
		return nil, err
	}

	u.invalidateRequestCaches(ctx)


	// If direct request, notify target provider via in-app notification & email
	if req.TargetProviderID != nil {
		if u.notifRepo != nil {
			notif := entity.Notification{
				ID:               uuid.New(),
				UserID:           *req.TargetProviderID,
				SenderID:         &customerID,
				NotificationType: "DIRECT_REQUEST",
				Title:            "New Direct Request",
				Message:          fmt.Sprintf("You have received a direct service request for %s", req.Title),
				Data:             entity.JSONMap{"request_id": req.ID.String()},
				CreatedAt:        time.Now(),
			}
			_ = u.notifRepo.Create(ctx, &notif)
		}

		u.sendPush(*req.TargetProviderID, "New Direct Request", fmt.Sprintf("You have received a direct service request for %s", req.Title), map[string]string{
			"notification_type": "direct_request",
			"request_id":        req.ID.String(),
			"sender_id":         customerID.String(),
		})

		// Asynchronously send direct request email
		if u.cfg != nil && u.cfg.SMTPHost != "" && u.userRepo != nil {
			go func() {
				pUser, pErr := u.userRepo.GetByID(context.Background(), *req.TargetProviderID)
				sUser, sErr := u.userRepo.GetByID(context.Background(), customerID)
				if pErr == nil && pUser != nil && sErr == nil && sUser != nil {
					pProfile, _ := u.profileRepo.GetByUserID(context.Background(), pUser.ID)
					sProfile, _ := u.profileRepo.GetByUserID(context.Background(), sUser.ID)

					providerName := pUser.Email
					if pProfile != nil && pProfile.FirstName != "" {
						providerName = pProfile.FirstName
					}
					seekerName := sUser.Email
					if sProfile != nil && sProfile.FirstName != "" {
						seekerName = sProfile.FirstName
					}

					price := 0.0
					if req.Price != nil {
						price = *req.Price
					}
					sched := "Flexible / Not set"
					if req.ScheduledTime != nil {
						sched = req.ScheduledTime.Format("January 02, 2006 at 03:04 PM")
					}

					_ = email.SendDirectRequestEmail(&email.Config{
						Host:     u.cfg.SMTPHost,
						Port:     u.cfg.SMTPPort,
						User:     u.cfg.SMTPUser,
						Password: u.cfg.SMTPPassword,
						From:     u.cfg.EmailFrom,
					}, pUser.Email, providerName, seekerName, req.Title, price, sched, req.Description)
				}
			}()
		}
	} else if req.Latitude != nil && req.Longitude != nil {
		// If broadcast public request with coordinates, notify verified providers within the configured radius
		go func() {
			radius := 25.0
			if u.adminRepo != nil {
				if s, err := u.adminRepo.GetSettings(context.Background()); err == nil && s != nil && s.BroadcastRadiusKm > 0 {
					radius = s.BroadcastRadiusKm
				}
			}

			nearbyProviders, err := u.profileRepo.List(context.Background(), repository.ProfileFilterParams{
				UserType:  "PROVIDER",
				Latitude:  req.Latitude,
				Longitude: req.Longitude,
				RadiusKm:  &radius,
			})
			if err == nil {
				for _, p := range nearbyProviders {
					if p.UserID == customerID {
						continue
					}

					dist := 0.0
					if p.Latitude != 0 && p.Longitude != 0 {
						dist = geo.HaversineDistance(*req.Latitude, *req.Longitude, p.Latitude, p.Longitude)
					}

					// In-App Push Notification
					if u.notifRepo != nil {
						notif := entity.Notification{
							ID:               uuid.New(),
							UserID:           p.UserID,
							SenderID:         &customerID,
							NotificationType: "BROADCAST_REQUEST",
							Title:            "New Service Request Nearby",
							Message:          fmt.Sprintf("A new request for '%s' was posted %.1f km away from your location.", req.Title, dist),
							Data:             entity.JSONMap{"request_id": req.ID.String()},
							CreatedAt:        time.Now(),
						}
						_ = u.notifRepo.Create(context.Background(), &notif)
					}

					u.sendPush(p.UserID, "New Service Request Nearby", fmt.Sprintf("A new request for '%s' was posted %.1f km away from your location.", req.Title, dist), map[string]string{
						"notification_type": "broadcast_request",
						"request_id":        req.ID.String(),
						"sender_id":         customerID.String(),
					})

					// Broadcast Email Notification
					if u.cfg != nil && u.cfg.SMTPHost != "" && p.User != nil && p.User.Email != "" {
						providerName := p.FirstName
						if providerName == "" {
							providerName = p.User.Email
						}
						seekerName := "A neighbor"
						if sUser, sErr := u.userRepo.GetByID(context.Background(), customerID); sErr == nil && sUser != nil {
							seekerName = sUser.Email
							if sProf, _ := u.profileRepo.GetByUserID(context.Background(), customerID); sProf != nil && sProf.FirstName != "" {
								seekerName = sProf.FirstName
							}
						}

						sched := "Flexible / Not set"
						if req.ScheduledTime != nil {
							sched = req.ScheduledTime.Format("January 02, 2006 at 03:04 PM")
						}

						priceVal := 0.0
						if req.Price != nil {
							priceVal = *req.Price
						}

						_ = email.SendBroadcastRequestEmail(&email.Config{
							Host:     u.cfg.SMTPHost,
							Port:     u.cfg.SMTPPort,
							User:     u.cfg.SMTPUser,
							Password: u.cfg.SMTPPassword,
							From:     u.cfg.EmailFrom,
						}, p.User.Email, providerName, seekerName, req.Title, dist, priceVal, sched, req.Description)
					}
				}
			}
		}()
	}

	return u.GetRequestByID(ctx, req.ID)
}

func (u *serviceUseCase) UpdateRequest(ctx context.Context, userID, id uuid.UUID, updates map[string]interface{}) (*entity.ServiceRequest, error) {
	req, err := u.requestRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("request not found")
	}

	if req.UserID != userID {
		return nil, errors.New("not authorized")
	}

	if title, ok := updates["title"].(string); ok && title != "" {
		req.Title = title
	}
	if desc, ok := updates["description"].(string); ok {
		req.Description = desc
	}
	if status, ok := updates["status"].(string); ok {
		req.Status = status
	}
	if price, ok := updates["price"].(float64); ok {
		req.Price = &price
	}
	if schedStr, ok := updates["scheduled_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, schedStr); err == nil {
			req.ScheduledTime = &t
		}
	}

	req.UpdatedAt = time.Now()
	if err := u.requestRepo.Update(ctx, req); err != nil {
		return nil, err
	}

	u.invalidateRequestCaches(ctx)
	return u.GetRequestByID(ctx, id)
}

func (u *serviceUseCase) UploadRequestImage(ctx context.Context, userID, id uuid.UUID, imageURL string) (*entity.ServiceRequest, error) {
	req, err := u.requestRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("request not found")
	}

	if req.UserID != userID {
		return nil, errors.New("not authorized")
	}

	req.Image = imageURL
	req.WithImage = (imageURL != "")
	req.UpdatedAt = time.Now()
	if err := u.requestRepo.Update(ctx, req); err != nil {
		return nil, err
	}

	u.invalidateRequestCaches(ctx)
	return u.GetRequestByID(ctx, id)
}

func (u *serviceUseCase) DeleteRequest(ctx context.Context, userID, id uuid.UUID) error {
	req, err := u.requestRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("request not found")
	}

	if req.UserID != userID {
		return errors.New("not authorized")
	}

	if u.aptRepo != nil {
		apts, _ := u.aptRepo.List(ctx, &userID, nil, "")
		for _, apt := range apts {
			if apt.ServiceRequestID != nil && *apt.ServiceRequestID == id {
				_ = u.aptRepo.Delete(ctx, apt.ID)
			}
		}
	}

	u.invalidateRequestCaches(ctx)
	return u.requestRepo.Delete(ctx, id)
}


func (u *serviceUseCase) ApproveProposal(ctx context.Context, userID, requestID, proposalID uuid.UUID) (*entity.ServiceRequest, error) {
	req, err := u.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return nil, errors.New("request not found")
	}

	if req.UserID != userID {
		return nil, errors.New("not authorized")
	}

	if req.Status == "IN_PROGRESS" || req.Status == "DONE" {
		return nil, fmt.Errorf("request is already in progress or completed (current status: %s)", req.Status)
	}

	proposal, err := u.proposalRepo.GetByID(ctx, proposalID)
	if err != nil || proposal.RequestID != requestID {
		return nil, errors.New("proposal not found")
	}

	proposal.IsApproved = true
	_ = u.proposalRepo.Update(ctx, proposal)

	var apt *entity.Appointment
	// Automatically create appointment if AppointmentRepository is available
	if u.aptRepo != nil {
		// Clean up any existing appointments linked to this request to prevent duplicates
		existingApts, _ := u.aptRepo.List(ctx, &req.UserID, nil, "")
		for _, existing := range existingApts {
			if existing.ServiceRequestID != nil && *existing.ServiceRequestID == req.ID {
				_ = u.aptRepo.Delete(ctx, existing.ID)
			}
		}

		totalPrice := 0.0
		if req.Price != nil && *req.Price > 0 {
			totalPrice = *req.Price
		}
		newApt := entity.Appointment{
			ID:               uuid.New(),
			SeekerID:         req.UserID,
			ProviderID:       proposal.ProviderID,
			Title:            req.Title,
			Description:      req.Description,
			AppointmentDate:  req.ScheduledTime,
			ServiceRequestID: &req.ID,
			ProposalID:       &proposal.ID,
			TotalPrice:       totalPrice,
			PaymentMode:      req.PreferredPaymentMode,
			Status:           "SCHEDULED",
			SecretCode:       fmt.Sprintf("%04d", rand.Intn(10000)),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		_ = u.aptRepo.Create(ctx, &newApt)
		apt = &newApt
	}

	req.Status = "IN_PROGRESS"
	req.UpdatedAt = time.Now()
	_ = u.requestRepo.Update(ctx, req)

	// Notify provider in-app
	if u.notifRepo != nil {
		notif := entity.Notification{
			ID:               uuid.New(),
			UserID:           proposal.ProviderID,
			SenderID:         &userID,
			NotificationType: "PROPOSAL",
			Title:            "Proposal Accepted",
			Message:          fmt.Sprintf("Your proposal for '%s' has been accepted!", req.Title),
			Data:             entity.JSONMap{"proposal_id": proposal.ID.String(), "request_id": req.ID.String()},
			CreatedAt:        time.Now(),
		}
		_ = u.notifRepo.Create(ctx, &notif)
	}

	u.sendPush(proposal.ProviderID, "Proposal Accepted", fmt.Sprintf("Your proposal for '%s' has been accepted!", req.Title), map[string]string{
		"notification_type": "proposal",
		"request_id":        req.ID.String(),
		"proposal_id":       proposal.ID.String(),
		"sender_id":         userID.String(),
	})

	// Asynchronously send approval email & appointment confirmation
	if u.cfg != nil && u.cfg.SMTPHost != "" && u.userRepo != nil {
		go func() {
			pUser, pErr := u.userRepo.GetByID(context.Background(), proposal.ProviderID)
			sUser, sErr := u.userRepo.GetByID(context.Background(), req.UserID)
			if pErr == nil && pUser != nil && sErr == nil && sUser != nil {
				pProfile, _ := u.profileRepo.GetByUserID(context.Background(), pUser.ID)
				sProfile, _ := u.profileRepo.GetByUserID(context.Background(), sUser.ID)

				providerName := pUser.Email
				if pProfile != nil && pProfile.FirstName != "" {
					providerName = pProfile.FirstName
				}
				seekerName := sUser.Email
				if sProfile != nil && sProfile.FirstName != "" {
					seekerName = sProfile.FirstName
				}

				sched := "Scheduled"
				if req.ScheduledTime != nil {
					sched = req.ScheduledTime.Format("January 02, 2006 at 03:04 PM")
				}

				priceVal := 0.0
				if req.Price != nil {
					priceVal = *req.Price
				}

				emailCfg := &email.Config{
					Host:     u.cfg.SMTPHost,
					Port:     u.cfg.SMTPPort,
					User:     u.cfg.SMTPUser,
					Password: u.cfg.SMTPPassword,
					From:     u.cfg.EmailFrom,
				}

				_ = email.SendProposalApprovedEmail(emailCfg, pUser.Email, providerName, seekerName, sUser.Email, req.Title, priceVal, sched)

				if apt != nil {
					_ = email.SendAppointmentBookedEmail(emailCfg, sUser.Email, seekerName, providerName, req.Title, sched, "")
					_ = email.SendAppointmentBookedEmail(emailCfg, pUser.Email, providerName, seekerName, req.Title, sched, "")
				}
			}
		}()
	}

	u.invalidateRequestCaches(ctx)
	return u.GetRequestByID(ctx, requestID)
}


func (u *serviceUseCase) CancelApproval(ctx context.Context, userID, requestID uuid.UUID) error {
	req, err := u.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return errors.New("request not found")
	}

	if req.UserID != userID {
		return errors.New("not authorized")
	}

	proposals, _ := u.proposalRepo.ListByRequest(ctx, requestID)
	var approvedProviderID *uuid.UUID
	for _, prop := range proposals {
		if prop.IsApproved {
			prop.IsApproved = false
			_ = u.proposalRepo.Update(ctx, &prop)
			pID := prop.ProviderID
			approvedProviderID = &pID
		}
	}

	req.Status = "OPEN"
	req.UpdatedAt = time.Now()
	_ = u.requestRepo.Update(ctx, req)

	// Delete the appointment created when the proposal was approved to prevent duplicate cancelled appointments
	if u.aptRepo != nil {
		apts, _ := u.aptRepo.List(ctx, &userID, nil, "")
		for _, apt := range apts {
			if apt.ServiceRequestID != nil && *apt.ServiceRequestID == requestID {
				_ = u.aptRepo.Delete(ctx, apt.ID)
			}
		}
	}

	if approvedProviderID != nil && u.notifRepo != nil {
		notif := entity.Notification{
			ID:               uuid.New(),
			UserID:           *approvedProviderID,
			SenderID:         &userID,
			NotificationType: "PROPOSAL",
			Title:            "Approval Cancelled",
			Message:          fmt.Sprintf("The seeker has cancelled the approval for your proposal on '%s'.", req.Title),
			Data:             entity.JSONMap{"request_id": req.ID.String()},
			CreatedAt:        time.Now(),
		}
		_ = u.notifRepo.Create(ctx, &notif)
	}

	if approvedProviderID != nil {
		u.sendPush(*approvedProviderID, "Approval Cancelled", fmt.Sprintf("The seeker has cancelled the approval for your proposal on '%s'.", req.Title), map[string]string{
			"notification_type": "proposal",
			"request_id":        req.ID.String(),
			"sender_id":         userID.String(),
		})
	}

	u.invalidateRequestCaches(ctx)
	return nil
}


func (u *serviceUseCase) GetProposals(ctx context.Context, userID uuid.UUID, userType string, requestID *uuid.UUID) ([]entity.Proposal, error) {
	if requestID != nil {
		return u.proposalRepo.ListByRequest(ctx, *requestID)
	}
	return u.proposalRepo.ListByProvider(ctx, userID)
}

func (u *serviceUseCase) CreateProposal(ctx context.Context, providerUserID uuid.UUID, prop *entity.Proposal) (*entity.Proposal, error) {
	prop.ID = uuid.New()
	prop.ProviderID = providerUserID
	prop.IsApproved = false
	prop.CreatedAt = time.Now()
	prop.UpdatedAt = time.Now()

	if err := u.proposalRepo.Create(ctx, prop); err != nil {
		return nil, err
	}

	// Notify seeker in-app & email
	req, err := u.requestRepo.GetByID(ctx, prop.RequestID)
	if err == nil && req != nil {
		if u.notifRepo != nil {
			notif := entity.Notification{
				ID:               uuid.New(),
				UserID:           req.UserID,
				SenderID:         &providerUserID,
				NotificationType: "PROPOSAL",
				Title:            "New Proposal",
				Message:          fmt.Sprintf("You received a new proposal on '%s'", req.Title),
				Data:             entity.JSONMap{"proposal_id": prop.ID.String(), "request_id": req.ID.String()},
				CreatedAt:        time.Now(),
			}
			_ = u.notifRepo.Create(ctx, &notif)
		}

		u.sendPush(req.UserID, "New Proposal", fmt.Sprintf("You received a new proposal on '%s'", req.Title), map[string]string{
			"notification_type": "proposal",
			"request_id":        req.ID.String(),
			"proposal_id":       prop.ID.String(),
			"sender_id":         providerUserID.String(),
		})

		if u.cfg != nil && u.cfg.SMTPHost != "" && u.userRepo != nil {
			go func() {
				sUser, sErr := u.userRepo.GetByID(context.Background(), req.UserID)
				pUser, pErr := u.userRepo.GetByID(context.Background(), providerUserID)
				if sErr == nil && sUser != nil && pErr == nil && pUser != nil {
					sProfile, _ := u.profileRepo.GetByUserID(context.Background(), sUser.ID)
					pProfile, _ := u.profileRepo.GetByUserID(context.Background(), pUser.ID)

					seekerName := sUser.Email
					if sProfile != nil && sProfile.FirstName != "" {
						seekerName = sProfile.FirstName
					}
					providerName := pUser.Email
					if pProfile != nil && pProfile.FirstName != "" {
						providerName = pProfile.FirstName
					}

					priceVal := 0.0
					if req.Price != nil {
						priceVal = *req.Price
					}

					_ = email.SendNewProposalEmail(&email.Config{
						Host:     u.cfg.SMTPHost,
						Port:     u.cfg.SMTPPort,
						User:     u.cfg.SMTPUser,
						Password: u.cfg.SMTPPassword,
						From:     u.cfg.EmailFrom,
					}, sUser.Email, seekerName, providerName, req.Title, priceVal, req.Description)
				}
			}()
		}
	}

	return u.proposalRepo.GetByID(ctx, prop.ID)
}

func (u *serviceUseCase) DeleteProposal(ctx context.Context, providerID, idOrReqID uuid.UUID) error {
	prop, err := u.proposalRepo.GetByID(ctx, idOrReqID)
	if err != nil || prop == nil {
		// Fallback: treat idOrReqID as request ID
		prop, err = u.proposalRepo.GetByRequestAndProvider(ctx, idOrReqID, providerID)
		if err != nil || prop == nil {
			return errors.New("proposal not found")
		}
	}

	if prop.ProviderID != providerID {
		return errors.New("not authorized")
	}

	if prop.IsApproved {
		return errors.New("Cannot delete an approved proposal.")
	}

	return u.proposalRepo.Delete(ctx, prop.ID)
}
