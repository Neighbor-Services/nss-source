package usecase

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"

	"github.com/google/uuid"
)

// Consultation UseCase
type consultationUseCase struct {
	cfg *config.Config
}

func NewConsultationUseCase(cfg *config.Config) domainUsecase.ConsultationUseCase {
	return &consultationUseCase{cfg: cfg}
}

func (u *consultationUseCase) GetRTCToken(ctx context.Context, channelName string, uid uint32, role string) (*domainUsecase.AgoraTokenResponse, error) {
	appID := ""
	appCert := ""
	if u.cfg != nil {
		appID = u.cfg.AgoraAppID
		appCert = u.cfg.AgoraAppCertificate
	}
	expiredTs := time.Now().Add(1 * time.Hour).Unix()
	var token string
	if appID != "" && appCert != "" {
		message := fmt.Sprintf("%s%s%d%d", appID, channelName, uid, expiredTs)
		mac := hmac.New(sha256.New, []byte(appCert))
		mac.Write([]byte(message))
		signature := hex.EncodeToString(mac.Sum(nil))
		token = fmt.Sprintf("006%s%s%d%d%s", appID, signature, uid, expiredTs, channelName)
	} else {
		token = fmt.Sprintf("AGORA_RTC_%s_%d_%d", channelName, uid, expiredTs)
	}

	return &domainUsecase.AgoraTokenResponse{
		Token:       token,
		ChannelName: channelName,
		Channel:     channelName,
		UID:         uid,
		AppID:       appID,
	}, nil
}

func (u *consultationUseCase) GetRTMToken(ctx context.Context, userAccount string) (*domainUsecase.AgoraTokenResponse, error) {
	appID := ""
	appCert := ""
	if u.cfg != nil {
		appID = u.cfg.AgoraAppID
		appCert = u.cfg.AgoraAppCertificate
	}
	expiredTs := time.Now().Add(1 * time.Hour).Unix()
	var token string
	if appID != "" && appCert != "" {
		message := fmt.Sprintf("%s%s%d", appID, userAccount, expiredTs)
		mac := hmac.New(sha256.New, []byte(appCert))
		mac.Write([]byte(message))
		signature := hex.EncodeToString(mac.Sum(nil))
		token = fmt.Sprintf("006%s%s%d%s", appID, signature, expiredTs, userAccount)
	} else {
		token = fmt.Sprintf("AGORA_RTM_%s_%d", userAccount, expiredTs)
	}

	return &domainUsecase.AgoraTokenResponse{
		Token:   token,
		UserID:  userAccount,
		Channel: userAccount,
	}, nil
}

// Moderation UseCase
type moderationUseCase struct {
	reportRepo       repository.ReportRepository
	verificationRepo repository.ProviderVerificationRepository
	bgCheckRepo      repository.BackgroundCheckRepository
	profileRepo      repository.ProfileRepository
	userRepo         repository.UserRepository
	notifUC          domainUsecase.NotificationUseCase
	cfg              *config.Config
}

func NewModerationUseCase(
	reportRepo repository.ReportRepository,
	verificationRepo repository.ProviderVerificationRepository,
	bgCheckRepo repository.BackgroundCheckRepository,
	profileRepo repository.ProfileRepository,
	userRepo repository.UserRepository,
	notifUC domainUsecase.NotificationUseCase,
	cfg *config.Config,
) domainUsecase.ModerationUseCase {
	return &moderationUseCase{
		reportRepo:       reportRepo,
		verificationRepo: verificationRepo,
		bgCheckRepo:      bgCheckRepo,
		profileRepo:      profileRepo,
		userRepo:         userRepo,
		notifUC:          notifUC,
		cfg:              cfg,
	}
}

func (u *moderationUseCase) SubmitReport(ctx context.Context, reporterID uuid.UUID, reportedUserID *uuid.UUID, contentType, objectID, reason, description string, evidence []string) (*entity.Report, error) {
	resourceType := contentType
	if resourceType == "" {
		resourceType = "General"
	}

	var primaryEvidence string
	if len(evidence) > 0 {
		primaryEvidence = evidence[0]
	}

	report := entity.Report{
		ID:             uuid.New(),
		ReporterID:     reporterID,
		ReportedUserID: reportedUserID,
		ResourceType:   resourceType,
		ResourceID:     objectID,
		Reason:         reason,
		Evidence:       primaryEvidence,
		EvidenceList:   evidence,
		Status:         "PENDING",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := u.reportRepo.Create(ctx, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

func (u *moderationUseCase) GetReports(ctx context.Context, status string, limit, offset int) ([]entity.Report, error) {
	return u.reportRepo.List(ctx, status, limit, offset)
}

func (u *moderationUseCase) GetVerifications(ctx context.Context, providerID *uuid.UUID) ([]entity.ProviderVerification, error) {
	return u.verificationRepo.List(ctx, providerID)
}

func (u *moderationUseCase) SubmitVerification(ctx context.Context, providerID uuid.UUID, input domainUsecase.VerificationSubmitInput) (*entity.ProviderVerification, error) {
	existingList, _ := u.verificationRepo.List(ctx, &providerID)
	var v *entity.ProviderVerification
	if len(existingList) > 0 {
		v = &existingList[0]
	} else {
		v = &entity.ProviderVerification{
			ID:         uuid.New(),
			ProviderID: providerID,
			CreatedAt:  time.Now(),
		}
	}

	if input.DocumentType != "" {
		v.DocumentType = input.DocumentType
	}
	if input.DocumentFront != "" {
		v.DocumentFront = input.DocumentFront
	}
	if input.DocumentBack != "" {
		v.DocumentBack = input.DocumentBack
	}
	if input.Selfie != "" {
		v.Selfie = input.Selfie
	}
	if input.TradeLicense != "" {
		v.TradeLicense = input.TradeLicense
	}
	if input.LicenseNumber != "" {
		v.LicenseNumber = input.LicenseNumber
	}
	if input.LicenseExpiry != nil {
		v.LicenseExpiry = input.LicenseExpiry
	}
	v.Status = "PENDING"
	v.ReviewerNotes = ""
	v.UpdatedAt = time.Now()

	if len(existingList) > 0 {
		if err := u.verificationRepo.Update(ctx, v); err != nil {
			return nil, err
		}
	} else {
		if err := u.verificationRepo.Create(ctx, v); err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (u *moderationUseCase) GetBackgroundChecks(ctx context.Context, providerID *uuid.UUID) ([]entity.BackgroundCheck, error) {
	return u.bgCheckRepo.List(ctx, providerID)
}

func (u *moderationUseCase) GetBackgroundCheckConfig(ctx context.Context) (map[string]interface{}, error) {
	setting, err := u.bgCheckRepo.GetModerationSetting(ctx)
	mode := "NATIVE_CHECKR"
	if err == nil && setting != nil && setting.BackgroundCheckPaymentMode != "" {
		mode = "NATIVE_CHECKR"
	}
	env := "production"
	if u.cfg != nil && u.cfg.Env != "" {
		env = u.cfg.Env
	}
	skipVerification := env == "staging" || env == "stage" || env == "dev" || env == "development"
	return map[string]interface{}{
		"payment_mode":      mode,
		"env":               env,
		"skip_verification": skipVerification,
	}, nil
}

func (u *moderationUseCase) checkrRequest(method, endpoint string, payload interface{}) (map[string]interface{}, error) {
	apiKey := ""
	if u.cfg != nil {
		apiKey = u.cfg.CheckrAPIKey
	}
	if apiKey == "" {
		return nil, errors.New("Checkr API key not configured")
	}

	url := "https://api.checkr.com" + endpoint
	var bodyReader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(apiKey, "")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Printf("[Checkr API] %s %s -> Status %d: %s", method, endpoint, resp.StatusCode, string(bodyBytes))

	var res map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, fmt.Errorf("Checkr response parsing error: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errDetails []string
		if msg, ok := res["message"].(string); ok && msg != "" {
			errDetails = append(errDetails, msg)
		}
		if errDesc, ok := res["error_description"].(string); ok && errDesc != "" {
			errDetails = append(errDetails, errDesc)
		}
		if errMsg, ok := res["error"].(string); ok && errMsg != "" {
			errDetails = append(errDetails, errMsg)
		} else if errObj, ok := res["error"].(map[string]interface{}); ok {
			if m, ok := errObj["message"].(string); ok {
				errDetails = append(errDetails, m)
			}
		}
		if errList, ok := res["errors"].([]interface{}); ok {
			for _, item := range errList {
				if s, ok := item.(string); ok {
					errDetails = append(errDetails, s)
				} else if m, ok := item.(map[string]interface{}); ok {
					if f, ok := m["field"].(string); ok {
						msg := fmt.Sprintf("%v", m["message"])
						errDetails = append(errDetails, fmt.Sprintf("%s: %s", f, msg))
					} else if msg, ok := m["message"].(string); ok {
						errDetails = append(errDetails, msg)
					}
				}
			}
		}

		errMsg := strings.Join(errDetails, "; ")
		if errMsg == "" {
			errMsg = string(bodyBytes)
		}
		return nil, fmt.Errorf("Checkr API error (HTTP %d): %s", resp.StatusCode, errMsg)
	}

	return res, nil
}

func (u *moderationUseCase) getFirstAvailablePackage() string {
	res, err := u.checkrRequest("GET", "/v1/packages", nil)
	if err != nil {
		return ""
	}
	if data, ok := res["data"].([]interface{}); ok && len(data) > 0 {
		for _, item := range data {
			if pkg, ok := item.(map[string]interface{}); ok {
				if slug, ok := pkg["slug"].(string); ok && slug != "" {
					return slug
				}
			}
		}
	}
	return ""
}

func (u *moderationUseCase) GetMyBackgroundCheck(ctx context.Context, providerID uuid.UUID) (*entity.BackgroundCheck, error) {
	return u.bgCheckRepo.GetLatestByProvider(ctx, providerID)
}

func (u *moderationUseCase) InitiateBackgroundCheck(ctx context.Context, userID uuid.UUID) (*entity.BackgroundCheck, error) {
	// 1. Guard against duplicate pending checks within validity window (7 days)
	existing, _ := u.bgCheckRepo.GetLatestByProvider(ctx, userID)
	if existing != nil {
		if existing.Status == "clear" {
			return existing, nil
		}
		if existing.Status == "pending" && time.Since(existing.CreatedAt) < 7*24*time.Hour && existing.InvitationURL != "" {
			return existing, nil
		}
		if existing.Status == "pending" && time.Since(existing.CreatedAt) >= 7*24*time.Hour {
			existing.Status = "expired"
			_ = u.bgCheckRepo.Update(ctx, existing)
		}
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	profile, err := u.profileRepo.GetByUserID(ctx, userID)
	if err != nil || profile == nil {
		return nil, errors.New("profile not found")
	}

	// 2. Resolve Candidate Names (Checkr requires both first_name and last_name)
	firstName := strings.TrimSpace(profile.FirstName)
	lastName := strings.TrimSpace(profile.LastName)

	if firstName == "" && lastName == "" {
		parts := strings.Split(user.Email, "@")
		firstName = parts[0]
		lastName = "Provider"
	} else if lastName == "" {
		parts := strings.Fields(firstName)
		if len(parts) > 1 {
			firstName = parts[0]
			lastName = strings.Join(parts[1:], " ")
		} else {
			lastName = "Provider"
		}
	} else if firstName == "" {
		firstName = "Provider"
	}

	candidatePayload := map[string]interface{}{
		"first_name": firstName,
		"last_name":  lastName,
		"email":      user.Email,
		"custom_id":  user.ID.String(),
	}
	if profile.Phone != "" {
		candidatePayload["phone"] = profile.Phone
	}
	if profile.ZipCode != "" {
		candidatePayload["zipcode"] = profile.ZipCode
	}

	packageName := "tasker_standard"
	if u.cfg != nil && u.cfg.CheckrPackageName != "" {
		packageName = u.cfg.CheckrPackageName
	}

	state := "CA"
	if profile.State != "" {
		st := strings.TrimSpace(strings.ToUpper(profile.State))
		if len(st) == 2 {
			state = st
		}
	}

	var candidateID string
	var invitationID string
	var invitationURL string

	// 3. Live Checkr API Call
	if u.cfg != nil && u.cfg.CheckrAPIKey != "" {
		candidateRes, err := u.checkrRequest("POST", "/v1/candidates", candidatePayload)
		if err != nil {
			return nil, fmt.Errorf("failed to create Checkr candidate: %w", err)
		}
		if idStr, ok := candidateRes["id"].(string); ok && idStr != "" {
			candidateID = idStr
		} else {
			return nil, errors.New("Checkr did not return a valid candidate ID")
		}

		invitationPayload := map[string]interface{}{
			"candidate_id": candidateID,
			"package":      packageName,
			"work_locations": []map[string]string{
				{
					"country": "US",
					"state":   state,
				},
			},
		}
		invitationRes, err := u.checkrRequest("POST", "/v1/invitations", invitationPayload)
		if err != nil {
			// If invitation failed (e.g. package not found), check account for available package
			if availablePkg := u.getFirstAvailablePackage(); availablePkg != "" && availablePkg != packageName {
				log.Printf("[Checkr API] Retrying invitation with discovered package '%s' (fallback from '%s')", availablePkg, packageName)
				invitationPayload["package"] = availablePkg
				packageName = availablePkg
				invitationRes, err = u.checkrRequest("POST", "/v1/invitations", invitationPayload)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create Checkr invitation: %w", err)
		}
		if idStr, ok := invitationRes["id"].(string); ok && idStr != "" {
			invitationID = idStr
		}
		if urlStr, ok := invitationRes["invitation_url"].(string); ok && urlStr != "" {
			invitationURL = urlStr
		} else {
			return nil, errors.New("Checkr did not return an invitation URL")
		}
	} else {
		// Fallback for tests / development when CHECKR_API_KEY is not configured
		candidateID = "cand_" + uuid.New().String()[:12]
		invitationID = "inv_" + uuid.New().String()[:12]
		invitationURL = "https://invitations.checkr.com/" + invitationID
	}

	bc := entity.BackgroundCheck{
		ID:                 uuid.New(),
		ProviderID:         userID,
		CheckrCandidateID:  candidateID,
		CheckrInvitationID: invitationID,
		InvitationURL:      invitationURL,
		Package:            packageName,
		Status:             "pending",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := u.bgCheckRepo.Create(ctx, &bc); err != nil {
		return nil, err
	}

	return &bc, nil
}

func (u *moderationUseCase) ResyncBackgroundCheck(ctx context.Context, id uuid.UUID) (*entity.BackgroundCheck, error) {
	bc, err := u.bgCheckRepo.GetByID(ctx, id)
	if err != nil || bc == nil {
		return nil, errors.New("background check not found")
	}

	if bc.CheckrReportID != "" {
		reportRes, err := u.checkrRequest("GET", "/v1/reports/"+bc.CheckrReportID, nil)
		if err == nil && reportRes != nil {
			u.applyReportToBackgroundCheck(ctx, bc, reportRes)
		}
	} else if bc.CheckrCandidateID != "" {
		candRes, err := u.checkrRequest("GET", "/v1/candidates/"+bc.CheckrCandidateID, nil)
		if err == nil && candRes != nil {
			if reportIDs, ok := candRes["report_ids"].([]interface{}); ok && len(reportIDs) > 0 {
				if rID, ok := reportIDs[0].(string); ok && rID != "" {
					bc.CheckrReportID = rID
					reportRes, err := u.checkrRequest("GET", "/v1/reports/"+rID, nil)
					if err == nil && reportRes != nil {
						u.applyReportToBackgroundCheck(ctx, bc, reportRes)
					}
				}
			}
		}
	}

	return bc, nil
}

func (u *moderationUseCase) applyReportToBackgroundCheck(ctx context.Context, bc *entity.BackgroundCheck, report map[string]interface{}) {
	if repID, ok := report["id"].(string); ok && repID != "" {
		bc.CheckrReportID = repID
	}
	if st, ok := report["status"].(string); ok {
		checkrStatus := strings.ToUpper(st)
		switch checkrStatus {
		case "CLEAR", "COMPLETE":
			bc.Status = "clear"
		case "CONSIDER":
			bc.Status = "consider"
		case "SUSPENDED":
			bc.Status = "suspended"
		case "DISPUTE":
			bc.Status = "dispute"
		case "CANCELED":
			bc.Status = "canceled"
		default:
			bc.Status = "pending"
		}
	}

	if adj, ok := report["adjudication"].(string); ok {
		bc.Adjudication = adj
	}
	bc.Result = report
	now := time.Now()
	bc.LastSyncedAt = &now
	bc.SyncAttemptCount++
	bc.UpdatedAt = now

	_ = u.bgCheckRepo.Update(ctx, bc)

	// Update profile identity verified status
	if profile, err := u.profileRepo.GetByUserID(ctx, bc.ProviderID); err == nil && profile != nil {
		if bc.Status == "clear" {
			profile.IsIdentityVerified = true
			_ = u.profileRepo.Update(ctx, profile)
			if u.notifUC != nil {
				_ = u.notifUC.SendPushNotification(ctx, bc.ProviderID, "Background Check Verified 🎉", "Congratulations! Your Checkr screening has passed. Your provider profile is now active.", map[string]string{
					"type":     "BACKGROUND_CHECK_CLEARED",
					"check_id": bc.ID.String(),
				})
			}
		} else if bc.Status == "consider" || bc.Status == "suspended" || bc.Status == "canceled" {
			profile.IsIdentityVerified = false
			_ = u.profileRepo.Update(ctx, profile)
			if u.notifUC != nil {
				_ = u.notifUC.SendPushNotification(ctx, bc.ProviderID, "Background Check Update", "Your background screening status was updated. Please check your email for details from Checkr.", map[string]string{
					"type":   "BACKGROUND_CHECK_UPDATED",
					"status": bc.Status,
				})
			}
		}
	}
}

func (u *moderationUseCase) ProcessCheckrWebhook(ctx context.Context, signature string, rawBody []byte, payload map[string]interface{}) error {
	// Verify HMAC-SHA256 signature if secret/API key is configured
	secret := ""
	if u.cfg != nil {
		if u.cfg.CheckrWebhookSecret != "" {
			secret = u.cfg.CheckrWebhookSecret
		} else if u.cfg.CheckrAPIKey != "" {
			secret = u.cfg.CheckrAPIKey
		}
	}
	if secret != "" && len(rawBody) > 0 && signature != "" {
		cleanSig := strings.TrimSpace(strings.TrimPrefix(signature, "sha256="))
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(rawBody)
		expectedSig := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(cleanSig), []byte(expectedSig)) {
			return errors.New("invalid Checkr webhook signature")
		}
	}

	eventType, _ := payload["type"].(string)
	dataMap, _ := payload["data"].(map[string]interface{})
	objMap, _ := dataMap["object"].(map[string]interface{})
	if objMap == nil {
		return nil
	}

	// 1. Handle invitation events
	if strings.HasPrefix(eventType, "invitation.") {
		candID, _ := objMap["candidate_id"].(string)
		invID, _ := objMap["id"].(string)
		var bc *entity.BackgroundCheck
		if candID != "" {
			bc, _ = u.bgCheckRepo.GetByCandidateID(ctx, candID)
		}
		if bc != nil {
			if eventType == "invitation.completed" {
				bc.Status = "in_progress"
				_ = u.bgCheckRepo.Update(ctx, bc)
			} else if eventType == "invitation.expired" {
				bc.Status = "expired"
				_ = u.bgCheckRepo.Update(ctx, bc)
			}
		}
		_ = invID
		return nil
	}

	// 2. Handle report events
	if strings.HasPrefix(eventType, "report.") {
		reportID, _ := objMap["id"].(string)
		candidateID, _ := objMap["candidate_id"].(string)

		var bc *entity.BackgroundCheck
		if reportID != "" {
			bc, _ = u.bgCheckRepo.GetByReportID(ctx, reportID)
		}
		if bc == nil && candidateID != "" {
			bc, _ = u.bgCheckRepo.GetByCandidateID(ctx, candidateID)
		}

		if bc != nil {
			u.applyReportToBackgroundCheck(ctx, bc, objMap)
		}
	}

	return nil
}

// Audit UseCase
type auditUseCase struct {
	auditRepo repository.AuditLogRepository
}

func NewAuditUseCase(auditRepo repository.AuditLogRepository) domainUsecase.AuditUseCase {
	return &auditUseCase{auditRepo: auditRepo}
}

func (u *auditUseCase) LogAction(ctx context.Context, userID *uuid.UUID, action, resourceType, resourceID, ip string, details map[string]interface{}) error {
	log := entity.AuditLog{
		ID:           uuid.New(),
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    ip,
		Details:      details,
		CreatedAt:    time.Now(),
	}
	return u.auditRepo.Create(ctx, &log)
}

func (u *auditUseCase) GetLogs(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]entity.AuditLog, error) {
	return u.auditRepo.List(ctx, userID, limit, offset)
}
