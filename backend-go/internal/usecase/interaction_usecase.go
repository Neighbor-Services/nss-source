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
		notifRepo:   notifRepo,
		tokenRepo:   tokenRepo,
		fcmClient:   fcmClient,
		cfg:         cfg,
	}
}

func (u *interactionUseCase) sendPush(userID uuid.UUID, title, body string, data map[string]string) {
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
	switch userType {
	case "CUSTOMER", "SEEKER":
		return u.aptRepo.List(ctx, &userID, nil, status)
	case "PROVIDER":
		return u.aptRepo.List(ctx, nil, &userID, status)
	}
	return u.aptRepo.List(ctx, nil, nil, status)
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

	if apt.ProviderID != providerID {
		return nil, errors.New("only the provider can verify the arrival code")
	}

	if apt.Status == "COMPLETED" {
		return nil, errors.New("appointment is already completed")
	}

	if strings.TrimSpace(code) != strings.TrimSpace(apt.SecretCode) {
		return nil, errors.New("invalid verification code. Please check with the seeker.")
	}

	apt.Status = "IN_PROGRESS"
	apt.UpdatedAt = time.Now()
	if err := u.aptRepo.Update(ctx, apt); err != nil {
		return nil, err
	}

	if u.notifRepo != nil {
		notif := entity.Notification{
			ID:               uuid.New(),
			UserID:           apt.SeekerID,
			SenderID:         &providerID,
			NotificationType: "APPOINTMENT",
			Title:            "Appointment Code Verified",
			Message:          "Your provider has verified the arrival code! Appointment started.",
			Data:             entity.JSONMap{"appointment_id": apt.ID.String()},
			CreatedAt:        time.Now(),
		}
		_ = u.notifRepo.Create(ctx, &notif)
	}
	u.sendPush(apt.SeekerID, "Appointment Code Verified", "Your provider has verified the arrival code! Appointment started.", map[string]string{
		"notification_type": "appointment",
		"appointment_id":    apt.ID.String(),
		"sender_id":         providerID.String(),
	})

	return apt, nil
}

func (u *interactionUseCase) CompleteAppointment(ctx context.Context, userID, appointmentID uuid.UUID, amount float64) (float64, float64, error) {
	apt, err := u.aptRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return 0, 0, errors.New("appointment not found")
	}

	if apt.SeekerID != userID && apt.ProviderID != userID {
		return 0, 0, errors.New("not authorized")
	}

	if apt.Status == "COMPLETED" {
		return 0, 0, errors.New("appointment already completed")
	}

	apt.Status = "COMPLETED"
	apt.UpdatedAt = time.Now()
	_ = u.aptRepo.Update(ctx, apt)

	otherPartyID := apt.ProviderID
	if userID == apt.ProviderID {
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

	if apt.SeekerID != userID && apt.ProviderID != userID {
		return errors.New("not authorized")
	}

	apt.Status = "CANCELLED"
	apt.UpdatedAt = time.Now()
	if err := u.aptRepo.Update(ctx, apt); err != nil {
		return err
	}

	cancelOtherID := apt.ProviderID
	if userID == apt.ProviderID {
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
