package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-go/internal/delivery/http/handler"
	"backend-go/internal/domain/entity"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type stubAuthUseCase struct {
	loginFunc func(ctx context.Context, email, password string) (*domainUsecase.AuthResult, error)
}

func (s *stubAuthUseCase) Register(ctx context.Context, input domainUsecase.RegisterInput) (*domainUsecase.AuthResult, error) {
	return nil, nil
}

func (s *stubAuthUseCase) Login(ctx context.Context, email, password string) (*domainUsecase.AuthResult, error) {
	if s.loginFunc != nil {
		return s.loginFunc(ctx, email, password)
	}
	return nil, errors.New("not implemented")
}

func (s *stubAuthUseCase) VerifyOTP(ctx context.Context, email, code string) (*domainUsecase.AuthResult, error) {
	return nil, nil
}

func (s *stubAuthUseCase) ResendOTP(ctx context.Context, email string) error {
	return nil
}

func (s *stubAuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	return "", nil
}

func (s *stubAuthUseCase) RotateRefreshToken(ctx context.Context, refreshToken string) (*domainUsecase.AuthResult, error) {
	return nil, nil
}

func (s *stubAuthUseCase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	return nil
}

func (s *stubAuthUseCase) PasswordResetRequest(ctx context.Context, email string) error {
	return nil
}

func (s *stubAuthUseCase) PasswordResetConfirm(ctx context.Context, email, otpCode, newPassword string) error {
	return nil
}

func (s *stubAuthUseCase) SocialLogin(ctx context.Context, provider, token, email, name string) (*domainUsecase.AuthResult, error) {
	return nil, nil
}

func (s *stubAuthUseCase) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (s *stubAuthUseCase) ExportUserData(ctx context.Context, userID uuid.UUID) (*entity.GDPRUserData, error) {
	return nil, nil
}

func setupAuthRouter(authUC domainUsecase.AuthUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authH := handler.NewAuthHandler(authUC)
	r.POST("/api/v1/accounts/login/", authH.Login)
	r.POST("/api/v1/admin/login", authH.AdminLogin)
	return r
}

func TestAdminLogin_RejectsStandardUserWith403(t *testing.T) {
	middleware.GetDefaultLoginBlocker().Reset()
	stubUC := &stubAuthUseCase{
		loginFunc: func(ctx context.Context, email, password string) (*domainUsecase.AuthResult, error) {
			return &domainUsecase.AuthResult{
				User: &entity.User{
					ID:          uuid.New(),
					Email:       email,
					IsStaff:     false,
					IsSuperuser: false,
					IsActive:    true,
				},
				AccessToken:  "valid-access-token",
				RefreshToken: "valid-refresh-token",
			}, nil
		},
	}
	r := setupAuthRouter(stubUC)

	body, _ := json.Marshal(map[string]interface{}{
		"email":    "user@neighborservice.com",
		"password": "validpass123",
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "ADMIN_ACCESS_REQUIRED")
	assert.Contains(t, w.Body.String(), "Access denied: You do not have permission to access the administrator portal.")
}

func TestAdminLogin_AllowsStaffAndSuperuser(t *testing.T) {
	middleware.GetDefaultLoginBlocker().Reset()
	stubUC := &stubAuthUseCase{
		loginFunc: func(ctx context.Context, email, password string) (*domainUsecase.AuthResult, error) {
			return &domainUsecase.AuthResult{
				User: &entity.User{
					ID:          uuid.New(),
					Email:       email,
					IsStaff:     true,
					IsSuperuser: false,
					IsActive:    true,
				},
				AccessToken:  "admin-access-token",
				RefreshToken: "admin-refresh-token",
			}, nil
		},
	}
	r := setupAuthRouter(stubUC)

	body, _ := json.Marshal(map[string]interface{}{
		"email":    "admin@neighborservice.com",
		"password": "secretpass",
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "admin-access-token")
}

func TestAdminLogin_BruteForceLockoutAfterFiveFailures(t *testing.T) {
	middleware.GetDefaultLoginBlocker().Reset()
	stubUC := &stubAuthUseCase{
		loginFunc: func(ctx context.Context, email, password string) (*domainUsecase.AuthResult, error) {
			return nil, errors.New("Invalid email or password.")
		},
	}
	r := setupAuthRouter(stubUC)

	targetEmail := "admin@victim.com"

	// 5 failed attempts
	for i := 1; i <= 5; i++ {
		body, _ := json.Marshal(map[string]interface{}{
			"email":    targetEmail,
			"password": "wrongpass",
		})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "192.0.2.10:44321"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if i < 5 {
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		} else {
			// 5th failure triggers lockout
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
			assert.Contains(t, w.Body.String(), "TOO_MANY_FAILED_ATTEMPTS")
			assert.NotEmpty(t, w.Header().Get("Retry-After"))
		}
	}

	// 6th attempt should be blocked immediately by brute force prevention without calling Login
	body, _ := json.Marshal(map[string]interface{}{
		"email":    targetEmail,
		"password": "wrongpass",
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.0.2.10:44321"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "TOO_MANY_FAILED_ATTEMPTS")
}
