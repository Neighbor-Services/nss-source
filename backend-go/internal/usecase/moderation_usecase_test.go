package usecase_test

import (
	"context"
	"testing"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/internal/usecase"
	"github.com/google/uuid"
)

type mockReportRepo struct {
	reports []*entity.Report
}

func (m *mockReportRepo) Create(ctx context.Context, report *entity.Report) error {
	m.reports = append(m.reports, report)
	return nil
}
func (m *mockReportRepo) List(ctx context.Context, status string, limit, offset int) ([]entity.Report, error) {
	return nil, nil
}

type mockVerificationRepo struct {
	verifs []*entity.ProviderVerification
}

func (m *mockVerificationRepo) Create(ctx context.Context, v *entity.ProviderVerification) error {
	m.verifs = append(m.verifs, v)
	return nil
}
func (m *mockVerificationRepo) List(ctx context.Context, providerID *uuid.UUID) ([]entity.ProviderVerification, error) {
	return nil, nil
}
func (m *mockVerificationRepo) Update(ctx context.Context, v *entity.ProviderVerification) error {
	return nil
}

type mockBgCheckRepo struct {
	checks  map[uuid.UUID]*entity.BackgroundCheck
	setting *entity.ModerationSetting
}

func (m *mockBgCheckRepo) Create(ctx context.Context, check *entity.BackgroundCheck) error {
	if m.checks == nil {
		m.checks = make(map[uuid.UUID]*entity.BackgroundCheck)
	}
	m.checks[check.ID] = check
	return nil
}
func (m *mockBgCheckRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.BackgroundCheck, error) {
	return m.checks[id], nil
}
func (m *mockBgCheckRepo) GetByCandidateID(ctx context.Context, candidateID string) (*entity.BackgroundCheck, error) {
	for _, c := range m.checks {
		if c.CheckrCandidateID == candidateID {
			return c, nil
		}
	}
	return nil, nil
}
func (m *mockBgCheckRepo) GetByReportID(ctx context.Context, reportID string) (*entity.BackgroundCheck, error) {
	for _, c := range m.checks {
		if c.CheckrReportID == reportID {
			return c, nil
		}
	}
	return nil, nil
}
func (m *mockBgCheckRepo) GetPendingByProvider(ctx context.Context, providerID uuid.UUID) (*entity.BackgroundCheck, error) {
	for _, c := range m.checks {
		if c.ProviderID == providerID && c.Status == "pending" {
			return c, nil
		}
	}
	return nil, nil
}
func (m *mockBgCheckRepo) List(ctx context.Context, providerID *uuid.UUID) ([]entity.BackgroundCheck, error) {
	var list []entity.BackgroundCheck
	for _, c := range m.checks {
		list = append(list, *c)
	}
	return list, nil
}
func (m *mockBgCheckRepo) Update(ctx context.Context, check *entity.BackgroundCheck) error {
	m.checks[check.ID] = check
	return nil
}
func (m *mockBgCheckRepo) GetModerationSetting(ctx context.Context) (*entity.ModerationSetting, error) {
	if m.setting == nil {
		return &entity.ModerationSetting{
			ID:                         uuid.New(),
			BackgroundCheckPaymentMode: "IN_APP_STRIPE",
			BackgroundCheckFee:         29.99,
		}, nil
	}
	return m.setting, nil
}

type mockUserRepoForMod struct {
	users map[uuid.UUID]*entity.User
}

func (m *mockUserRepoForMod) Create(ctx context.Context, user *entity.User) error { return nil }
func (m *mockUserRepoForMod) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return m.users[id], nil
}
func (m *mockUserRepoForMod) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepoForMod) Update(ctx context.Context, user *entity.User) error { return nil }
func (m *mockUserRepoForMod) Delete(ctx context.Context, id uuid.UUID) error     { return nil }
func (m *mockUserRepoForMod) CountByEmail(ctx context.Context, email string) (int64, error) {
	return 0, nil
}

func TestModerationUseCase_BackgroundCheckFlow(t *testing.T) {
	providerID := uuid.New()
	reportRepo := &mockReportRepo{}
	verifRepo := &mockVerificationRepo{}
	bgRepo := &mockBgCheckRepo{checks: make(map[uuid.UUID]*entity.BackgroundCheck)}
	profileRepo := &mockProfileRepo{profiles: map[uuid.UUID]*entity.Profile{
		providerID: {ID: uuid.New(), UserID: providerID, FirstName: "Jane", LastName: "Doe", UserType: "PROVIDER"},
	}}
	userRepo := &mockUserRepoForMod{users: map[uuid.UUID]*entity.User{
		providerID: {ID: providerID, Email: "provider@example.com"},
	}}
	cfg := &config.Config{CheckrAPIKey: ""}

	modUC := usecase.NewModerationUseCase(reportRepo, verifRepo, bgRepo, profileRepo, userRepo, cfg)
	ctx := context.Background()

	// 1. Get Config
	setting, err := modUC.GetBackgroundCheckConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get background check config: %v", err)
	}
	if setting["payment_mode"] != "IN_APP_STRIPE" {
		t.Errorf("Expected IN_APP_STRIPE, got %v", setting["payment_mode"])
	}

	// 2. Initiate Background Check
	check, err := modUC.InitiateBackgroundCheck(ctx, providerID, "pi_test_12345")
	if err != nil {
		t.Fatalf("Failed to initiate background check: %v", err)
	}
	if check.ProviderID != providerID {
		t.Errorf("Expected ProviderID %s, got %s", providerID, check.ProviderID)
	}
	if check.Status != "pending" {
		t.Errorf("Expected pending status, got %s", check.Status)
	}

	// 3. Webhook simulation (Report CLEAR)
	payload := map[string]interface{}{
		"type": "report.completed",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":           "rep_test_12345",
				"candidate_id": check.CheckrCandidateID,
				"status":       "CLEAR",
				"adjudication": "engaged",
			},
		},
	}
	err = modUC.ProcessCheckrWebhook(ctx, "", nil, payload)
	if err != nil {
		t.Fatalf("Failed to process webhook: %v", err)
	}

	if check.Status != "clear" {
		t.Errorf("Expected status clear, got %s", check.Status)
	}

	profile, _ := profileRepo.GetByUserID(ctx, providerID)
	if !profile.IsIdentityVerified {
		t.Errorf("Expected provider profile to be identity verified after CLEAR report")
	}
}
