package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"sync"

	"backend-go/pkg/auth"
	"backend-go/pkg/email"
	"github.com/google/uuid"
)

type otpAttemptTracker struct {
	count     int
	firstFail time.Time
}

type authUseCase struct {
	userRepo    repository.UserRepository
	profileRepo repository.ProfileRepository
	walletRepo  repository.WalletRepository
	cfg         *config.Config
	otpMu       sync.Mutex
	otpAttempts map[string]*otpAttemptTracker
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	profileRepo repository.ProfileRepository,
	walletRepo repository.WalletRepository,
	cfg *config.Config,
) domainUsecase.AuthUseCase {
	return &authUseCase{
		userRepo:    userRepo,
		profileRepo: profileRepo,
		walletRepo:  walletRepo,
		cfg:         cfg,
		otpAttempts: make(map[string]*otpAttemptTracker),
	}
}

func (u *authUseCase) Register(ctx context.Context, input domainUsecase.RegisterInput) (*domainUsecase.AuthResult, error) {
	emailStr := strings.ToLower(strings.TrimSpace(input.Email))
	if emailStr == "" {
		return nil, errors.New("email is required")
	}

	count, err := u.userRepo.CountByEmail(ctx, emailStr)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("an account with this email already exists")
	}

	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to process password: %w", err)
	}

	userType := strings.ToUpper(input.UserType)
	if userType != "PROVIDER" {
		userType = "SEEKER"
	}

	otp := fmt.Sprintf("%04d", rand.Intn(10000))
	otpExpiry := time.Now().Add(10 * time.Minute)

	user := entity.User{
		ID:         uuid.New(),
		Email:      emailStr,
		Password:   hashedPassword,
		IsActive:   true,
		IsVerified: false,
		OTPCode:    otp,
		OTPExpiry:  &otpExpiry,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := u.userRepo.Create(ctx, &user); err != nil {
		return nil, err
	}

	// Dispatch OTP email asynchronously
	log.Printf("[AUTH] Generated registration OTP %s for %s", otp, user.Email)
	if u.cfg != nil && u.cfg.SMTPHost != "" {
		go func(to, code string) {
			if err := email.SendOTPEmail(&email.Config{
				Host:     u.cfg.SMTPHost,
				Port:     u.cfg.SMTPPort,
				User:     u.cfg.SMTPUser,
				Password: u.cfg.SMTPPassword,
				From:     u.cfg.EmailFrom,
			}, to, code); err != nil {
				log.Printf("[EMAIL ERROR] SendOTPEmail to %s failed: %v", to, err)
			} else {
				log.Printf("[EMAIL SUCCESS] Sent registration OTP email to %s", to)
			}
		}(user.Email, otp)
	} else {
		log.Printf("[EMAIL WARNING] SMTP not configured, OTP for %s is %s", user.Email, otp)
	}

	profile := entity.Profile{
		ID:        uuid.New(),
		UserID:    user.ID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Phone:     input.Phone,
		UserType:  userType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := u.profileRepo.Create(ctx, &profile); err != nil {
		return nil, err
	}

	// Create initial wallet
	wallet := entity.Wallet{
		ID:        uuid.New(),
		UserID:    user.ID,
		Balance:   0.0,
		Currency:  "USD",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = u.walletRepo.Create(ctx, &wallet)

	accessToken, refreshToken, _ := auth.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		userType,
		u.cfg.JWTSecret,
		u.cfg.JWTSecret,
		24,
		30,
	)

	return &domainUsecase.AuthResult{
		User:         &user,
		Profile:      &profile,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *authUseCase) Login(ctx context.Context, email, password string) (*domainUsecase.AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, errors.New("Invalid email or password.")
	}

	if !auth.CheckPassword(user.Password, password) {
		return nil, errors.New("Invalid email or password.")
	}

	if !user.IsVerified {
		return nil, errors.New("Email not verified. Please verify your email before logging in.")
	}

	profile, err := u.profileRepo.GetByUserID(ctx, user.ID)
	if err != nil || profile == nil {
		profile = &entity.Profile{
			ID:       uuid.New(),
			UserID:   user.ID,
			UserType: "SEEKER",
		}
	}

	accessToken, refreshToken, err := auth.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		profile.UserType,
		u.cfg.JWTSecret,
		u.cfg.JWTSecret,
		24,
		30,
	)
	if err != nil {
		return nil, errors.New("failed to generate auth tokens")
	}

	return &domainUsecase.AuthResult{
		User:         user,
		Profile:      profile,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *authUseCase) VerifyOTP(ctx context.Context, email, code string) (*domainUsecase.AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	u.otpMu.Lock()
	tracker, exists := u.otpAttempts[email]
	if exists {
		if time.Since(tracker.firstFail) > 10*time.Minute {
			delete(u.otpAttempts, email)
		} else if tracker.count >= 5 {
			u.otpMu.Unlock()
			return nil, errors.New("Too many failed attempts. Account temporarily locked for 10 minutes.")
		}
	}
	u.otpMu.Unlock()

	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil || user.OTPCode != code || (user.OTPExpiry != nil && time.Now().After(*user.OTPExpiry)) {
		u.otpMu.Lock()
		if tr, ok := u.otpAttempts[email]; ok {
			tr.count++
		} else {
			u.otpAttempts[email] = &otpAttemptTracker{count: 1, firstFail: time.Now()}
		}
		u.otpMu.Unlock()
		return nil, errors.New("Invalid or expired OTP.")
	}

	// Success - clear lockout tracker
	u.otpMu.Lock()
	delete(u.otpAttempts, email)
	u.otpMu.Unlock()

	user.IsVerified = true
	user.OTPCode = ""
	user.OTPExpiry = nil
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	profile, err := u.profileRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		profile = &entity.Profile{ID: uuid.New(), UserID: user.ID, UserType: "SEEKER"}
	}

	accessToken, refreshToken, _ := auth.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		profile.UserType,
		u.cfg.JWTSecret,
		u.cfg.JWTSecret,
		24,
		30,
	)

	return &domainUsecase.AuthResult{
		User:         user,
		Profile:      profile,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *authUseCase) ResendOTP(ctx context.Context, emailStr string) error {
	emailStr = strings.ToLower(strings.TrimSpace(emailStr))
	user, err := u.userRepo.GetByEmail(ctx, emailStr)
	if err != nil || user == nil {
		return nil // Don't leak user existence per Django behavior
	}

	otp := fmt.Sprintf("%04d", rand.Intn(10000))
	otpExpiry := time.Now().Add(10 * time.Minute)
	user.OTPCode = otp
	user.OTPExpiry = &otpExpiry

	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}

	log.Printf("[AUTH] Resent OTP %s for %s", otp, user.Email)
	if u.cfg != nil && u.cfg.SMTPHost != "" {
		go func(to, code string) {
			if err := email.SendOTPEmail(&email.Config{
				Host:     u.cfg.SMTPHost,
				Port:     u.cfg.SMTPPort,
				User:     u.cfg.SMTPUser,
				Password: u.cfg.SMTPPassword,
				From:     u.cfg.EmailFrom,
			}, to, code); err != nil {
				log.Printf("[EMAIL ERROR] Resend OTP email to %s failed: %v", to, err)
			} else {
				log.Printf("[EMAIL SUCCESS] Sent OTP resend email to %s", to)
			}
		}(user.Email, otp)
	} else {
		log.Printf("[EMAIL WARNING] SMTP not configured, Resend OTP for %s is %s", user.Email, otp)
	}

	return nil
}

func (u *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := auth.ValidateToken(refreshToken, u.cfg.JWTSecret)
	if err != nil {
		return "", errors.New("invalid or expired refresh token")
	}

	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", errors.New("invalid token payload")
	}

	user, err := u.userRepo.GetByID(ctx, uid)
	if err != nil || user == nil {
		return "", errors.New("user not found")
	}

	profile, _ := u.profileRepo.GetByUserID(ctx, user.ID)
	userType := "SEEKER"
	if profile != nil {
		userType = profile.UserType
	}

	access, _, err := auth.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		userType,
		u.cfg.JWTSecret,
		u.cfg.JWTSecret,
		24,
		30,
	)
	return access, err
}

func (u *authUseCase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	if !auth.CheckPassword(user.Password, oldPassword) {
		return errors.New("Wrong password.")
	}

	hashed, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashed
	return u.userRepo.Update(ctx, user)
}

func (u *authUseCase) PasswordResetRequest(ctx context.Context, emailStr string) error {
	emailStr = strings.ToLower(strings.TrimSpace(emailStr))
	user, err := u.userRepo.GetByEmail(ctx, emailStr)
	if err != nil || user == nil {
		log.Printf("[AUTH] PasswordResetRequest: user %s not found in DB (%v)", emailStr, err)
		return nil
	}

	otp := fmt.Sprintf("%04d", rand.Intn(10000))
	otpExpiry := time.Now().Add(10 * time.Minute)
	user.OTPCode = otp
	user.OTPExpiry = &otpExpiry

	if err := u.userRepo.Update(ctx, user); err != nil {
		log.Printf("[AUTH ERROR] Failed to save reset OTP for %s: %v", user.Email, err)
		return err
	}

	log.Printf("[AUTH] Generated Password Reset OTP %s for %s", otp, user.Email)
	if u.cfg != nil && u.cfg.SMTPHost != "" {
		go func(to, code string) {
			if err := email.SendPasswordResetEmail(&email.Config{
				Host:     u.cfg.SMTPHost,
				Port:     u.cfg.SMTPPort,
				User:     u.cfg.SMTPUser,
				Password: u.cfg.SMTPPassword,
				From:     u.cfg.EmailFrom,
			}, to, code); err != nil {
				log.Printf("[EMAIL ERROR] SendPasswordResetEmail to %s failed: %v", to, err)
			} else {
				log.Printf("[EMAIL SUCCESS] Sent password reset email to %s", to)
			}
		}(user.Email, otp)
	} else {
		log.Printf("[EMAIL WARNING] SMTP not configured, Password Reset OTP for %s is %s", user.Email, otp)
	}

	return nil
}

func (u *authUseCase) PasswordResetConfirm(ctx context.Context, email, otpCode, newPassword string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return errors.New("Invalid or expired OTP.")
	}

	if user.OTPCode != otpCode {
		return errors.New("Invalid or expired OTP.")
	}

	if user.OTPExpiry != nil && time.Now().After(*user.OTPExpiry) {
		return errors.New("Invalid or expired OTP.")
	}

	hashed, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashed
	user.OTPCode = ""
	user.OTPExpiry = nil
	return u.userRepo.Update(ctx, user)
}

func (u *authUseCase) SocialLogin(ctx context.Context, provider, token, email, name string) (*domainUsecase.AuthResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		log.Printf("[SOCIAL LOGIN] %s: email is empty — rejecting", provider)
		return nil, errors.New("email is required from social provider")
	}

	log.Printf("[SOCIAL LOGIN] %s: attempt for %s (name=%q)", provider, email, name)

	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		// Create new user & profile
		log.Printf("[SOCIAL LOGIN] %s: creating new user for %s", provider, email)
		newUser := entity.User{
			ID:         uuid.New(),
			Email:      email,
			Password:   "",
			IsActive:   true,
			IsVerified: true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if createErr := u.userRepo.Create(ctx, &newUser); createErr != nil {
			log.Printf("[SOCIAL LOGIN ERROR] %s: failed to create user for %s: %v", provider, email, createErr)
		}
		user = &newUser

		// Split "First Last" name
		firstName := name
		lastName := ""
		if idx := strings.Index(name, " "); idx != -1 {
			firstName = name[:idx]
			lastName = name[idx+1:]
		}

		newProfile := entity.Profile{
			ID:        uuid.New(),
			UserID:    user.ID,
			FirstName: firstName,
			LastName:  lastName,
			UserType:  "SEEKER",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if createErr := u.profileRepo.Create(ctx, &newProfile); createErr != nil {
			log.Printf("[SOCIAL LOGIN ERROR] %s: failed to create profile for %s: %v", provider, email, createErr)
		}
	} else {
		log.Printf("[SOCIAL LOGIN] %s: found existing user for %s (id=%s)", provider, email, user.ID)
	}

	profile, _ := u.profileRepo.GetByUserID(ctx, user.ID)
	userType := "SEEKER"
	if profile != nil {
		userType = profile.UserType
	}

	access, refresh, err := auth.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		userType,
		u.cfg.JWTSecret,
		u.cfg.JWTSecret,
		24,
		30,
	)
	if err != nil {
		return nil, err
	}

	log.Printf("[SOCIAL LOGIN] %s: success for %s (userType=%s)", provider, email, userType)

	return &domainUsecase.AuthResult{
		User:         user,
		Profile:      profile,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

func (u *authUseCase) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	return u.userRepo.Delete(ctx, userID)
}

func (u *authUseCase) RotateRefreshToken(ctx context.Context, refreshToken string) (*domainUsecase.AuthResult, error) {
	claims, err := auth.ValidateToken(refreshToken, u.cfg.JWTSecret)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid token payload")
	}

	user, err := u.userRepo.GetByID(ctx, uid)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	profile, _ := u.profileRepo.GetByUserID(ctx, user.ID)
	userType := "SEEKER"
	if profile != nil {
		userType = profile.UserType
	}

	access, newRefresh, err := auth.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		userType,
		u.cfg.JWTSecret,
		u.cfg.JWTSecret,
		24,
		30,
	)
	if err != nil {
		return nil, err
	}

	return &domainUsecase.AuthResult{
		User:         user,
		Profile:      profile,
		AccessToken:  access,
		RefreshToken: newRefresh,
	}, nil
}

func (u *authUseCase) ExportUserData(ctx context.Context, userID uuid.UUID) (*entity.GDPRUserData, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	profile, _ := u.profileRepo.GetByUserID(ctx, userID)
	wallet, _ := u.walletRepo.GetByUserID(ctx, userID)

	return &entity.GDPRUserData{
		User:       *user,
		Profile:    profile,
		Wallet:     wallet,
		ExportedAt: time.Now().UTC(),
	}, nil
}

