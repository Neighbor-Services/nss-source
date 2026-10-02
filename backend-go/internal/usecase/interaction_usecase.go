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
	"backend-go/internal/websocket"
	"backend-go/pkg/email"
	"backend-go/pkg/fcm"

	"github.com/google/uuid"
)

type interactionUseCase struct {
	favRepo     repository.FavoriteRepository
	reviewRepo  repository.ReviewRepository
	aptRepo     repository.AppointmentRepository
	disputeRepo repository.DisputeRepository
	profileRepo repository.ProfileRepository
	walletRepo  repository.WalletRepository
	txRepo      repository.WalletTransactionRepository
	userRepo    repository.UserRepository
	requestRepo repository.ServiceRequestRepository
	notifRepo   repository.NotificationRepository
	tokenRepo   repository.DeviceTokenRepository
	fcmClient   fcm.Client
	cfg         *config.Config
}

func NewInteractionUseCase(
	favRepo repository.FavoriteRepository,
	reviewRepo repository.ReviewRepository,
	aptRepo repository.AppointmentRepository,
	disputeRepo repository.DisputeRepository,
	profileRepo repository.ProfileRepository,
	walletRepo repository.WalletRepository,
	txRepo repository.WalletTransactionRepository,
	userRepo repository.UserRepository,
	requestRepo repository.ServiceRequestRepository,
	notifRepo repository.NotificationRepository,
	tokenRepo repository.DeviceTokenRepository,
	fcmClient fcm.Client,
	cfg *config.Config,
) domainUsecase.InteractionUseCase {
	return &interactionUseCase{
		favRepo:     favRepo,
		reviewRepo:  reviewRepo,
		aptRepo:     aptRepo,
		disputeRepo: disputeRepo,
		profileRepo: profileRepo,
		walletRepo:  walletRepo,
		txRepo:      txRepo,
		userRepo:    userRepo,
		requestRepo: requestRepo,
		notifRepo:   notifRepo,
		tokenRepo:   tokenRepo,
		fcmClient:   fcmClient,
		cfg:         cfg,
	}
}

func (u *interactionUseCase) sendPush(userID uuid.UUID, title, body string, data map[string]string) {
	// Broadcast over WebSocket immediately to any active sessions
	websocket.GlobalHub.SendToUser(userID.String(), map[string]interface{}{
		"type":       "notification",
		"title":      title,
		"body":       body,
		"data":       data,
		"created_at": time.Now().UTC(),
	})

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

func (u *interactionUseCase) GetFavorites(ctx context.Context, userID uuid.UUID) ([]entity.Favorite, error) {
	return u.favRepo.ListByUser(ctx, userID)
}

func (u *interactionUseCase) CreateFavorite(ctx context.Context, userID, favoriteUserID uuid.UUID) (*entity.Favorite, error) {
	if prof, err := u.profileRepo.GetByID(ctx, favoriteUserID); err == nil && prof != nil && prof.UserID != uuid.Nil {
		favoriteUserID = prof.UserID
	}
	fav := entity.Favorite{
		ID:             uuid.New(),
		UserID:         userID,
		FavoriteUserID: favoriteUserID,
		CreatedAt:      time.Now(),
	}
	if err := u.favRepo.Create(ctx, &fav); err != nil {
		return nil, err
	}
	return &fav, nil
}

func (u *interactionUseCase) DeleteFavorite(ctx context.Context, userID, favoriteUserID uuid.UUID) error {
	if prof, err := u.profileRepo.GetByID(ctx, favoriteUserID); err == nil && prof != nil && prof.UserID != uuid.Nil {
		favoriteUserID = prof.UserID
	}
	return u.favRepo.Delete(ctx, userID, favoriteUserID)
}

func (u *interactionUseCase) GetReviews(ctx context.Context, providerID uuid.UUID) ([]entity.Review, error) {
	if prof, err := u.profileRepo.GetByID(ctx, providerID); err == nil && prof != nil && prof.UserID != uuid.Nil {
		providerID = prof.UserID
	}
	reviews, err := u.reviewRepo.ListByProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}
	for i := range reviews {
		if reviews[i].Reviewer != nil && reviews[i].Reviewer.Profile != nil {
			reviews[i].ReviewerProfile = reviews[i].Reviewer.Profile
		}
		if reviews[i].Provider != nil && reviews[i].Provider.Profile != nil {
			reviews[i].ProviderProfile = reviews[i].Provider.Profile
		}
	}
	return reviews, nil
}

func (u *interactionUseCase) CreateReview(ctx context.Context, reviewerID uuid.UUID, review *entity.Review) (*entity.Review, error) {
	if prof, err := u.profileRepo.GetByID(ctx, review.ProviderID); err == nil && prof != nil && prof.UserID != uuid.Nil {
		review.ProviderID = prof.UserID
	}

	// Prevent providers from reviewing their own profile
	if review.ProviderID == reviewerID {
		return nil, errors.New("you cannot review your own profile")
	}

	review.ID = uuid.New()
	review.ReviewerID = reviewerID
	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()

	if err := u.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	// Recalculate provider average rating and total reviews
	if avg, count, err := u.reviewRepo.CalculateProviderRating(ctx, review.ProviderID); err == nil {
		if profile, pErr := u.profileRepo.GetByUserID(ctx, review.ProviderID); pErr == nil && profile != nil {
			profile.AverageRating = avg
			profile.TotalReviews = count
			profile.XP += 25 // Award XP
			_ = u.profileRepo.Update(ctx, profile)
		}
	}

	if u.notifRepo != nil {
		notif := entity.Notification{
			ID:               uuid.New(),
			UserID:           review.ProviderID,
			SenderID:         &reviewerID,
			NotificationType: "REVIEW",
			Title:            "New Review Received",
			Message:          fmt.Sprintf("A client left you a %d-star review!", int(review.Rating)),
			Data:             entity.JSONMap{"review_id": review.ID.String(), "provider_id": review.ProviderID.String()},
			CreatedAt:        time.Now(),
		}
		_ = u.notifRepo.Create(ctx, &notif)
	}

	u.sendPush(review.ProviderID, "New Review Received", fmt.Sprintf("A client left you a %d-star review!", int(review.Rating)), map[string]string{
		"notification_type": "review",
		"provider_id":       review.ProviderID.String(),
		"review_id":         review.ID.String(),
	})

	return review, nil
}

func (u *interactionUseCase) GetAppointments(ctx context.Context, userID uuid.UUID, userType, status string) ([]entity.Appointment, error) {
	normType := strings.ToUpper(strings.TrimSpace(userType))
	var list []entity.Appointment
	var err error

	switch normType {
	case "CUSTOMER", "SEEKER":
		list, err = u.aptRepo.List(ctx, &userID, nil, status)
		if err != nil || len(list) == 0 {
			list, err = u.aptRepo.List(ctx, &userID, &userID, status)
		}
	case "PROVIDER":
		list, err = u.aptRepo.List(ctx, nil, &userID, status)
		if err != nil || len(list) == 0 {
			list, err = u.aptRepo.List(ctx, &userID, &userID, status)
		}
	default:
		list, err = u.aptRepo.List(ctx, &userID, &userID, status)
	}

	if err != nil {
		return nil, err
	}

	for i := range list {
		if list[i].Seeker != nil && list[i].Seeker.Profile != nil {
			list[i].SeekerProfile = list[i].Seeker.Profile
		}
		if list[i].Provider != nil && list[i].Provider.Profile != nil {
			list[i].ProviderProfile = list[i].Provider.Profile
		}
	}

	return list, nil
}

func (u *interactionUseCase) checkAppointmentParties(ctx context.Context, apt *entity.Appointment, userID uuid.UUID) (isSeeker bool, isProvider bool) {
	if apt.SeekerID == userID || (apt.Seeker != nil && (apt.Seeker.ID == userID || (apt.Seeker.Profile != nil && apt.Seeker.Profile.ID == userID))) {
		isSeeker = true
	}
	if apt.ProviderID == userID || (apt.Provider != nil && (apt.Provider.ID == userID || (apt.Provider.Profile != nil && apt.Provider.Profile.ID == userID))) {
		isProvider = true
	}
	if !isSeeker && !isProvider {
		if prof, err := u.profileRepo.GetByID(ctx, userID); err == nil && prof != nil {
			if apt.SeekerID == prof.UserID || (apt.Seeker != nil && apt.Seeker.ID == prof.UserID) {
				isSeeker = true
			}
			if apt.ProviderID == prof.UserID || (apt.Provider != nil && apt.Provider.ID == prof.UserID) {
				isProvider = true
			}
		}
	}
	return isSeeker, isProvider
}

func parseAppointmentDate(val interface{}) *time.Time {
	if val == nil {
		return nil
	}
	if t, ok := val.(time.Time); ok {
		return &t
	}
	str, ok := val.(string)
	if !ok || strings.TrimSpace(str) == "" {
		return nil
	}
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, layout := range formats {
		if t, err := time.Parse(layout, strings.TrimSpace(str)); err == nil {
			return &t
		}
	}
	return nil
}

func (u *interactionUseCase) GetAppointmentByID(ctx context.Context, userID, appointmentID uuid.UUID) (*entity.Appointment, error) {
	apt, err := u.aptRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, errors.New("appointment not found")
	}

	isSeeker, isProvider := u.checkAppointmentParties(ctx, apt, userID)
	if !isSeeker && !isProvider {
		return nil, errors.New("not authorized")
	}

	if apt.Seeker != nil && apt.Seeker.Profile != nil {
		apt.SeekerProfile = apt.Seeker.Profile
	}
	if apt.Provider != nil && apt.Provider.Profile != nil {
		apt.ProviderProfile = apt.Provider.Profile
	}

	return apt, nil
}

func (u *interactionUseCase) UpdateAppointment(ctx context.Context, userID, appointmentID uuid.UUID, updates map[string]interface{}) (*entity.Appointment, error) {
	apt, err := u.aptRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, errors.New("appointment not found")
	}

	isSeeker, isProvider := u.checkAppointmentParties(ctx, apt, userID)
	if !isSeeker && !isProvider {
		return nil, errors.New("not authorized")
	}

	if title, ok := updates["title"].(string); ok && title != "" {
		apt.Title = title
	}
	if desc, ok := updates["description"].(string); ok {
		apt.Description = desc
	}
	if status, ok := updates["status"].(string); ok && status != "" {
		apt.Status = strings.ToUpper(status)
	}
	if price, ok := updates["total_price"].(float64); ok {
		apt.TotalPrice = price
	} else if priceAlt, ok := updates["totalPrice"].(float64); ok {
		apt.TotalPrice = priceAlt
	}
	if payMode, ok := updates["payment_mode"].(string); ok && payMode != "" {
		apt.PaymentMode = payMode
	}

	dateRaw := updates["appointment_date"]
	if dateRaw == nil {
		dateRaw = updates["appointmentDate"]
	}
	if dateRaw != nil {
		if parsed := parseAppointmentDate(dateRaw); parsed != nil {
			apt.AppointmentDate = parsed
			apt.NoShowProcessed = false
		}
	}

	apt.UpdatedAt = time.Now()
	if err := u.aptRepo.Update(ctx, apt); err != nil {
		return nil, err
	}

	// Sync scheduled time with linked service request
	if apt.ServiceRequestID != nil && *apt.ServiceRequestID != uuid.Nil && apt.AppointmentDate != nil && u.requestRepo != nil {
		if req, rErr := u.requestRepo.GetByID(ctx, *apt.ServiceRequestID); rErr == nil && req != nil {
			req.ScheduledTime = apt.AppointmentDate
			req.UpdatedAt = time.Now()
			_ = u.requestRepo.Update(ctx, req)
		}
	}

	// Broadcast real-time update over WebSocket to both parties
	websocket.GlobalHub.SendToUser(apt.SeekerID.String(), map[string]interface{}{
		"type":           "appointment_updated",
		"appointment_id": apt.ID.String(),
		"status":         apt.Status,
		"title":          "Appointment Updated",
		"message":        fmt.Sprintf("The appointment for '%s' has been updated.", apt.Title),
		"created_at":     time.Now().UTC(),
	})
	websocket.GlobalHub.SendToUser(apt.ProviderID.String(), map[string]interface{}{
		"type":           "appointment_updated",
		"appointment_id": apt.ID.String(),
		"status":         apt.Status,
		"title":          "Appointment Updated",
		"message":        fmt.Sprintf("The appointment for '%s' has been updated.", apt.Title),
		"created_at":     time.Now().UTC(),
	})

	otherPartyID := apt.ProviderID
	if isProvider {
		otherPartyID = apt.SeekerID
	}
	u.sendPush(otherPartyID, "Appointment Updated", fmt.Sprintf("The appointment for '%s' has been updated.", apt.Title), map[string]string{
		"notification_type": "appointment",
		"appointment_id":    apt.ID.String(),
		"sender_id":         userID.String(),
	})

	return u.GetAppointmentByID(ctx, userID, appointmentID)
}

func (u *interactionUseCase) CreateAppointment(ctx context.Context, customerID uuid.UUID, apt *entity.Appointment) (*entity.Appointment, error) {
	apt.ID = uuid.New()
	apt.SeekerID = customerID
	if apt.Status == "" {
		apt.Status = "SCHEDULED"
	}
	if apt.PaymentMode == "" {
		apt.PaymentMode = "ON_SITE"
	}
	// Generate 4-digit secret code for arrival verification
	apt.SecretCode = fmt.Sprintf("%04d", rand.Intn(10000))
	apt.CreatedAt = time.Now()
	apt.UpdatedAt = time.Now()

	if err := u.aptRepo.Create(ctx, apt); err != nil {
		return nil, err
	}

	if u.notifRepo != nil {
		notif := entity.Notification{
			ID:               uuid.New(),
			UserID:           apt.ProviderID,
			SenderID:         &customerID,
			NotificationType: "APPOINTMENT",
			Title:            "New Appointment Booked",
			Message:          fmt.Sprintf("You have a new appointment for %s", apt.Title),
			Data:             entity.JSONMap{"appointment_id": apt.ID.String()},
			CreatedAt:        time.Now(),
		}
		_ = u.notifRepo.Create(ctx, &notif)
	}
	u.sendPush(apt.ProviderID, "New Appointment Booked", fmt.Sprintf("You have a new appointment for %s", apt.Title), map[string]string{
		"notification_type": "appointment",
		"appointment_id":    apt.ID.String(),
		"sender_id":         customerID.String(),
	})

	// Asynchronously send appointment confirmation email to both Seeker and Provider
	if u.cfg != nil && u.cfg.SMTPHost != "" && u.userRepo != nil {
		go func() {
			sUser, sErr := u.userRepo.GetByID(context.Background(), apt.SeekerID)
			pUser, pErr := u.userRepo.GetByID(context.Background(), apt.ProviderID)
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

				sched := "Scheduled"
				if apt.AppointmentDate != nil {
					sched = apt.AppointmentDate.Format("January 02, 2006 at 03:04 PM")
				}

				emailCfg := &email.Config{
					Host:     u.cfg.SMTPHost,
					Port:     u.cfg.SMTPPort,
					User:     u.cfg.SMTPUser,
					Password: u.cfg.SMTPPassword,
					From:     u.cfg.EmailFrom,
				}

				_ = email.SendAppointmentBookedEmail(emailCfg, sUser.Email, seekerName, providerName, apt.Title, sched, "")
				_ = email.SendAppointmentBookedEmail(emailCfg, pUser.Email, providerName, seekerName, apt.Title, sched, "")
			}
		}()
	}

	return u.aptRepo.GetByID(ctx, apt.ID)
}

func (u *interactionUseCase) VerifyArrivalCode(ctx context.Context, providerID, appointmentID uuid.UUID, code string) (*entity.Appointment, error) {
	apt, err := u.aptRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, errors.New("appointment not found")
	}

	_, isProvider := u.checkAppointmentParties(ctx, apt, providerID)
	if !isProvider {
		return nil, errors.New("only the provider can verify the arrival code")
	}

	if apt.Status == "COMPLETED" {
		return nil, errors.New("appointment is already completed")
	}

	if strings.TrimSpace(code) != strings.TrimSpace(apt.SecretCode) {
		// Real-time security alert to seeker that an invalid code was entered
		websocket.GlobalHub.SendToUser(apt.SeekerID.String(), map[string]interface{}{
			"type":           "appointment_code_failed",
			"appointment_id": apt.ID.String(),
			"title":          "Security Alert: Invalid Code Attempt",
			"message":        "An incorrect arrival code was entered for your appointment.",
			"created_at":     time.Now().UTC(),
		})

		u.sendPush(apt.SeekerID, "Security Alert: Invalid Code Attempt", "An incorrect arrival code was entered for your appointment.", map[string]string{
			"notification_type": "appointment_code_failed",
			"appointment_id":    apt.ID.String(),
			"sender_id":         providerID.String(),
		})

		return nil, errors.New("invalid verification code. Please check with the seeker.")
	}

	apt.Status = "IN_PROGRESS"
	apt.UpdatedAt = time.Now()
	if err := u.aptRepo.Update(ctx, apt); err != nil {
		return nil, err
	}

	// Real-time WebSocket broadcast to seeker and provider
	websocket.GlobalHub.SendToUser(apt.SeekerID.String(), map[string]interface{}{
		"type":           "appointment_code_verified",
		"appointment_id": apt.ID.String(),
		"status":         "IN_PROGRESS",
		"title":          "Appointment Code Verified!",
		"message":        "Your provider has verified the arrival code! Appointment is now in progress.",
		"created_at":     time.Now().UTC(),
	})
	websocket.GlobalHub.SendToUser(apt.ProviderID.String(), map[string]interface{}{
		"type":           "appointment_code_verified",
		"appointment_id": apt.ID.String(),
		"status":         "IN_PROGRESS",
		"title":          "Code Verified!",
		"message":        "Arrival code verified. Appointment is now in progress.",
		"created_at":     time.Now().UTC(),
	})

	if u.notifRepo != nil {
		notif := entity.Notification{
			ID:               uuid.New(),
			UserID:           apt.SeekerID,
			SenderID:         &providerID,
			NotificationType: "APPOINTMENT",
			Title:            "Appointment Code Verified",
			Message:          "Your provider has verified the arrival code! Appointment started.",
			Data:             entity.JSONMap{"appointment_id": apt.ID.String(), "status": "IN_PROGRESS"},
			CreatedAt:        time.Now(),
		}
		_ = u.notifRepo.Create(ctx, &notif)
	}
	u.sendPush(apt.SeekerID, "Appointment Code Verified", "Your provider has verified the arrival code! Appointment started.", map[string]string{
		"notification_type": "appointment_code_verified",
		"appointment_id":    apt.ID.String(),
		"status":            "IN_PROGRESS",
		"sender_id":         providerID.String(),
	})

	return apt, nil
}

func (u *interactionUseCase) NotifyOnTheWay(ctx context.Context, providerID, appointmentID uuid.UUID) (*entity.Appointment, error) {
	apt, err := u.aptRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, errors.New("appointment not found")
	}

	_, isProvider := u.checkAppointmentParties(ctx, apt, providerID)
	if !isProvider {
		return nil, errors.New("only the provider can send on-the-way notification")
	}

	if apt.Status == "COMPLETED" || apt.Status == "CANCELLED" {
		return nil, fmt.Errorf("appointment is already %s", strings.ToLower(apt.Status))
	}

	providerName := "Your service provider"
	if pProfile, err := u.profileRepo.GetByID(ctx, providerID); err == nil && pProfile != nil && pProfile.FirstName != "" {
		if pProfile.LastName != "" {
			providerName = fmt.Sprintf("%s %s", pProfile.FirstName, pProfile.LastName)
		} else {
			providerName = pProfile.FirstName
		}
	}

	title := "Provider is on the way!"
	msg := fmt.Sprintf("%s is on the way for '%s'. Tap to track live location.", providerName, apt.Title)

	if u.notifRepo != nil {
		notif := entity.Notification{
			ID:               uuid.New(),
			UserID:           apt.SeekerID,
			SenderID:         &providerID,
			NotificationType: "TRACKING",
			Title:            title,
			Message:          msg,
			Data: entity.JSONMap{
				"notification_type": "provider_on_the_way",
				"appointment_id":    apt.ID.String(),
				"request_id":        apt.ID.String(),
			},
			CreatedAt: time.Now(),
		}
		_ = u.notifRepo.Create(ctx, &notif)
	}

	u.sendPush(apt.SeekerID, title, msg, map[string]string{
		"notification_type": "provider_on_the_way",
		"appointment_id":    apt.ID.String(),
		"request_id":        apt.ID.String(),
		"sender_id":         providerID.String(),
	})

	return apt, nil
}

func (u *interactionUseCase) NotifyArrived(ctx context.Context, providerID, appointmentID uuid.UUID, providerLat, providerLng *float64) (*entity.Appointment, error) {
	apt, err := u.aptRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, errors.New("appointment not found")
	}

	_, isProvider := u.checkAppointmentParties(ctx, apt, providerID)
	if !isProvider {
		return nil, errors.New("only the provider can send arrival notification")
	}

	if apt.Status == "COMPLETED" || apt.Status == "CANCELLED" {
		return nil, fmt.Errorf("appointment is already %s", strings.ToLower(apt.Status))
	}

	providerName := "Your service provider"
	if pProfile, err := u.profileRepo.GetByID(ctx, providerID); err == nil && pProfile != nil && pProfile.FirstName != "" {
		if pProfile.LastName != "" {
			providerName = fmt.Sprintf("%s %s", pProfile.FirstName, pProfile.LastName)
		} else {
			providerName = pProfile.FirstName
		}
	}

	now := time.Now().UTC()
	apt.Status = "ARRIVED"
	apt.UpdatedAt = now
	_ = u.aptRepo.Update(ctx, apt)

	title := "📍 Provider Has Arrived!"
	msg := fmt.Sprintf("%s has arrived at your location for '%s'. Please meet them outside.", providerName, apt.Title)

	if u.notifRepo != nil {
		notif := entity.Notification{
			ID:               uuid.New(),
			UserID:           apt.SeekerID,
			SenderID:         &providerID,
			NotificationType: "TRACKING",
			Title:            title,
			Message:          msg,
			Data: entity.JSONMap{
				"notification_type": "provider_arrived",
				"appointment_id":    apt.ID.String(),
				"request_id":        apt.ID.String(),
				"provider_lat":      providerLat,
				"provider_lng":      providerLng,
				"arrived_at":        now.Format(time.RFC3339),
			},
			CreatedAt: now,
		}
		_ = u.notifRepo.Create(ctx, &notif)
	}

	u.sendPush(apt.SeekerID, title, msg, map[string]string{
		"notification_type": "provider_arrived",
		"appointment_id":    apt.ID.String(),
		"request_id":        apt.ID.String(),
		"sender_id":         providerID.String(),
	})

	return apt, nil
}

func (u *interactionUseCase) CompleteAppointment(ctx context.Context, userID, appointmentID uuid.UUID, amount float64) (float64, float64, error) {
	apt, err := u.aptRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return 0, 0, errors.New("appointment not found")
	}

	isSeeker, isProvider := u.checkAppointmentParties(ctx, apt, userID)
	if !isSeeker && !isProvider {
		return 0, 0, errors.New("not authorized")
	}

	if apt.Status == "COMPLETED" {
		return 0, 0, errors.New("appointment already completed")
	}

	apt.Status = "COMPLETED"
	apt.UpdatedAt = time.Now()
	_ = u.aptRepo.Update(ctx, apt)

	// Sync linked ServiceRequest to DONE
	if apt.ServiceRequestID != nil && *apt.ServiceRequestID != uuid.Nil && u.requestRepo != nil {
		if req, rErr := u.requestRepo.GetByID(ctx, *apt.ServiceRequestID); rErr == nil && req != nil {
			req.Status = "DONE"
			req.UpdatedAt = time.Now()
			_ = u.requestRepo.Update(ctx, req)
		}
	}

	otherPartyID := apt.ProviderID
	if isProvider {
		otherPartyID = apt.SeekerID
	}
	if u.notifRepo != nil {
		notif := entity.Notification{
			ID:               uuid.New(),
			UserID:           otherPartyID,
			SenderID:         &userID,
			NotificationType: "APPOINTMENT",
			Title:            "Appointment Completed",
			Message:          fmt.Sprintf("The appointment for '%s' has been marked as completed.", apt.Title),
			Data:             entity.JSONMap{"appointment_id": apt.ID.String()},
			CreatedAt:        time.Now(),
		}
		_ = u.notifRepo.Create(ctx, &notif)
	}
	u.sendPush(otherPartyID, "Appointment Completed", fmt.Sprintf("The appointment for '%s' has been marked as completed.", apt.Title), map[string]string{
		"notification_type": "appointment",
		"appointment_id":    apt.ID.String(),
		"sender_id":         userID.String(),
	})

	// Real-time WebSocket broadcast to both seeker and provider
	websocket.GlobalHub.SendToUser(apt.SeekerID.String(), map[string]interface{}{
		"type":           "appointment_completed",
		"appointment_id": apt.ID.String(),
		"status":         "COMPLETED",
		"title":          "Appointment Completed",
		"message":        fmt.Sprintf("The appointment for '%s' has been marked as completed.", apt.Title),
		"created_at":     time.Now().UTC(),
	})
	websocket.GlobalHub.SendToUser(apt.ProviderID.String(), map[string]interface{}{
		"type":           "appointment_completed",
		"appointment_id": apt.ID.String(),
		"status":         "COMPLETED",
		"title":          "Appointment Completed",
		"message":        fmt.Sprintf("The appointment for '%s' has been marked as completed.", apt.Title),
		"created_at":     time.Now().UTC(),
	})

	// Award XP
	if provProfile, pErr := u.profileRepo.GetByUserID(ctx, apt.ProviderID); pErr == nil && provProfile != nil {
		provProfile.XP += 200
		_ = u.profileRepo.Update(ctx, provProfile)
	}
	if seekerProfile, sErr := u.profileRepo.GetByUserID(ctx, apt.SeekerID); sErr == nil && seekerProfile != nil {
		seekerProfile.XP += 50
		_ = u.profileRepo.Update(ctx, seekerProfile)
	}

	var netAmount float64
	var commission float64

	if apt.IsFunded {
		total := apt.TotalPrice
		if amount > 0 {
			total = amount
		}

		// Tier rates: NONE=20%, SILVER=15%, GOLD=10%, PLATINUM=5%
		tier := "NONE"
		if p, err := u.profileRepo.GetByUserID(ctx, apt.ProviderID); err == nil && p != nil {
			tier = p.SubscriptionTier
		}

		rate := 0.20
		switch tier {
		case "PLATINUM":
			rate = 0.05
		case "GOLD":
			rate = 0.10
		case "SILVER":
			rate = 0.15
		default:
			rate = 0.20
		}

		commission = total * rate
		netAmount = total - commission

		// Credit Provider Wallet
		wallet, _ := u.walletRepo.GetByUserID(ctx, apt.ProviderID)
		if wallet != nil {
			wallet.Balance += netAmount
			_ = u.walletRepo.Update(ctx, wallet)

			tx := entity.WalletTransaction{
				ID:              uuid.New(),
				WalletID:        wallet.ID,
				Amount:          netAmount,
				TransactionType: "CREDIT",
				Description:     fmt.Sprintf("Job Completion (Funded): %s", apt.Title),
				Status:          "COMPLETED",
				ReferenceID:     apt.ID.String(),
				CreatedAt:       time.Now(),
			}
			_ = u.txRepo.Create(ctx, &tx)
		}
	}

	// Asynchronously send completion emails
	if u.cfg != nil && u.cfg.SMTPHost != "" && u.userRepo != nil {
		go func() {
			sUser, sErr := u.userRepo.GetByID(context.Background(), apt.SeekerID)
			pUser, pErr := u.userRepo.GetByID(context.Background(), apt.ProviderID)
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

				emailCfg := &email.Config{
					Host:     u.cfg.SMTPHost,
					Port:     u.cfg.SMTPPort,
					User:     u.cfg.SMTPUser,
					Password: u.cfg.SMTPPassword,
					From:     u.cfg.EmailFrom,
				}

				_ = email.SendAppointmentCompletedEmail(emailCfg, sUser.Email, seekerName, providerName, apt.Title, apt.TotalPrice)
				_ = email.SendAppointmentCompletedEmail(emailCfg, pUser.Email, providerName, seekerName, apt.Title, apt.TotalPrice)
			}
		}()
	}

	return netAmount, commission, nil
}

func (u *interactionUseCase) CancelAppointment(ctx context.Context, userID, appointmentID uuid.UUID) error {
	apt, err := u.aptRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return errors.New("appointment not found")
	}

	isSeeker, isProvider := u.checkAppointmentParties(ctx, apt, userID)
	if !isSeeker && !isProvider {
		return errors.New("not authorized")
	}

	apt.Status = "CANCELLED"
	apt.UpdatedAt = time.Now()
	if err := u.aptRepo.Update(ctx, apt); err != nil {
		return err
	}

	// If linked service request was in progress, reset to OPEN so proposals can be accepted or re-scheduled
	if apt.ServiceRequestID != nil && *apt.ServiceRequestID != uuid.Nil && u.requestRepo != nil {
		if req, rErr := u.requestRepo.GetByID(ctx, *apt.ServiceRequestID); rErr == nil && req != nil {
			if req.Status == "IN_PROGRESS" {
				req.Status = "OPEN"
				req.UpdatedAt = time.Now()
				_ = u.requestRepo.Update(ctx, req)
			}
		}
	}

	cancelOtherID := apt.ProviderID
	if isProvider {
		cancelOtherID = apt.SeekerID
	}
	if u.notifRepo != nil {
		notif := entity.Notification{
			ID:               uuid.New(),
			UserID:           cancelOtherID,
			SenderID:         &userID,
			NotificationType: "APPOINTMENT",
			Title:            "Appointment Cancelled",
			Message:          fmt.Sprintf("The appointment for '%s' has been cancelled.", apt.Title),
			Data:             entity.JSONMap{"appointment_id": apt.ID.String()},
			CreatedAt:        time.Now(),
		}
		_ = u.notifRepo.Create(ctx, &notif)
	}
	u.sendPush(cancelOtherID, "Appointment Cancelled", fmt.Sprintf("The appointment for '%s' has been cancelled.", apt.Title), map[string]string{
		"notification_type": "appointment",
		"appointment_id":    apt.ID.String(),
		"sender_id":         userID.String(),
	})

	// Real-time WebSocket broadcast to both seeker and provider
	websocket.GlobalHub.SendToUser(apt.SeekerID.String(), map[string]interface{}{
		"type":           "appointment_cancelled",
		"appointment_id": apt.ID.String(),
		"status":         "CANCELLED",
		"title":          "Appointment Cancelled",
		"message":        fmt.Sprintf("The appointment for '%s' has been cancelled.", apt.Title),
		"created_at":     time.Now().UTC(),
	})
	websocket.GlobalHub.SendToUser(apt.ProviderID.String(), map[string]interface{}{
		"type":           "appointment_cancelled",
		"appointment_id": apt.ID.String(),
		"status":         "CANCELLED",
		"title":          "Appointment Cancelled",
		"message":        fmt.Sprintf("The appointment for '%s' has been cancelled.", apt.Title),
		"created_at":     time.Now().UTC(),
	})

	// Asynchronously send cancellation emails
	if u.cfg != nil && u.cfg.SMTPHost != "" && u.userRepo != nil {
		go func() {
			sUser, sErr := u.userRepo.GetByID(context.Background(), apt.SeekerID)
			pUser, pErr := u.userRepo.GetByID(context.Background(), apt.ProviderID)
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

				emailCfg := &email.Config{
					Host:     u.cfg.SMTPHost,
					Port:     u.cfg.SMTPPort,
					User:     u.cfg.SMTPUser,
					Password: u.cfg.SMTPPassword,
					From:     u.cfg.EmailFrom,
				}

				_ = email.SendAppointmentCancelledEmail(emailCfg, sUser.Email, seekerName, providerName, apt.Title, "Cancelled by user")
				_ = email.SendAppointmentCancelledEmail(emailCfg, pUser.Email, providerName, seekerName, apt.Title, "Cancelled by user")
			}
		}()
	}

	return nil
}

func (u *interactionUseCase) GetDisputes(ctx context.Context, userID uuid.UUID) ([]entity.Dispute, error) {
	return u.disputeRepo.ListByUser(ctx, userID)
}

func (u *interactionUseCase) CreateDispute(ctx context.Context, initiatorID uuid.UUID, dispute *entity.Dispute) (*entity.Dispute, error) {
	if dispute.AppointmentID == nil || *dispute.AppointmentID == uuid.Nil {
		return nil, errors.New("appointment_id is required")
	}

	dispute.ID = uuid.New()
	dispute.RaisedByID = initiatorID
	if dispute.Status == "" {
		dispute.Status = "OPEN"
	}
	dispute.CreatedAt = time.Now()
	dispute.UpdatedAt = time.Now()

	if err := u.disputeRepo.Create(ctx, dispute); err != nil {
		return nil, err
	}

	if u.aptRepo != nil {
		if apt, aErr := u.aptRepo.GetByID(ctx, *dispute.AppointmentID); aErr == nil && apt != nil {
			otherPartyID := apt.ProviderID
			if initiatorID == apt.ProviderID {
				otherPartyID = apt.SeekerID
			}
			if u.notifRepo != nil {
				notif := entity.Notification{
					ID:               uuid.New(),
					UserID:           otherPartyID,
					SenderID:         &initiatorID,
					NotificationType: "DISPUTE",
					Title:            "Dispute Filed",
					Message:          fmt.Sprintf("A dispute was filed regarding '%s'.", apt.Title),
					Data:             entity.JSONMap{"dispute_id": dispute.ID.String(), "appointment_id": apt.ID.String()},
					CreatedAt:        time.Now(),
				}
				_ = u.notifRepo.Create(ctx, &notif)
			}
			u.sendPush(otherPartyID, "Dispute Filed", fmt.Sprintf("A dispute was filed regarding '%s'.", apt.Title), map[string]string{
				"notification_type": "dispute",
				"dispute_id":        dispute.ID.String(),
				"appointment_id":    apt.ID.String(),
			})
		}
	}

	// Asynchronously send dispute filed email to both parties
	if u.cfg != nil && u.cfg.SMTPHost != "" && u.userRepo != nil && u.aptRepo != nil {
		go func() {
			apt, aErr := u.aptRepo.GetByID(context.Background(), *dispute.AppointmentID)
			if aErr == nil && apt != nil {
				sUser, _ := u.userRepo.GetByID(context.Background(), apt.SeekerID)
				pUser, _ := u.userRepo.GetByID(context.Background(), apt.ProviderID)
				emailCfg := &email.Config{
					Host:     u.cfg.SMTPHost,
					Port:     u.cfg.SMTPPort,
					User:     u.cfg.SMTPUser,
					Password: u.cfg.SMTPPassword,
					From:     u.cfg.EmailFrom,
				}

				if sUser != nil {
					_ = email.SendDisputeFiledEmail(emailCfg, sUser.Email, sUser.Email, apt.ID.String(), dispute.Reason)
				}
				if pUser != nil {
					_ = email.SendDisputeFiledEmail(emailCfg, pUser.Email, pUser.Email, apt.ID.String(), dispute.Reason)
				}
			}
		}()
	}

	return u.disputeRepo.GetByID(ctx, dispute.ID)
}

func (u *interactionUseCase) UploadDisputeEvidence(ctx context.Context, userID, disputeID uuid.UUID, evidenceURL string) error {
	dispute, err := u.disputeRepo.GetByID(ctx, disputeID)
	if err != nil || dispute == nil {
		return errors.New("dispute not found")
	}

	if dispute.RaisedByID != userID && (dispute.DefendantID == nil || *dispute.DefendantID != userID) {
		return errors.New("not authorized")
	}

	dispute.Evidence = evidenceURL
	dispute.UpdatedAt = time.Now()
	return u.disputeRepo.Update(ctx, dispute)
}
