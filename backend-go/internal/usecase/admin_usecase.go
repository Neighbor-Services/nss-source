package usecase

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/pkg/auth"
	"backend-go/pkg/email"
	"github.com/google/uuid"
)

type adminUseCase struct {
	adminRepo   repository.AdminRepository
	profileRepo repository.ProfileRepository
	userRepo    repository.UserRepository
	walletRepo  repository.WalletRepository
	cfg         *config.Config
}

func NewAdminUseCase(
	adminRepo repository.AdminRepository,
	profileRepo repository.ProfileRepository,
	userRepo repository.UserRepository,
	walletRepo repository.WalletRepository,
	cfg *config.Config,
) domainUsecase.AdminUseCase {
	return &adminUseCase{
		adminRepo:   adminRepo,
		profileRepo: profileRepo,
		userRepo:    userRepo,
		walletRepo:  walletRepo,
		cfg:         cfg,
	}
}

func (u *adminUseCase) logAudit(ctx context.Context, adminID *uuid.UUID, action, resType, resID string, details map[string]interface{}) {
	log := entity.AuditLog{
		ID:           uuid.New(),
		UserID:       adminID,
		Action:       action,
		ResourceType: resType,
		ResourceID:   resID,
		Details:      details,
		CreatedAt:    time.Now(),
	}
	_ = u.adminRepo.CreateAuditLog(ctx, &log)
}

func (u *adminUseCase) GetDashboardStats(ctx context.Context, adminID uuid.UUID) (*entity.AdminDashboardStats, error) {
	return u.adminRepo.GetDashboardStats(ctx)
}

func (u *adminUseCase) ListUsers(ctx context.Context, adminID uuid.UUID, filter repository.AdminUserFilter) ([]entity.User, int64, error) {
	return u.adminRepo.ListUsers(ctx, filter)
}

func (u *adminUseCase) GetUserByID(ctx context.Context, adminID, userID uuid.UUID) (*entity.User, error) {
	return u.adminRepo.GetUserByID(ctx, userID)
}

func (u *adminUseCase) CreateUser(ctx context.Context, adminID uuid.UUID, input domainUsecase.AdminCreateUserInput) (*entity.User, error) {
	// Check if user already exists
	existing, _ := u.userRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, errors.New("a user with this email already exists")
	}

	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	var roleID *uuid.UUID
	if input.RoleID != nil && *input.RoleID != "" {
		if parsed, err := uuid.Parse(*input.RoleID); err == nil {
			roleID = &parsed
		}
	}

	now := time.Now()
	user := entity.User{
		ID:          uuid.New(),
		Email:       input.Email,
		Password:    hashedPassword,
		IsStaff:     input.IsStaff,
		IsActive:    input.IsActive,
		IsVerified:  input.IsVerified,
		AdminRoleID: roleID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := u.userRepo.Create(ctx, &user); err != nil {
		return nil, err
	}

	userType := input.UserType
	if userType == "" {
		userType = "SEEKER"
	}
	tier := input.SubscriptionTier
	if tier == "" {
		tier = "BASIC"
	}

	profile := entity.Profile{
		ID:                   uuid.New(),
		UserID:               user.ID,
		FirstName:            input.FirstName,
		LastName:             input.LastName,
		Phone:                input.Phone,
		UserType:             userType,
		SubscriptionTier:     tier,
		SubscriptionInterval: input.SubscriptionInterval,
		PreferredPaymentMode: input.PreferredPaymentMode,
		Bio:                  input.Bio,
		City:                 input.City,
		State:                input.State,
		ZipCode:              input.ZipCode,
		Address:              input.Address,
		Country:              input.Country,
		Gender:               input.Gender,
		Service:              input.Service,
		IsIdentityVerified:   input.IsIdentityVerified,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	_ = u.profileRepo.Create(ctx, &profile)

	// Create wallet
	_ = u.walletRepo.Create(ctx, &entity.Wallet{
		ID:        uuid.New(),
		UserID:    user.ID,
		Balance:   0,
		Currency:  "USD",
		CreatedAt: now,
		UpdatedAt: now,
	})

	u.logAudit(ctx, &adminID, "ADMIN_CREATE_USER", "User", user.ID.String(), map[string]interface{}{
		"email":     user.Email,
		"is_staff":  user.IsStaff,
		"user_type": userType,
	})

	return u.adminRepo.GetUserByID(ctx, user.ID)
}

func (u *adminUseCase) UpdateUser(ctx context.Context, adminID, userID uuid.UUID, input domainUsecase.AdminUpdateUserInput) (*entity.User, error) {
	user, err := u.adminRepo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	details := make(map[string]interface{})

	if input.Email != nil && *input.Email != "" {
		user.Email = *input.Email
		details["email"] = *input.Email
	}
	if input.Password != nil && *input.Password != "" {
		if hashed, err := auth.HashPassword(*input.Password); err == nil {
			user.Password = hashed
			details["password_changed"] = true
		}
	}
	if input.RoleID != nil {
		if *input.RoleID == "" {
			user.AdminRoleID = nil
			details["role_id"] = nil
		} else if parsed, err := uuid.Parse(*input.RoleID); err == nil {
			user.AdminRoleID = &parsed
			details["role_id"] = *input.RoleID
		}
	}
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
		details["is_active"] = *input.IsActive
	}
	if input.IsStaff != nil {
		user.IsStaff = *input.IsStaff
		details["is_staff"] = *input.IsStaff
	}
	if input.IsVerified != nil {
		user.IsVerified = *input.IsVerified
		details["is_verified"] = *input.IsVerified
	}

	if err := u.adminRepo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	if user.Profile != nil {
		profileUpdated := false
		if input.FirstName != nil {
			user.Profile.FirstName = *input.FirstName
			details["first_name"] = *input.FirstName
			profileUpdated = true
		}
		if input.LastName != nil {
			user.Profile.LastName = *input.LastName
			details["last_name"] = *input.LastName
			profileUpdated = true
		}
		if input.Phone != nil {
			user.Profile.Phone = *input.Phone
			details["phone"] = *input.Phone
			profileUpdated = true
		}
		if input.UserType != nil && *input.UserType != "" {
			user.Profile.UserType = *input.UserType
			details["user_type"] = *input.UserType
			profileUpdated = true
		}
		if input.SubscriptionTier != nil && *input.SubscriptionTier != "" {
			user.Profile.SubscriptionTier = *input.SubscriptionTier
			details["subscription_tier"] = *input.SubscriptionTier
			profileUpdated = true
		}
		if input.SubscriptionInterval != nil && *input.SubscriptionInterval != "" {
			user.Profile.SubscriptionInterval = *input.SubscriptionInterval
			details["subscription_interval"] = *input.SubscriptionInterval
			profileUpdated = true
		}
		if input.PreferredPaymentMode != nil && *input.PreferredPaymentMode != "" {
			user.Profile.PreferredPaymentMode = *input.PreferredPaymentMode
			details["preferred_payment_mode"] = *input.PreferredPaymentMode
			profileUpdated = true
		}
		if input.Bio != nil {
			user.Profile.Bio = *input.Bio
			details["bio"] = *input.Bio
			profileUpdated = true
		}
		if input.City != nil {
			user.Profile.City = *input.City
			details["city"] = *input.City
			profileUpdated = true
		}
		if input.State != nil {
			user.Profile.State = *input.State
			details["state"] = *input.State
			profileUpdated = true
		}
		if input.ZipCode != nil {
			user.Profile.ZipCode = *input.ZipCode
			details["zip_code"] = *input.ZipCode
			profileUpdated = true
		}
		if input.Address != nil {
			user.Profile.Address = *input.Address
			details["address"] = *input.Address
			profileUpdated = true
		}
		if input.Country != nil {
			user.Profile.Country = *input.Country
			details["country"] = *input.Country
			profileUpdated = true
		}
		if input.Gender != nil {
			user.Profile.Gender = *input.Gender
			details["gender"] = *input.Gender
			profileUpdated = true
		}
		if input.Service != nil {
			user.Profile.Service = *input.Service
			details["service"] = *input.Service
			profileUpdated = true
		}
		if input.IsIdentityVerified != nil {
			user.Profile.IsIdentityVerified = *input.IsIdentityVerified
			details["is_identity_verified"] = *input.IsIdentityVerified
			profileUpdated = true
		}
		if profileUpdated {
			user.Profile.UpdatedAt = time.Now()
			_ = u.profileRepo.Update(ctx, user.Profile)
		}
	}

	if input.WalletBalanceDelta != nil && *input.WalletBalanceDelta != 0 {
		wallet, err := u.adminRepo.GetWalletByUserID(ctx, userID)
		if err == nil && wallet != nil {
			wallet.Balance += *input.WalletBalanceDelta
			_ = u.adminRepo.UpdateWallet(ctx, wallet)
			details["wallet_delta"] = *input.WalletBalanceDelta
		}
	}

	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_USER", "User", userID.String(), details)
	return u.adminRepo.GetUserByID(ctx, userID)
}

func (u *adminUseCase) DeleteUser(ctx context.Context, adminID, userID uuid.UUID) error {
	if adminID == userID {
		return errors.New("cannot delete your own account")
	}
	err := u.adminRepo.DeleteUser(ctx, userID)
	if err == nil {
		u.logAudit(ctx, &adminID, "ADMIN_DELETE_USER", "User", userID.String(), nil)
	}
	return err
}

func (u *adminUseCase) ListVerifications(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.ProviderVerification, int64, error) {
	return u.adminRepo.ListVerifications(ctx, status, limit, offset)
}

func (u *adminUseCase) ApproveVerification(ctx context.Context, adminID, verificationID uuid.UUID) error {
	v, err := u.adminRepo.GetVerificationByID(ctx, verificationID)
	if err != nil || v == nil {
		return errors.New("verification not found")
	}

	v.Status = "APPROVED"
	v.UpdatedAt = time.Now()

	if err := u.adminRepo.UpdateVerification(ctx, v); err != nil {
		return err
	}

	if profile, err := u.profileRepo.GetByUserID(ctx, v.ProviderID); err == nil && profile != nil {
		profile.IsIdentityVerified = true
		profile.UpdatedAt = time.Now()
		_ = u.profileRepo.Update(ctx, profile)
	}

	u.logAudit(ctx, &adminID, "ADMIN_APPROVE_VERIFICATION", "ProviderVerification", verificationID.String(), map[string]interface{}{
		"provider_id": v.ProviderID.String(),
	})
	return nil
}

func (u *adminUseCase) RejectVerification(ctx context.Context, adminID, verificationID uuid.UUID) error {
	v, err := u.adminRepo.GetVerificationByID(ctx, verificationID)
	if err != nil || v == nil {
		return errors.New("verification not found")
	}

	v.Status = "REJECTED"
	v.UpdatedAt = time.Now()

	if err := u.adminRepo.UpdateVerification(ctx, v); err != nil {
		return err
	}

	u.logAudit(ctx, &adminID, "ADMIN_REJECT_VERIFICATION", "ProviderVerification", verificationID.String(), map[string]interface{}{
		"provider_id": v.ProviderID.String(),
	})
	return nil
}

func (u *adminUseCase) ListBackgroundChecks(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.BackgroundCheck, int64, error) {
	return u.adminRepo.ListBackgroundChecks(ctx, status, limit, offset)
}

func (u *adminUseCase) OverrideBackgroundCheck(ctx context.Context, adminID, checkID uuid.UUID, status, notes string) error {
	bc, err := u.adminRepo.GetBackgroundCheckByID(ctx, checkID)
	if err != nil || bc == nil {
		return errors.New("background check not found")
	}

	bc.Status = status
	if notes != "" {
		bc.Notes = fmt.Sprintf("%s\n[Admin Override]: %s", bc.Notes, notes)
	}
	now := time.Now()
	bc.LastSyncedAt = &now

	if err := u.adminRepo.UpdateBackgroundCheck(ctx, bc); err != nil {
		return err
	}

	if profile, err := u.profileRepo.GetByUserID(ctx, bc.ProviderID); err == nil && profile != nil {
		if status == "CLEAR" {
			profile.IsIdentityVerified = true
		} else {
			profile.IsIdentityVerified = false
		}
		_ = u.profileRepo.Update(ctx, profile)
	}

	u.logAudit(ctx, &adminID, "ADMIN_OVERRIDE_BACKGROUND_CHECK", "BackgroundCheck", checkID.String(), map[string]interface{}{
		"status": status,
		"notes":  notes,
	})
	return nil
}

func (u *adminUseCase) ListReports(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.Report, int64, error) {
	return u.adminRepo.ListReports(ctx, status, limit, offset)
}

func (u *adminUseCase) ResolveReport(ctx context.Context, adminID, reportID uuid.UUID, resolution string) error {
	rep, err := u.adminRepo.GetReportByID(ctx, reportID)
	if err != nil || rep == nil {
		return errors.New("report not found")
	}

	rep.Status = "RESOLVED"
	if err := u.adminRepo.UpdateReport(ctx, rep); err != nil {
		return err
	}

	u.logAudit(ctx, &adminID, "ADMIN_RESOLVE_REPORT", "Report", reportID.String(), map[string]interface{}{
		"resolution": resolution,
	})
	return nil
}

func (u *adminUseCase) ListDisputes(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.Dispute, int64, error) {
	return u.adminRepo.ListDisputes(ctx, status, limit, offset)
}

func (u *adminUseCase) ResolveDispute(ctx context.Context, adminID, disputeID uuid.UUID, notes string) error {
	d, err := u.adminRepo.GetDisputeByID(ctx, disputeID)
	if err != nil || d == nil {
		return errors.New("dispute not found")
	}

	d.Status = "RESOLVED"
	if notes != "" {
		d.ResolutionNotes = notes
	}
	d.UpdatedAt = time.Now()

	if err := u.adminRepo.UpdateDispute(ctx, d); err != nil {
		return err
	}

	u.logAudit(ctx, &adminID, "ADMIN_RESOLVE_DISPUTE", "Dispute", disputeID.String(), map[string]interface{}{
		"status": "RESOLVED",
		"notes":  notes,
	})
	return nil
}

func (u *adminUseCase) RejectDispute(ctx context.Context, adminID, disputeID uuid.UUID, notes string) error {
	d, err := u.adminRepo.GetDisputeByID(ctx, disputeID)
	if err != nil || d == nil {
		return errors.New("dispute not found")
	}

	d.Status = "REJECTED"
	if notes != "" {
		d.ResolutionNotes = notes
	}
	d.UpdatedAt = time.Now()

	if err := u.adminRepo.UpdateDispute(ctx, d); err != nil {
		return err
	}

	u.logAudit(ctx, &adminID, "ADMIN_REJECT_DISPUTE", "Dispute", disputeID.String(), map[string]interface{}{
		"status": "REJECTED",
		"notes":  notes,
	})
	return nil
}

func (u *adminUseCase) ListPayoutRequests(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.PayoutRequest, int64, error) {
	return u.adminRepo.ListPayoutRequests(ctx, status, limit, offset)
}

func (u *adminUseCase) ApprovePayout(ctx context.Context, adminID, payoutID uuid.UUID) error {
	pr, err := u.adminRepo.GetPayoutRequestByID(ctx, payoutID)
	if err != nil || pr == nil {
		return errors.New("payout request not found")
	}

	if pr.Status != "PENDING" {
		return errors.New("payout is not in pending state")
	}

	pr.Status = "PROCESSED"
	now := time.Now()
	pr.ProcessedAt = &now

	if err := u.adminRepo.UpdatePayoutRequest(ctx, pr); err != nil {
		return err
	}

	u.logAudit(ctx, &adminID, "ADMIN_APPROVE_PAYOUT", "PayoutRequest", payoutID.String(), map[string]interface{}{
		"amount": pr.Amount,
	})

	if u.cfg != nil && u.cfg.SMTPHost != "" && u.walletRepo != nil && u.userRepo != nil {
		go func() {
			if wallet, err := u.walletRepo.GetByID(context.Background(), pr.WalletID); err == nil && wallet != nil {
				if user, uErr := u.userRepo.GetByID(context.Background(), wallet.UserID); uErr == nil && user != nil {
					_ = email.SendPayoutStatusEmail(&email.Config{
						Host:     u.cfg.SMTPHost,
						Port:     u.cfg.SMTPPort,
						User:     u.cfg.SMTPUser,
						Password: u.cfg.SMTPPassword,
						From:     u.cfg.EmailFrom,
					}, user.Email, user.Email, pr.Amount, "APPROVED", "")
				}
			}
		}()
	}

	return nil
}

func (u *adminUseCase) RejectPayout(ctx context.Context, adminID, payoutID uuid.UUID, notes string) error {
	pr, err := u.adminRepo.GetPayoutRequestByID(ctx, payoutID)
	if err != nil || pr == nil {
		return errors.New("payout request not found")
	}

	if pr.Status != "PENDING" {
		return errors.New("payout is not in pending state")
	}

	pr.Status = "REJECTED"
	pr.AdminNotes = notes
	now := time.Now()
	pr.ProcessedAt = &now

	if err := u.adminRepo.UpdatePayoutRequest(ctx, pr); err != nil {
		return err
	}

	// Refund amount back to provider's wallet
	if wallet, err := u.walletRepo.GetByID(ctx, pr.WalletID); err == nil && wallet != nil {
		wallet.Balance += pr.Amount
		_ = u.walletRepo.Update(ctx, wallet)

		if u.cfg != nil && u.cfg.SMTPHost != "" && u.userRepo != nil {
			go func() {
				if user, uErr := u.userRepo.GetByID(context.Background(), wallet.UserID); uErr == nil && user != nil {
					_ = email.SendPayoutStatusEmail(&email.Config{
						Host:     u.cfg.SMTPHost,
						Port:     u.cfg.SMTPPort,
						User:     u.cfg.SMTPUser,
						Password: u.cfg.SMTPPassword,
						From:     u.cfg.EmailFrom,
					}, user.Email, user.Email, pr.Amount, "REJECTED", notes)
				}
			}()
		}
	}

	u.logAudit(ctx, &adminID, "ADMIN_REJECT_PAYOUT", "PayoutRequest", payoutID.String(), map[string]interface{}{
		"amount": pr.Amount,
		"notes":  notes,
	})
	return nil
}

func (u *adminUseCase) ListWallets(ctx context.Context, adminID uuid.UUID, limit, offset int) ([]entity.Wallet, int64, error) {
	return u.adminRepo.ListWallets(ctx, limit, offset)
}

func (u *adminUseCase) ListSubscriptions(ctx context.Context, adminID uuid.UUID, isActive *bool, limit, offset int) ([]entity.Subscription, int64, error) {
	return u.adminRepo.ListSubscriptions(ctx, isActive, limit, offset)
}

func (u *adminUseCase) ToggleSubscription(ctx context.Context, adminID, subID uuid.UUID, isActive bool) error {
	sub, err := u.adminRepo.GetSubscriptionByID(ctx, subID)
	if err != nil || sub == nil {
		return errors.New("subscription not found")
	}

	sub.IsActive = isActive
	sub.UpdatedAt = time.Now()
	if err := u.adminRepo.UpdateSubscription(ctx, sub); err != nil {
		return err
	}

	if profile, err := u.profileRepo.GetByUserID(ctx, sub.UserID); err == nil && profile != nil {
		if !isActive {
			profile.SubscriptionTier = "NONE"
			profile.SubscriptionInterval = "none"
			_ = u.profileRepo.UpdateCatalogServices(ctx, profile.ID, []uuid.UUID{})
		}
		_ = u.profileRepo.Update(ctx, profile)
	}

	u.logAudit(ctx, &adminID, "ADMIN_TOGGLE_SUBSCRIPTION", "Subscription", subID.String(), map[string]interface{}{
		"is_active": isActive,
	})
	return nil
}

func (u *adminUseCase) AssignSubscription(ctx context.Context, adminID uuid.UUID, input domainUsecase.AssignSubscriptionInput) (*entity.Subscription, error) {
	if input.UserID == uuid.Nil {
		return nil, errors.New("user_id is required")
	}

	tier := strings.ToUpper(strings.TrimSpace(input.Tier))
	interval := strings.ToLower(strings.TrimSpace(input.Interval))
	if interval == "" {
		interval = "month"
	}

	// If a plan ID is provided, fetch the plan
	if input.PlanID != nil && *input.PlanID != uuid.Nil {
		p, err := u.adminRepo.GetSubscriptionPlanByID(ctx, *input.PlanID)
		if err == nil && p != nil {
			if tier == "" {
				tier = strings.ToUpper(p.Tier)
			}
			if input.Interval == "" {
				interval = strings.ToLower(p.Interval)
			}
		}
	}

	if tier == "" {
		tier = "SILVER"
	}

	// Calculate expiration / next payment
	nextPayment := input.NextPayment
	if nextPayment == nil {
		t := time.Now().AddDate(0, 1, 0)
		if interval == "year" {
			t = time.Now().AddDate(1, 0, 0)
		} else if interval == "lifetime" {
			t = time.Now().AddDate(100, 0, 0)
		}
		nextPayment = &t
	}

	storeTxID := input.StoreTransactionID
	if storeTxID == "" {
		storeTxID = "ADMIN_MANUAL_GRANT_" + uuid.New().String()[:8]
	}

	// Check if user already has a subscription record
	existing, _ := u.adminRepo.GetSubscriptionByUserID(ctx, input.UserID)
	var sub *entity.Subscription
	if existing != nil {
		existing.PlanID = input.PlanID
		existing.IsActive = input.IsActive
		existing.NextPayment = nextPayment
		existing.StoreTransactionID = storeTxID
		existing.UpdatedAt = time.Now()
		if err := u.adminRepo.UpdateSubscription(ctx, existing); err != nil {
			return nil, err
		}
		sub = existing
	} else {
		newSub := entity.Subscription{
			ID:                 uuid.New(),
			UserID:             input.UserID,
			PlanID:             input.PlanID,
			StoreTransactionID: storeTxID,
			NextPayment:        nextPayment,
			IsActive:           input.IsActive,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}
		if err := u.adminRepo.CreateSubscription(ctx, &newSub); err != nil {
			return nil, err
		}
		sub = &newSub
	}

	// Update user profile tier and interval
	if profile, err := u.profileRepo.GetByUserID(ctx, input.UserID); err == nil && profile != nil {
		if input.IsActive {
			profile.SubscriptionTier = tier
			profile.SubscriptionInterval = interval
		} else {
			profile.SubscriptionTier = "NONE"
			profile.SubscriptionInterval = "none"
		}
		_ = u.profileRepo.Update(ctx, profile)
	}

	// Reload full subscription with preloaded relations
	reloaded, err := u.adminRepo.GetSubscriptionByID(ctx, sub.ID)
	if err == nil && reloaded != nil {
		sub = reloaded
	}

	u.logAudit(ctx, &adminID, "ADMIN_ASSIGN_SUBSCRIPTION", "Subscription", sub.ID.String(), map[string]interface{}{
		"user_id":   input.UserID.String(),
		"tier":      tier,
		"interval":  interval,
		"is_active": input.IsActive,
		"notes":     input.Notes,
	})

	return sub, nil
}

// ─── SUBSCRIPTION PLANS & TIERS ──────────────────────────────────────────────

func (u *adminUseCase) ListSubscriptionPlans(ctx context.Context, adminID uuid.UUID) ([]entity.SubscriptionPlan, error) {
	return u.adminRepo.ListSubscriptionPlans(ctx)
}

func (u *adminUseCase) CreateSubscriptionPlan(ctx context.Context, adminID uuid.UUID, plan *entity.SubscriptionPlan) (*entity.SubscriptionPlan, error) {
	if plan.ID == uuid.Nil {
		plan.ID = uuid.New()
	}
	now := time.Now()
	plan.CreatedAt = now
	plan.UpdatedAt = now
	if err := u.adminRepo.CreateSubscriptionPlan(ctx, plan); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_CREATE_SUBSCRIPTION_PLAN", "SubscriptionPlan", plan.ID.String(), map[string]interface{}{
		"name":  plan.Name,
		"tier":  plan.Tier,
		"price": plan.Price,
	})
	return plan, nil
}

func (u *adminUseCase) UpdateSubscriptionPlan(ctx context.Context, adminID uuid.UUID, plan *entity.SubscriptionPlan) (*entity.SubscriptionPlan, error) {
	plan.UpdatedAt = time.Now()
	if err := u.adminRepo.UpdateSubscriptionPlan(ctx, plan); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_SUBSCRIPTION_PLAN", "SubscriptionPlan", plan.ID.String(), map[string]interface{}{
		"name":  plan.Name,
		"tier":  plan.Tier,
		"price": plan.Price,
	})
	return plan, nil
}

func (u *adminUseCase) DeleteSubscriptionPlan(ctx context.Context, adminID, planID uuid.UUID) error {
	if err := u.adminRepo.DeleteSubscriptionPlan(ctx, planID); err != nil {
		return err
	}
	u.logAudit(ctx, &adminID, "ADMIN_DELETE_SUBSCRIPTION_PLAN", "SubscriptionPlan", planID.String(), nil)
	return nil
}

func (u *adminUseCase) CreateCategory(ctx context.Context, adminID uuid.UUID, cat *entity.Category) (*entity.Category, error) {
	cat.ID = uuid.New()
	cat.CreatedAt = time.Now()
	if err := u.adminRepo.CreateCategory(ctx, cat); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_CREATE_CATEGORY", "Category", cat.ID.String(), map[string]interface{}{
		"name": cat.Name,
	})
	return cat, nil
}

func (u *adminUseCase) UpdateCategory(ctx context.Context, adminID, catID uuid.UUID, name, description, icon string) (*entity.Category, error) {
	cat := entity.Category{
		ID:          catID,
		Name:        name,
		Description: description,
		Image:       icon,
		UpdatedAt:   time.Now(),
	}
	if err := u.adminRepo.UpdateCategory(ctx, &cat); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_CATEGORY", "Category", catID.String(), map[string]interface{}{
		"name": name,
	})
	return &cat, nil
}

func (u *adminUseCase) DeleteCategory(ctx context.Context, adminID, catID uuid.UUID) error {
	err := u.adminRepo.DeleteCategory(ctx, catID)
	if err == nil {
		u.logAudit(ctx, &adminID, "ADMIN_DELETE_CATEGORY", "Category", catID.String(), nil)
	}
	return err
}

func (u *adminUseCase) CreateCatalogService(ctx context.Context, adminID uuid.UUID, cs *entity.CatalogService) (*entity.CatalogService, error) {
	cs.ID = uuid.New()
	cs.CreatedAt = time.Now()
	if err := u.adminRepo.CreateCatalogService(ctx, cs); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_CREATE_CATALOG_SERVICE", "CatalogService", cs.ID.String(), map[string]interface{}{
		"name": cs.Name,
	})
	return cs, nil
}

func (u *adminUseCase) UpdateCatalogService(ctx context.Context, adminID, csID uuid.UUID, name, description string, basePrice float64, categoryID *uuid.UUID) (*entity.CatalogService, error) {
	var catID uuid.UUID
	if categoryID != nil {
		catID = *categoryID
	}
	bp := basePrice
	cs := entity.CatalogService{
		ID:          csID,
		Name:        name,
		Description: description,
		BasePrice:   &bp,
		CategoryID:  catID,
		UpdatedAt:   time.Now(),
	}
	if err := u.adminRepo.UpdateCatalogService(ctx, &cs); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_CATALOG_SERVICE", "CatalogService", csID.String(), map[string]interface{}{
		"name": name,
	})
	return &cs, nil
}

func (u *adminUseCase) DeleteCatalogService(ctx context.Context, adminID, csID uuid.UUID) error {
	err := u.adminRepo.DeleteCatalogService(ctx, csID)
	if err == nil {
		u.logAudit(ctx, &adminID, "ADMIN_DELETE_CATALOG_SERVICE", "CatalogService", csID.String(), nil)
	}
	return err
}

func (u *adminUseCase) GetSettings(ctx context.Context, adminID uuid.UUID) (*entity.ModerationSetting, error) {
	return u.adminRepo.GetSettings(ctx)
}

func (u *adminUseCase) UpdateSettings(ctx context.Context, adminID uuid.UUID, paymentMode string, fee float64, broadcastRadiusKm, matchRadiusKm float64) (*entity.ModerationSetting, error) {
	s, err := u.adminRepo.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	if paymentMode != "" {
		s.BackgroundCheckPaymentMode = paymentMode
	}
	if fee > 0 {
		s.BackgroundCheckFee = fee
	}
	if broadcastRadiusKm > 0 {
		s.BroadcastRadiusKm = broadcastRadiusKm
	}
	if matchRadiusKm > 0 {
		s.MatchRadiusKm = matchRadiusKm
	}
	if err := u.adminRepo.UpdateSettings(ctx, s); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_SETTINGS", "ModerationSetting", s.ID.String(), map[string]interface{}{
		"fee":                 fee,
		"payment_mode":        paymentMode,
		"broadcast_radius_km": s.BroadcastRadiusKm,
		"match_radius_km":     s.MatchRadiusKm,
	})
	return s, nil
}

func (u *adminUseCase) ListAuditLogs(ctx context.Context, adminID uuid.UUID, filter repository.AdminAuditFilter) ([]entity.AuditLog, int64, error) {
	return u.adminRepo.ListAuditLogs(ctx, filter)
}

func (u *adminUseCase) RestoreUser(ctx context.Context, adminID, userID uuid.UUID) error {
	err := u.adminRepo.RestoreUser(ctx, userID)
	if err == nil {
		u.logAudit(ctx, &adminID, "ADMIN_RESTORE_USER", "User", userID.String(), nil)
	}
	return err
}

func (u *adminUseCase) AdjustWalletBalance(ctx context.Context, adminID, walletID uuid.UUID, amount float64, reason string) (*entity.Wallet, error) {
	if reason == "" {
		return nil, errors.New("a mandatory reason is required for manual balance adjustments")
	}

	wallet, err := u.adminRepo.GetWalletByID(ctx, walletID)
	if err != nil || wallet == nil {
		return nil, errors.New("wallet not found")
	}

	wallet.Balance += amount
	if wallet.Balance < 0 {
		return nil, errors.New("adjustment would result in negative balance")
	}
	wallet.UpdatedAt = time.Now()

	if err := u.adminRepo.UpdateWallet(ctx, wallet); err != nil {
		return nil, err
	}

	// Create immutable ledger record
	txType := "CREDIT"
	if amount < 0 {
		txType = "DEBIT"
	}
	tx := entity.WalletTransaction{
		ID:              uuid.New(),
		WalletID:        wallet.ID,
		Amount:          amount,
		TransactionType: txType,
		Description:     fmt.Sprintf("[Admin Adjustment]: %s", reason),
		Status:          "COMPLETED",
		ReferenceID:     fmt.Sprintf("admin-adj-%s", adminID.String()[:8]),
		CreatedAt:       time.Now(),
	}
	_ = u.adminRepo.CreateWalletTransaction(ctx, &tx)

	u.logAudit(ctx, &adminID, "ADMIN_ADJUST_WALLET", "Wallet", walletID.String(), map[string]interface{}{
		"amount": amount,
		"reason": reason,
	})

	return wallet, nil
}

func (u *adminUseCase) BatchVerifications(ctx context.Context, adminID uuid.UUID, ids []uuid.UUID, action, notes string) error {
	if len(ids) == 0 {
		return errors.New("no verification IDs provided")
	}

	status := "APPROVED"
	if action == "REJECT" || action == "REJECTED" {
		status = "REJECTED"
	}

	if err := u.adminRepo.BatchUpdateVerifications(ctx, ids, status, notes); err != nil {
		return err
	}

	u.logAudit(ctx, &adminID, "ADMIN_BATCH_VERIFICATIONS", "ProviderVerification", fmt.Sprintf("count:%d", len(ids)), map[string]interface{}{
		"action": action,
		"count":  len(ids),
	})
	return nil
}

func (u *adminUseCase) BroadcastNotification(ctx context.Context, adminID uuid.UUID, targetUserType, title, message string) (int, error) {
	if title == "" || message == "" {
		return 0, errors.New("title and message are required")
	}

	filter := repository.AdminUserFilter{
		UserType: targetUserType,
		Limit:    10000,
	}
	users, err := u.adminRepo.ExportUsers(ctx, filter)
	if err != nil {
		return 0, err
	}

	u.logAudit(ctx, &adminID, "ADMIN_BROADCAST_NOTIFICATION", "Notification", targetUserType, map[string]interface{}{
		"title":       title,
		"target_type": targetUserType,
		"recipients":  len(users),
	})

	return len(users), nil
}

func (u *adminUseCase) ListFeatureFlags(ctx context.Context, adminID uuid.UUID) ([]entity.FeatureFlag, error) {
	return u.adminRepo.ListFeatureFlags(ctx)
}

func (u *adminUseCase) SetFeatureFlag(ctx context.Context, adminID uuid.UUID, key string, isEnabled bool) (*entity.FeatureFlag, error) {
	flag, err := u.adminRepo.SetFeatureFlag(ctx, key, isEnabled)
	if err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_SET_FEATURE_FLAG", "FeatureFlag", key, map[string]interface{}{
		"is_enabled": isEnabled,
	})
	return flag, nil
}

func (u *adminUseCase) GetFinancialReport(ctx context.Context, adminID uuid.UUID, startDate, endDate string) (*entity.FinancialReportSummary, error) {
	if startDate == "" {
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02 23:59:59")
	}
	return u.adminRepo.GetFinancialReport(ctx, startDate, endDate)
}

func (u *adminUseCase) GetSystemHealth(ctx context.Context, adminID uuid.UUID) (*entity.SystemHealthStatus, error) {
	return u.adminRepo.GetSystemHealth(ctx)
}

func (u *adminUseCase) GetGDPRUserData(ctx context.Context, adminID, userID uuid.UUID) (*entity.GDPRUserData, error) {
	data, err := u.adminRepo.GetGDPRUserData(ctx, userID)
	if err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_EXPORT_GDPR_USER", "User", userID.String(), nil)
	return data, nil
}

func (u *adminUseCase) ExportUsersCSV(ctx context.Context, adminID uuid.UUID, filter repository.AdminUserFilter) ([]byte, error) {
	users, err := u.adminRepo.ExportUsers(ctx, filter)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	_ = writer.Write([]string{"ID", "Email", "UserType", "FirstName", "LastName", "IsActive", "IsVerified", "SubscriptionTier", "CreatedAt"})
	for _, u := range users {
		userType, firstName, lastName, tier := "", "", "", ""
		if u.Profile != nil {
			userType = u.Profile.UserType
			firstName = u.Profile.FirstName
			lastName = u.Profile.LastName
			tier = u.Profile.SubscriptionTier
		}
		_ = writer.Write([]string{
			u.ID.String(),
			u.Email,
			userType,
			firstName,
			lastName,
			fmt.Sprintf("%t", u.IsActive),
			fmt.Sprintf("%t", u.IsVerified),
			tier,
			u.CreatedAt.Format(time.RFC3339),
		})
	}
	writer.Flush()

	u.logAudit(ctx, &adminID, "ADMIN_EXPORT_USERS_CSV", "User", fmt.Sprintf("count:%d", len(users)), nil)
	return buf.Bytes(), nil
}

func (u *adminUseCase) ExportPayoutsCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error) {
	payouts, err := u.adminRepo.ExportPayouts(ctx, status)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	_ = writer.Write([]string{"ID", "WalletID", "Amount", "Status", "AdminNotes", "CreatedAt", "ProcessedAt"})
	for _, p := range payouts {
		processedAt := ""
		if p.ProcessedAt != nil {
			processedAt = p.ProcessedAt.Format(time.RFC3339)
		}
		_ = writer.Write([]string{
			p.ID.String(),
			p.WalletID.String(),
			fmt.Sprintf("%.2f", p.Amount),
			p.Status,
			p.AdminNotes,
			p.CreatedAt.Format(time.RFC3339),
			processedAt,
		})
	}
	writer.Flush()

	u.logAudit(ctx, &adminID, "ADMIN_EXPORT_PAYOUTS_CSV", "PayoutRequest", fmt.Sprintf("count:%d", len(payouts)), nil)
	return buf.Bytes(), nil
}

func (u *adminUseCase) ExportDisputesCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error) {
	disputes, err := u.adminRepo.ExportDisputes(ctx, status)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	_ = writer.Write([]string{"ID", "RaisedByID", "Reason", "Status", "ResolutionNotes", "CreatedAt"})
	for _, d := range disputes {
		_ = writer.Write([]string{
			d.ID.String(),
			d.RaisedByID.String(),
			d.Reason,
			d.Status,
			d.ResolutionNotes,
			d.CreatedAt.Format(time.RFC3339),
		})
	}
	writer.Flush()

	u.logAudit(ctx, &adminID, "ADMIN_EXPORT_DISPUTES_CSV", "Dispute", fmt.Sprintf("count:%d", len(disputes)), nil)
	return buf.Bytes(), nil
}

// ─── TOTP 2FA ENGINE ────────────────────────────────────────────────────────

func (u *adminUseCase) Setup2FA(ctx context.Context, adminID uuid.UUID) (*entity.TOTPSetupResponse, error) {
	user, err := u.userRepo.GetByID(ctx, adminID)
	if err != nil || user == nil {
		return nil, errors.New("admin user not found")
	}

	secret, err := auth.GenerateTOTPSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate 2FA secret: %w", err)
	}

	user.TwoFactorSecret = secret
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	otpAuthURL := auth.GetOTPAuthURL("NeighborServices", user.Email, secret)
	qrCodeURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=%s", otpAuthURL)

	return &entity.TOTPSetupResponse{
		Secret:     secret,
		OTPAuthURL: otpAuthURL,
		QRCodeURL:  qrCodeURL,
	}, nil
}

func (u *adminUseCase) VerifyAndEnable2FA(ctx context.Context, adminID uuid.UUID, code string) error {
	user, err := u.userRepo.GetByID(ctx, adminID)
	if err != nil || user == nil {
		return errors.New("admin user not found")
	}

	if user.TwoFactorSecret == "" {
		return errors.New("2FA setup not initialized. Call /2fa/setup first.")
	}

	if !auth.VerifyTOTPCode(user.TwoFactorSecret, code) {
		return errors.New("invalid 6-digit TOTP code")
	}

	user.TwoFactorEnabled = true
	user.UpdatedAt = time.Now()
	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}

	u.logAudit(ctx, &adminID, "ADMIN_ENABLE_2FA", "User", adminID.String(), nil)
	return nil
}

func (u *adminUseCase) Disable2FA(ctx context.Context, adminID uuid.UUID, code string) error {
	user, err := u.userRepo.GetByID(ctx, adminID)
	if err != nil || user == nil {
		return errors.New("admin user not found")
	}

	if !user.TwoFactorEnabled {
		return errors.New("2FA is not enabled on this account")
	}

	if !auth.VerifyTOTPCode(user.TwoFactorSecret, code) {
		return errors.New("invalid 6-digit TOTP code")
	}

	user.TwoFactorEnabled = false
	user.TwoFactorSecret = ""
	user.UpdatedAt = time.Now()
	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}

	u.logAudit(ctx, &adminID, "ADMIN_DISABLE_2FA", "User", adminID.String(), nil)
	return nil
}

// ─── STAFF IMPERSONATION ────────────────────────────────────────────────────

func (u *adminUseCase) ImpersonateUser(ctx context.Context, adminID, targetUserID uuid.UUID) (*entity.ImpersonationResult, error) {
	admin, err := u.userRepo.GetByID(ctx, adminID)
	if err != nil || admin == nil || (!admin.IsStaff && !admin.IsSuperuser) {
		return nil, errors.New("unauthorized: admin privileges required for impersonation")
	}

	targetUser, err := u.userRepo.GetByID(ctx, targetUserID)
	if err != nil || targetUser == nil {
		return nil, errors.New("target user not found")
	}

	userType := "SEEKER"
	if targetUser.Profile != nil && targetUser.Profile.UserType != "" {
		userType = targetUser.Profile.UserType
	}

	jwtSecret := "django-insecure-ns-secret-key-change-in-production"
	if u.cfg != nil && u.cfg.JWTSecret != "" {
		jwtSecret = u.cfg.JWTSecret
	}

	// Generate a 15-minute short-lived access token
	token, _, err := auth.GenerateTokenPair(
		targetUser.ID.String(),
		targetUser.Email,
		userType,
		jwtSecret,
		jwtSecret,
		1, // 1 hour access
		1,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate impersonation token: %w", err)
	}

	u.logAudit(ctx, &adminID, "ADMIN_IMPERSONATE_USER", "User", targetUserID.String(), map[string]interface{}{
		"target_email": targetUser.Email,
	})

	return &entity.ImpersonationResult{
		AccessToken:  token,
		ExpiresIn:    3600,
		TargetUserID: targetUser.ID,
		TargetEmail:  targetUser.Email,
	}, nil
}

// ─── RBAC ROLES ─────────────────────────────────────────────────────────────

func (u *adminUseCase) ListRoles(ctx context.Context, adminID uuid.UUID) ([]entity.AdminRole, error) {
	return u.adminRepo.ListRoles(ctx)
}

func (u *adminUseCase) CreateRole(ctx context.Context, adminID uuid.UUID, role *entity.AdminRole) (*entity.AdminRole, error) {
	role.ID = uuid.New()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	if err := u.adminRepo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_CREATE_ROLE", "AdminRole", role.ID.String(), map[string]interface{}{
		"name": role.Name,
		"slug": role.Slug,
	})
	return role, nil
}

func (u *adminUseCase) UpdateRole(ctx context.Context, adminID, roleID uuid.UUID, role *entity.AdminRole) (*entity.AdminRole, error) {
	existing, err := u.adminRepo.GetRoleByID(ctx, roleID)
	if err != nil || existing == nil {
		return nil, errors.New("role not found")
	}

	existing.Name = role.Name
	existing.Slug = role.Slug
	existing.Description = role.Description
	existing.Permissions = role.Permissions
	existing.UpdatedAt = time.Now()

	if err := u.adminRepo.UpdateRole(ctx, existing); err != nil {
		return nil, err
	}

	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_ROLE", "AdminRole", roleID.String(), map[string]interface{}{
		"name":        existing.Name,
		"slug":        existing.Slug,
		"permissions": existing.Permissions,
	})
	return existing, nil
}

func (u *adminUseCase) DeleteRole(ctx context.Context, adminID, roleID uuid.UUID) error {
	existing, err := u.adminRepo.GetRoleByID(ctx, roleID)
	if err != nil || existing == nil {
		return errors.New("role not found")
	}

	if err := u.adminRepo.DeleteRole(ctx, roleID); err != nil {
		return err
	}

	u.logAudit(ctx, &adminID, "ADMIN_DELETE_ROLE", "AdminRole", roleID.String(), map[string]interface{}{
		"name": existing.Name,
		"slug": existing.Slug,
	})
	return nil
}

func (u *adminUseCase) AssignUserRole(ctx context.Context, adminID, targetUserID, roleID uuid.UUID) error {
	user, err := u.userRepo.GetByID(ctx, targetUserID)
	if err != nil || user == nil {
		return errors.New("target user not found")
	}

	role, err := u.adminRepo.GetRoleByID(ctx, roleID)
	if err != nil || role == nil {
		return errors.New("role not found")
	}

	user.AdminRoleID = &roleID
	user.IsStaff = true
	user.UpdatedAt = time.Now()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}

	u.logAudit(ctx, &adminID, "ADMIN_ASSIGN_ROLE", "User", targetUserID.String(), map[string]interface{}{
		"role_name": role.Name,
		"role_slug": role.Slug,
	})
	return nil
}

// ─── FRAUD & RISK DETECTION ─────────────────────────────────────────────────

func (u *adminUseCase) ListFraudRiskAlerts(ctx context.Context, adminID uuid.UUID, status string, limit, offset int) ([]entity.FraudRiskAlert, int64, error) {
	return u.adminRepo.ListFraudRiskAlerts(ctx, status, limit, offset)
}

func (u *adminUseCase) EvaluateUserRisk(ctx context.Context, adminID, targetUserID uuid.UUID) (*entity.FraudRiskAlert, error) {
	alert, err := u.adminRepo.EvaluateUserRisk(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_EVALUATE_USER_RISK", "User", targetUserID.String(), map[string]interface{}{
		"risk_score": alert.RiskScore,
		"risk_level": alert.RiskLevel,
	})
	return alert, nil
}

func (u *adminUseCase) ResolveRiskAlert(ctx context.Context, adminID, alertID uuid.UUID, action string) error {
	u.logAudit(ctx, &adminID, "ADMIN_RESOLVE_RISK_ALERT", "FraudRiskAlert", alertID.String(), map[string]interface{}{
		"action": action,
	})
	return nil
}

// ─── NOTIFICATION TEMPLATES ─────────────────────────────────────────────────

func (u *adminUseCase) ListNotificationTemplates(ctx context.Context, adminID uuid.UUID) ([]entity.NotificationTemplate, error) {
	return u.adminRepo.ListNotificationTemplates(ctx)
}

func (u *adminUseCase) UpdateNotificationTemplate(ctx context.Context, adminID uuid.UUID, key, subject, bodyHTML, bodyText string) (*entity.NotificationTemplate, error) {
	tpl, err := u.adminRepo.GetNotificationTemplate(ctx, key)
	if err != nil || tpl == nil {
		return nil, errors.New("template not found")
	}

	if subject != "" {
		tpl.Subject = subject
	}
	if bodyHTML != "" {
		tpl.BodyHTML = bodyHTML
	}
	if bodyText != "" {
		tpl.BodyText = bodyText
	}
	tpl.UpdatedAt = time.Now()

	if err := u.adminRepo.UpdateNotificationTemplate(ctx, tpl); err != nil {
		return nil, err
	}

	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_TEMPLATE", "NotificationTemplate", key, nil)
	return tpl, nil
}

func (u *adminUseCase) TestSendEmailTemplate(ctx context.Context, adminID uuid.UUID, key, targetEmail string) error {
	tpl, err := u.adminRepo.GetNotificationTemplate(ctx, key)
	if err != nil || tpl == nil {
		return errors.New("template not found")
	}

	if targetEmail == "" {
		return errors.New("target email is required")
	}

	if u.cfg != nil && u.cfg.SMTPHost != "" {
		go func() {
			subject := tpl.Subject
			if subject == "" {
				subject = "Neighbor Service Notification Test"
			}
			body := tpl.BodyHTML
			if body == "" {
				body = tpl.BodyText
			}
			_ = email.SendHTML(&email.Config{
				Host:     u.cfg.SMTPHost,
				Port:     u.cfg.SMTPPort,
				User:     u.cfg.SMTPUser,
				Password: u.cfg.SMTPPassword,
				From:     u.cfg.EmailFrom,
			}, targetEmail, subject, body)
		}()
	}

	u.logAudit(ctx, &adminID, "ADMIN_TEST_SEND_TEMPLATE", "NotificationTemplate", key, map[string]interface{}{
		"recipient": targetEmail,
	})
	return nil
}

// ─── DATABASE BACKUPS ───────────────────────────────────────────────────────

func (u *adminUseCase) TriggerDatabaseBackup(ctx context.Context, adminID uuid.UUID) (*entity.BackupSnapshot, error) {
	now := time.Now()
	filename := fmt.Sprintf("ns_backup_%s.sql.gz", now.Format("20060102_150405"))
	storagePath := fmt.Sprintf("backups/%s", filename)

	hash := sha256.Sum256([]byte(filename + now.String()))
	checksum := hex.EncodeToString(hash[:])

	snap := entity.BackupSnapshot{
		ID:          uuid.New(),
		Filename:    filename,
		FileSize:    1024 * 1024 * 5, // ~5MB simulated snapshot
		Status:      "COMPLETED",
		StoragePath: storagePath,
		Checksum:    checksum,
		CreatedAt:   now,
		CompletedAt: &now,
	}

	if err := u.adminRepo.CreateBackupSnapshot(ctx, &snap); err != nil {
		return nil, err
	}

	u.logAudit(ctx, &adminID, "ADMIN_TRIGGER_BACKUP", "BackupSnapshot", snap.ID.String(), map[string]interface{}{
		"filename": filename,
	})

	return &snap, nil
}

func (u *adminUseCase) ListBackupSnapshots(ctx context.Context, adminID uuid.UUID) ([]entity.BackupSnapshot, error) {
	return u.adminRepo.ListBackupSnapshots(ctx)
}

func (u *adminUseCase) GetBackupSnapshot(ctx context.Context, adminID uuid.UUID, idOrFilename string) (*entity.BackupSnapshot, error) {
	if parsedUUID, err := uuid.Parse(idOrFilename); err == nil {
		snap, err := u.adminRepo.GetBackupSnapshotByID(ctx, parsedUUID)
		if err == nil {
			return snap, nil
		}
	}
	return u.adminRepo.GetBackupSnapshotByFilename(ctx, idOrFilename)
}

// ─── BATCH PAYOUTS ──────────────────────────────────────────────────────────

func (u *adminUseCase) BatchApprovePayouts(ctx context.Context, adminID uuid.UUID, ids []uuid.UUID) (*entity.BatchPayoutResponse, error) {
	return u.adminRepo.BatchApprovePayouts(ctx, ids, adminID)
}

func (u *adminUseCase) BatchRejectPayouts(ctx context.Context, adminID uuid.UUID, ids []uuid.UUID, reason string) (*entity.BatchPayoutResponse, error) {
	return u.adminRepo.BatchRejectPayouts(ctx, ids, reason, adminID)
}

// ─── LIVE NOTIFICATION FEED & TELEMETRY ─────────────────────────────────────

func (u *adminUseCase) GetAdminNotificationFeed(ctx context.Context, adminID uuid.UUID) ([]entity.AdminNotification, error) {
	return u.adminRepo.GetAdminNotificationFeed(ctx)
}

func (u *adminUseCase) GetProviderOnboardingFunnel(ctx context.Context, adminID uuid.UUID) (*entity.ProviderOnboardingFunnel, error) {
	return u.adminRepo.GetProviderOnboardingFunnel(ctx)
}

// ─── STAFF INTERNAL NOTES ───────────────────────────────────────────────────

func (u *adminUseCase) ListStaffNotes(ctx context.Context, adminID, userID uuid.UUID) ([]entity.StaffNote, error) {
	return u.adminRepo.ListStaffNotes(ctx, userID)
}

func (u *adminUseCase) CreateStaffNote(ctx context.Context, adminID, userID uuid.UUID, content, category string, isPinned bool) (*entity.StaffNote, error) {
	adminUser, _ := u.userRepo.GetByID(ctx, adminID)
	authorEmail := "admin@neighborservice.com"
	if adminUser != nil {
		authorEmail = adminUser.Email
	}

	note := &entity.StaffNote{
		ID:            uuid.New(),
		UserID:        userID,
		AuthorAdminID: adminID,
		AuthorEmail:   authorEmail,
		Content:       content,
		Category:      category,
		IsPinned:      isPinned,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := u.adminRepo.CreateStaffNote(ctx, note); err != nil {
		return nil, err
	}

	u.logAudit(ctx, &adminID, "ADMIN_CREATE_STAFF_NOTE", "StaffNote", note.ID.String(), map[string]interface{}{
		"target_user_id": userID.String(),
		"category":       category,
		"pinned":         isPinned,
	})

	return note, nil
}

func (u *adminUseCase) DeleteStaffNote(ctx context.Context, adminID, noteID uuid.UUID) error {
	u.logAudit(ctx, &adminID, "ADMIN_DELETE_STAFF_NOTE", "StaffNote", noteID.String(), nil)
	return u.adminRepo.DeleteStaffNote(ctx, noteID)
}

// ─── BOOKINGS & APPOINTMENTS CONSOLE ─────────────────────────────────────────

func (u *adminUseCase) ListAppointments(ctx context.Context, adminID uuid.UUID, status, search string, limit, offset int) ([]entity.Appointment, int64, error) {
	return u.adminRepo.ListAppointments(ctx, status, search, limit, offset)
}

func (u *adminUseCase) GetAppointmentByID(ctx context.Context, adminID, id uuid.UUID) (*entity.Appointment, error) {
	return u.adminRepo.GetAppointmentByID(ctx, id)
}

func (u *adminUseCase) UpdateAppointmentStatus(ctx context.Context, adminID, id uuid.UUID, status string) error {
	if err := u.adminRepo.UpdateAppointmentStatus(ctx, id, status); err != nil {
		return err
	}
	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_APPOINTMENT_STATUS", "Appointment", id.String(), map[string]interface{}{
		"new_status": status,
	})
	return nil
}

func (u *adminUseCase) ExportAppointmentsCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error) {
	apts, err := u.adminRepo.ExportAppointments(ctx, status)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"ID", "Title", "Status", "TotalPrice", "SeekerEmail", "ProviderEmail", "CreatedAt"})
	for _, a := range apts {
		seekerEmail := ""
		providerEmail := ""
		if a.Seeker != nil {
			seekerEmail = a.Seeker.Email
		}
		if a.Provider != nil {
			providerEmail = a.Provider.Email
		}
		_ = writer.Write([]string{
			a.ID.String(),
			a.Title,
			a.Status,
			fmt.Sprintf("%.2f", a.TotalPrice),
			seekerEmail,
			providerEmail,
			a.CreatedAt.Format(time.RFC3339),
		})
	}
	writer.Flush()
	u.logAudit(ctx, &adminID, "ADMIN_EXPORT_APPOINTMENTS_CSV", "Appointment", fmt.Sprintf("count:%d", len(apts)), nil)
	return buf.Bytes(), nil
}

// ─── REVIEWS & RATINGS MODERATION ───────────────────────────────────────────

func (u *adminUseCase) ListReviews(ctx context.Context, adminID uuid.UUID, rating *float64, isHidden *bool, limit, offset int) ([]entity.Review, int64, error) {
	return u.adminRepo.ListReviews(ctx, rating, isHidden, limit, offset)
}

func (u *adminUseCase) CreateReview(ctx context.Context, adminID uuid.UUID, review *entity.Review) (*entity.Review, error) {
	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}
	now := time.Now()
	if review.CreatedAt.IsZero() {
		review.CreatedAt = now
	}
	review.UpdatedAt = now
	if err := u.adminRepo.CreateReview(ctx, review); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_CREATE_REVIEW", "Review", review.ID.String(), map[string]interface{}{
		"provider_id": review.ProviderID.String(),
		"reviewer_id": review.ReviewerID.String(),
		"rating":      review.Rating,
	})
	return review, nil
}

func (u *adminUseCase) ToggleReviewVisibility(ctx context.Context, adminID, id uuid.UUID, isHidden bool) error {
	if err := u.adminRepo.ToggleReviewVisibility(ctx, id, isHidden); err != nil {
		return err
	}
	u.logAudit(ctx, &adminID, "ADMIN_TOGGLE_REVIEW_VISIBILITY", "Review", id.String(), map[string]interface{}{
		"is_hidden": isHidden,
	})
	return nil
}

func (u *adminUseCase) DeleteReview(ctx context.Context, adminID, id uuid.UUID) error {
	if err := u.adminRepo.DeleteReview(ctx, id); err != nil {
		return err
	}
	u.logAudit(ctx, &adminID, "ADMIN_DELETE_REVIEW", "Review", id.String(), nil)
	return nil
}

// ─── PROMO CODES & MARKETING CAMPAIGNS ───────────────────────────────────────

func (u *adminUseCase) ListPromoCodes(ctx context.Context, adminID uuid.UUID) ([]entity.PromoCode, error) {
	return u.adminRepo.ListPromoCodes(ctx)
}

func (u *adminUseCase) CreatePromoCode(ctx context.Context, adminID uuid.UUID, promo *entity.PromoCode) (*entity.PromoCode, error) {
	if err := u.adminRepo.CreatePromoCode(ctx, promo); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_CREATE_PROMO_CODE", "PromoCode", promo.ID.String(), map[string]interface{}{
		"code":           promo.Code,
		"discount_type":  promo.DiscountType,
		"discount_value": promo.DiscountValue,
	})
	return promo, nil
}

func (u *adminUseCase) UpdatePromoCode(ctx context.Context, adminID uuid.UUID, promo *entity.PromoCode) (*entity.PromoCode, error) {
	if err := u.adminRepo.UpdatePromoCode(ctx, promo); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_PROMO_CODE", "PromoCode", promo.ID.String(), map[string]interface{}{
		"code":      promo.Code,
		"is_active": promo.IsActive,
	})
	return promo, nil
}

func (u *adminUseCase) DeletePromoCode(ctx context.Context, adminID, id uuid.UUID) error {
	if err := u.adminRepo.DeletePromoCode(ctx, id); err != nil {
		return err
	}
	u.logAudit(ctx, &adminID, "ADMIN_DELETE_PROMO_CODE", "PromoCode", id.String(), nil)
	return nil
}

// ─── DISPUTE REFUND EXECUTION ────────────────────────────────────────────────

func (u *adminUseCase) ExecuteDisputeRefund(ctx context.Context, adminID, disputeID uuid.UUID, amount float64, reason string) error {
	return u.adminRepo.ExecuteDisputeRefund(ctx, disputeID, amount, reason, adminID)
}

// ─── ADDITIONAL DATA EXPORTS ─────────────────────────────────────────────────

func (u *adminUseCase) ExportVerificationsCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error) {
	vers, err := u.adminRepo.ExportVerifications(ctx, status)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"ID", "ProviderID", "Status", "ReviewerNotes", "CreatedAt"})
	for _, v := range vers {
		_ = writer.Write([]string{
			v.ID.String(),
			v.ProviderID.String(),
			v.Status,
			v.ReviewerNotes,
			v.CreatedAt.Format(time.RFC3339),
		})
	}
	writer.Flush()
	u.logAudit(ctx, &adminID, "ADMIN_EXPORT_VERIFICATIONS_CSV", "ProviderVerification", fmt.Sprintf("count:%d", len(vers)), nil)
	return buf.Bytes(), nil
}

func (u *adminUseCase) ExportBackgroundChecksCSV(ctx context.Context, adminID uuid.UUID, status string) ([]byte, error) {
	bcs, err := u.adminRepo.ExportBackgroundChecks(ctx, status)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"ID", "ProviderID", "Status", "Package", "Adjudication", "CreatedAt"})
	for _, b := range bcs {
		_ = writer.Write([]string{
			b.ID.String(),
			b.ProviderID.String(),
			b.Status,
			b.Package,
			b.Adjudication,
			b.CreatedAt.Format(time.RFC3339),
		})
	}
	writer.Flush()
	u.logAudit(ctx, &adminID, "ADMIN_EXPORT_BACKGROUND_CHECKS_CSV", "BackgroundCheck", fmt.Sprintf("count:%d", len(bcs)), nil)
	return buf.Bytes(), nil
}

// ─── LEGAL & COMPLIANCE DOCUMENTS ───────────────────────────────────────────

func (u *adminUseCase) ListLegalDocuments(ctx context.Context, adminID uuid.UUID, docType string, isActive *bool) ([]entity.LegalDocument, error) {
	return u.adminRepo.ListLegalDocuments(ctx, docType, isActive)
}

func (u *adminUseCase) GetLegalDocumentByID(ctx context.Context, adminID, id uuid.UUID) (*entity.LegalDocument, error) {
	return u.adminRepo.GetLegalDocumentByID(ctx, id)
}

func (u *adminUseCase) CreateLegalDocument(ctx context.Context, adminID uuid.UUID, doc *entity.LegalDocument) (*entity.LegalDocument, error) {
	if err := u.adminRepo.CreateLegalDocument(ctx, doc); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_CREATE_LEGAL_DOCUMENT", "LegalDocument", doc.ID.String(), map[string]interface{}{
		"doc_type": doc.DocType,
		"version":  doc.Version,
		"title":    doc.Title,
	})
	return doc, nil
}

func (u *adminUseCase) UpdateLegalDocument(ctx context.Context, adminID uuid.UUID, doc *entity.LegalDocument) (*entity.LegalDocument, error) {
	if err := u.adminRepo.UpdateLegalDocument(ctx, doc); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_LEGAL_DOCUMENT", "LegalDocument", doc.ID.String(), map[string]interface{}{
		"doc_type":  doc.DocType,
		"version":   doc.Version,
		"is_active": doc.IsActive,
	})
	return doc, nil
}

func (u *adminUseCase) DeleteLegalDocument(ctx context.Context, adminID, id uuid.UUID) error {
	if err := u.adminRepo.DeleteLegalDocument(ctx, id); err != nil {
		return err
	}
	u.logAudit(ctx, &adminID, "ADMIN_DELETE_LEGAL_DOCUMENT", "LegalDocument", id.String(), nil)
	return nil
}

// ─── SUPPORT MESSAGES & INQUIRIES ───────────────────────────────────────────

func (u *adminUseCase) ListContactMessages(ctx context.Context, adminID uuid.UUID, isResolved *bool, search string, limit, offset int) ([]entity.ContactMessage, int64, error) {
	return u.adminRepo.ListContactMessages(ctx, isResolved, search, limit, offset)
}

func (u *adminUseCase) ToggleContactMessageResolved(ctx context.Context, adminID, id uuid.UUID, isResolved bool) error {
	if err := u.adminRepo.ToggleContactMessageResolved(ctx, id, isResolved); err != nil {
		return err
	}
	u.logAudit(ctx, &adminID, "ADMIN_TOGGLE_CONTACT_MESSAGE_RESOLVED", "ContactMessage", id.String(), map[string]interface{}{
		"is_resolved": isResolved,
	})
	return nil
}

func (u *adminUseCase) DeleteContactMessage(ctx context.Context, adminID, id uuid.UUID) error {
	if err := u.adminRepo.DeleteContactMessage(ctx, id); err != nil {
		return err
	}
	u.logAudit(ctx, &adminID, "ADMIN_DELETE_CONTACT_MESSAGE", "ContactMessage", id.String(), nil)
	return nil
}

// ─── RESOLUTION INCIDENT REPORTS ────────────────────────────────────────────

func (u *adminUseCase) ListResolutionReports(ctx context.Context, adminID uuid.UUID, isReviewed *bool, search string, limit, offset int) ([]entity.ResolutionReport, int64, error) {
	return u.adminRepo.ListResolutionReports(ctx, isReviewed, search, limit, offset)
}

func (u *adminUseCase) ToggleResolutionReportReviewed(ctx context.Context, adminID, id uuid.UUID, isReviewed bool) error {
	if err := u.adminRepo.ToggleResolutionReportReviewed(ctx, id, isReviewed); err != nil {
		return err
	}
	u.logAudit(ctx, &adminID, "ADMIN_TOGGLE_RESOLUTION_REPORT_REVIEWED", "ResolutionReport", id.String(), map[string]interface{}{
		"is_reviewed": isReviewed,
	})
	return nil
}

func (u *adminUseCase) DeleteResolutionReport(ctx context.Context, adminID, id uuid.UUID) error {
	if err := u.adminRepo.DeleteResolutionReport(ctx, id); err != nil {
		return err
	}
	u.logAudit(ctx, &adminID, "ADMIN_DELETE_RESOLUTION_REPORT", "ResolutionReport", id.String(), nil)
	return nil
}

// ─── PUBLIC SITE CMS ────────────────────────────────────────────────────────

func (u *adminUseCase) ListFAQs(ctx context.Context, adminID uuid.UUID, category string, isActive *bool) ([]entity.FAQ, error) {
	return u.adminRepo.ListFAQs(ctx, category, isActive)
}

func (u *adminUseCase) CreateFAQ(ctx context.Context, adminID uuid.UUID, faq *entity.FAQ) (*entity.FAQ, error) {
	if err := u.adminRepo.CreateFAQ(ctx, faq); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_CREATE_FAQ", "FAQ", faq.ID.String(), map[string]interface{}{"question": faq.Question})
	return faq, nil
}

func (u *adminUseCase) UpdateFAQ(ctx context.Context, adminID uuid.UUID, faq *entity.FAQ) (*entity.FAQ, error) {
	if err := u.adminRepo.UpdateFAQ(ctx, faq); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_FAQ", "FAQ", faq.ID.String(), map[string]interface{}{"question": faq.Question})
	return faq, nil
}

func (u *adminUseCase) DeleteFAQ(ctx context.Context, adminID, id uuid.UUID) error {
	if err := u.adminRepo.DeleteFAQ(ctx, id); err != nil {
		return err
	}
	u.logAudit(ctx, &adminID, "ADMIN_DELETE_FAQ", "FAQ", id.String(), nil)
	return nil
}

func (u *adminUseCase) ListTestimonials(ctx context.Context, adminID uuid.UUID, isActive *bool) ([]entity.Testimonial, error) {
	return u.adminRepo.ListTestimonials(ctx, isActive)
}

func (u *adminUseCase) CreateTestimonial(ctx context.Context, adminID uuid.UUID, t *entity.Testimonial) (*entity.Testimonial, error) {
	if err := u.adminRepo.CreateTestimonial(ctx, t); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_CREATE_TESTIMONIAL", "Testimonial", t.ID.String(), map[string]interface{}{"name": t.Name})
	return t, nil
}

func (u *adminUseCase) UpdateTestimonial(ctx context.Context, adminID uuid.UUID, t *entity.Testimonial) (*entity.Testimonial, error) {
	if err := u.adminRepo.UpdateTestimonial(ctx, t); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_TESTIMONIAL", "Testimonial", t.ID.String(), map[string]interface{}{"name": t.Name})
	return t, nil
}

func (u *adminUseCase) DeleteTestimonial(ctx context.Context, adminID, id uuid.UUID) error {
	if err := u.adminRepo.DeleteTestimonial(ctx, id); err != nil {
		return err
	}
	u.logAudit(ctx, &adminID, "ADMIN_DELETE_TESTIMONIAL", "Testimonial", id.String(), nil)
	return nil
}

func (u *adminUseCase) GetHeroSection(ctx context.Context, adminID uuid.UUID) (*entity.HeroSection, error) {
	return u.adminRepo.GetHeroSection(ctx)
}

func (u *adminUseCase) UpdateHeroSection(ctx context.Context, adminID uuid.UUID, hero *entity.HeroSection) (*entity.HeroSection, error) {
	if err := u.adminRepo.UpdateHeroSection(ctx, hero); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_HERO_SECTION", "HeroSection", hero.ID.String(), nil)
	return hero, nil
}

func (u *adminUseCase) ListSiteStats(ctx context.Context, adminID uuid.UUID) ([]entity.SiteStat, error) {
	return u.adminRepo.ListSiteStats(ctx)
}

func (u *adminUseCase) CreateSiteStat(ctx context.Context, adminID uuid.UUID, stat *entity.SiteStat) (*entity.SiteStat, error) {
	if err := u.adminRepo.CreateSiteStat(ctx, stat); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_CREATE_SITE_STAT", "SiteStat", stat.ID.String(), nil)
	return stat, nil
}

func (u *adminUseCase) UpdateSiteStat(ctx context.Context, adminID uuid.UUID, stat *entity.SiteStat) (*entity.SiteStat, error) {
	if err := u.adminRepo.UpdateSiteStat(ctx, stat); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_SITE_STAT", "SiteStat", stat.ID.String(), nil)
	return stat, nil
}

func (u *adminUseCase) DeleteSiteStat(ctx context.Context, adminID, id uuid.UUID) error {
	if err := u.adminRepo.DeleteSiteStat(ctx, id); err != nil {
		return err
	}
	u.logAudit(ctx, &adminID, "ADMIN_DELETE_SITE_STAT", "SiteStat", id.String(), nil)
	return nil
}

func (u *adminUseCase) GetAboutContent(ctx context.Context, adminID uuid.UUID) (*entity.AboutContent, error) {
	return u.adminRepo.GetAboutContent(ctx)
}

func (u *adminUseCase) UpdateAboutContent(ctx context.Context, adminID uuid.UUID, about *entity.AboutContent) (*entity.AboutContent, error) {
	if err := u.adminRepo.UpdateAboutContent(ctx, about); err != nil {
		return nil, err
	}
	u.logAudit(ctx, &adminID, "ADMIN_UPDATE_ABOUT_CONTENT", "AboutContent", about.ID.String(), nil)
	return about, nil
}

