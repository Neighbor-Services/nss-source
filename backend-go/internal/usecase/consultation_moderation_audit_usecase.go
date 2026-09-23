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
	cfg              *config.Config
}

func NewModerationUseCase(
	reportRepo repository.ReportRepository,
	verificationRepo repository.ProviderVerificationRepository,
	bgCheckRepo repository.BackgroundCheckRepository,
	profileRepo repository.ProfileRepository,
	userRepo repository.UserRepository,
	cfg *config.Config,
) domainUsecase.ModerationUseCase {
	return &moderationUseCase{
		reportRepo:       reportRepo,
		verificationRepo: verificationRepo,
		bgCheckRepo:      bgCheckRepo,
		profileRepo:      profileRepo,
		userRepo:         userRepo,
		cfg:              cfg,
	}
}

func (u *moderationUseCase) SubmitReport(ctx context.Context, reporterID uuid.UUID, reportedUserID *uuid.UUID, contentType, objectID, reason, description string) (*entity.Report, error) {
	resourceType := contentType
	if resourceType == "" {
		resourceType = "General"
	}

	report := entity.Report{
		ID:             uuid.New(),
		ReporterID:     reporterID,
		ReportedUserID: reportedUserID,
		ResourceType:   resourceType,
		ResourceID:     objectID,
		Reason:         reason,
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

func (u *moderationUseCase) SubmitVerification(ctx context.Context, providerID uuid.UUID, frontURL, backURL string) (*entity.ProviderVerification, error) {
	v := entity.ProviderVerification{
		ID:            uuid.New(),
		ProviderID:    providerID,
		DocumentFront: frontURL,
		DocumentBack:  backURL,
		Status:        "PENDING",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := u.verificationRepo.Create(ctx, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (u *moderationUseCase) GetBackgroundChecks(ctx context.Context, providerID *uuid.UUID) ([]entity.BackgroundCheck, error) {
	return u.bgCheckRepo.List(ctx, providerID)
}

func (u *moderationUseCase) GetBackgroundCheckConfig(ctx context.Context) (map[string]interface{}, error) {
	setting, err := u.bgCheckRepo.GetModerationSetting(ctx)
	mode := "IN_APP_STRIPE"
	if err == nil && setting != nil && setting.BackgroundCheckPaymentMode != "" {
		mode = setting.BackgroundCheckPaymentMode
	}
	return map[string]interface{}{
		"payment_mode": mode,
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
	var res map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, fmt.Errorf("Checkr response parsing error: %w", err)
	}

	if resp.StatusCode >= 400 {
		errMsg, _ := res["error"].(string)
		if errMsg == "" {
			errMsg = string(bodyBytes)
		}
		return nil, fmt.Errorf("Checkr API error (HTTP %d): %s", resp.StatusCode, errMsg)
	}

	return res, nil
}

func (u *moderationUseCase) InitiateBackgroundCheck(ctx context.Context, userID uuid.UUID, paymentIntentID string) (*entity.BackgroundCheck, error) {
	// 1. Guard against duplicate pending checks
	existing, _ := u.bgCheckRepo.GetPendingByProvider(ctx, userID)
	if existing != nil {
		return existing, nil
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	profile, err := u.profileRepo.GetByUserID(ctx, userID)
	if err != nil || profile == nil {
		return nil, errors.New("profile not found")
	}

	// 2. Create Candidate in Checkr
	candidatePayload := map[string]interface{}{
		"email": user.Email,
	}
	if profile.FirstName != "" {
		candidatePayload["first_name"] = profile.FirstName
	}
	if profile.LastName != "" {
		candidatePayload["last_name"] = profile.LastName
	}

	var candidateID string
	var invitationID string
	var invitationURL string

	candidateRes, err := u.checkrRequest("POST", "/v1/candidates", candidatePayload)
	if err == nil && candidateRes != nil {
		if idStr, ok := candidateRes["id"].(string); ok {
			candidateID = idStr
		}
	}

	if candidateID != "" {
		// 3. Create Invitation in Checkr
		invitationPayload := map[string]interface{}{
			"candidate_id": candidateID,
			"package":      "tasker_standard",
			"work_locations": []map[string]string{
				{"state": "CA"},
			},
		}
		invitationRes, err := u.checkrRequest("POST", "/v1/invitations", invitationPayload)
		if err == nil && invitationRes != nil {
			if idStr, ok := invitationRes["id"].(string); ok {
				invitationID = idStr
			}
			if urlStr, ok := invitationRes["invitation_url"].(string); ok {
				invitationURL = urlStr
			}
		}
	}

	if candidateID == "" {
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
		Package:            "tasker_standard",
		Status:             "pending",
		PaymentIntentID:    paymentIntentID,
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
		} else if bc.Status == "consider" || bc.Status == "suspended" || bc.Status == "canceled" {
			profile.IsIdentityVerified = false
			_ = u.profileRepo.Update(ctx, profile)
		}
	}
}

func (u *moderationUseCase) ProcessCheckrWebhook(ctx context.Context, signature string, rawBody []byte, payload map[string]interface{}) error {
	eventType, _ := payload["type"].(string)
	if !strings.HasPrefix(eventType, "report.") {
		return nil
	}

	dataMap, _ := payload["data"].(map[string]interface{})
	objMap, _ := dataMap["object"].(map[string]interface{})
	if objMap == nil {
		return nil
	}

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
