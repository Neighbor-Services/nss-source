package usecase

import (
	"context"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
)

type RegisterInput struct {
	Email     string
	Password  string
	UserType  string
	FirstName string
	LastName  string
	Phone     string
}

type AuthResult struct {
	User         *entity.User    `json:"user"`
	Profile      *entity.Profile `json:"profile"`
	AccessToken  string          `json:"access"`
	RefreshToken string          `json:"refresh"`
}

type AuthUseCase interface {
	Register(ctx context.Context, input RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, email, password string) (*AuthResult, error)
	VerifyOTP(ctx context.Context, email, code string) (*AuthResult, error)
	ResendOTP(ctx context.Context, email string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	RotateRefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	PasswordResetRequest(ctx context.Context, email string) error
	PasswordResetConfirm(ctx context.Context, email, otpCode, newPassword string) error
	SocialLogin(ctx context.Context, provider, token, email, name string) (*AuthResult, error)
	DeleteAccount(ctx context.Context, userID uuid.UUID) error
	ExportUserData(ctx context.Context, userID uuid.UUID) (*entity.GDPRUserData, error)
}
