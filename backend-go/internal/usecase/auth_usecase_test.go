package usecase_test

import (
	"context"
	"testing"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/internal/usecase"
	"github.com/google/uuid"
)

// MockUserRepository implements repository.UserRepository for testing
type mockUserRepo struct {
	users map[string]*entity.User
}

func (m *mockUserRepo) Create(ctx context.Context, u *entity.User) error {
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepo) Update(ctx context.Context, u *entity.User) error {
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	for email, u := range m.users {
		if u.ID == id {
			delete(m.users, email)
		}
	}
	return nil
}

func (m *mockUserRepo) CountByEmail(ctx context.Context, email string) (int64, error) {
	if _, ok := m.users[email]; ok {
		return 1, nil
	}
	return 0, nil
}

// MockProfileRepository
type mockProfileRepo struct {
	profiles map[uuid.UUID]*entity.Profile
}

func (m *mockProfileRepo) Create(ctx context.Context, p *entity.Profile) error {
	m.profiles[p.UserID] = p
	return nil
}

func (m *mockProfileRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Profile, error) {
	return m.profiles[userID], nil
}

func (m *mockProfileRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Profile, error) {
	for _, p := range m.profiles {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockProfileRepo) Update(ctx context.Context, p *entity.Profile) error {
	m.profiles[p.UserID] = p
	return nil
}

func (m *mockProfileRepo) UpdateCatalogServices(ctx context.Context, profileID uuid.UUID, serviceIDs []uuid.UUID) error {
	return nil
}

func (m *mockProfileRepo) List(ctx context.Context, params repository.ProfileFilterParams) ([]entity.Profile, error) {
	var list []entity.Profile
	for _, p := range m.profiles {
		if params.UserType == "" || p.UserType == params.UserType {
			list = append(list, *p)
		}
	}
	return list, nil
}

func (m *mockProfileRepo) ListProviders(ctx context.Context, categorySlug string, limit, offset int) ([]entity.Profile, error) {
	var list []entity.Profile
	for _, p := range m.profiles {
		if p.UserType == "PROVIDER" {
			list = append(list, *p)
		}
	}
	return list, nil
}

// MockWalletRepository
type mockWalletRepo struct {
	wallets map[uuid.UUID]*entity.Wallet
}

func (m *mockWalletRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Wallet, error) {
	if m.wallets != nil {
		for _, w := range m.wallets {
			if w.ID == id {
				return w, nil
			}
		}
	}
	return &entity.Wallet{ID: id, Balance: 0.0, Currency: "USD"}, nil
}

func (m *mockWalletRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	if m.wallets != nil {
		if w, ok := m.wallets[userID]; ok {
			return w, nil
		}
	}
	return &entity.Wallet{ID: uuid.New(), UserID: userID, Balance: 0.0, Currency: "USD"}, nil
}
func (m *mockWalletRepo) Create(ctx context.Context, w *entity.Wallet) error {
	if m.wallets != nil {
		m.wallets[w.UserID] = w
	}
	return nil
}
func (m *mockWalletRepo) Update(ctx context.Context, w *entity.Wallet) error {
	if m.wallets != nil {
		m.wallets[w.UserID] = w
	}
	return nil
}

func TestAuthUseCase_RegisterAndLogin(t *testing.T) {
	userRepo := &mockUserRepo{users: make(map[string]*entity.User)}
	profileRepo := &mockProfileRepo{profiles: make(map[uuid.UUID]*entity.Profile)}
	walletRepo := &mockWalletRepo{}
	cfg := &config.Config{JWTSecret: "test-secret-key-1234567890123456"}

	authUC := usecase.NewAuthUseCase(userRepo, profileRepo, walletRepo, cfg)

	// 1. Register User
	ctx := context.Background()
	regResult, err := authUC.Register(ctx, domainUsecase.RegisterInput{
		Email:     "provider@example.com",
		Password:  "SecurePassword123!",
		UserType:  "PROVIDER",
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "+1234567890",
	})

	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if regResult.User.Email != "provider@example.com" {
		t.Errorf("Expected email provider@example.com, got %s", regResult.User.Email)
	}
	if regResult.Profile.UserType != "PROVIDER" {
		t.Errorf("Expected userType PROVIDER, got %s", regResult.Profile.UserType)
	}

	// 2. Verify OTP (Django requires verification before login)
	otp := regResult.User.OTPCode
	_, err = authUC.VerifyOTP(ctx, "provider@example.com", otp)
	if err != nil {
		t.Fatalf("VerifyOTP failed: %v", err)
	}

	// 3. Login User
	loginResult, err := authUC.Login(ctx, "provider@example.com", "SecurePassword123!")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if loginResult.AccessToken == "" {
		t.Error("Expected access token on login")
	}

	// 4. Verify Login fails with wrong password
	_, err = authUC.Login(ctx, "provider@example.com", "WrongPassword")
	if err == nil {
		t.Error("Expected login to fail with wrong password")
	}

	// 5. Register with ONLY email and password (no user_type, no names, no phone)
	minimalReg, err := authUC.Register(ctx, domainUsecase.RegisterInput{
		Email:    "minimal@example.com",
		Password: "SecurePassword123!",
	})
	if err != nil {
		t.Fatalf("Register with only email and password failed: %v", err)
	}
	if minimalReg.User.Email != "minimal@example.com" {
		t.Errorf("Expected email minimal@example.com, got %s", minimalReg.User.Email)
	}
	if minimalReg.Profile.UserType != "SEEKER" {
		t.Errorf("Expected default userType SEEKER, got %s", minimalReg.Profile.UserType)
	}
	if minimalReg.AccessToken == "" || minimalReg.RefreshToken == "" {
		t.Error("Expected token pair on minimal registration")
	}
}
