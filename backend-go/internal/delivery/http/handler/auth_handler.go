package handler

import (
	"net/http"
	"strings"

	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authUC domainUsecase.AuthUseCase
}

func NewAuthHandler(authUC domainUsecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	UserType  string `json:"user_type"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	result, err := h.authUC.Register(c.Request.Context(), domainUsecase.RegisterInput{
		Email:     req.Email,
		Password:  req.Password,
		UserType:  req.UserType,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	})

	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, "User registered successfully", result)
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Email and password are required")
		return
	}

	result, err := h.authUC.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	// Returns SimpleJWT compatible access/refresh + user & profile metadata
	response.JSON(c, http.StatusOK, result)
}

type OTPVerifyRequest struct {
	Email   string `json:"email" binding:"required,email"`
	OTPCode string `json:"otp_code"`
	Code    string `json:"code"`
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req OTPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Email and code are required")
		return
	}

	code := req.OTPCode
	if code == "" {
		code = req.Code
	}

	_, err := h.authUC.VerifyOTP(c.Request.Context(), req.Email, code)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"detail": "Email verified successfully."})
}

type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var req ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Email is required")
		return
	}

	_ = h.authUC.ResendOTP(c.Request.Context(), req.Email)
	response.JSON(c, http.StatusOK, gin.H{"detail": "If an account exists, a new OTP has been sent."})
}

type TokenRefreshRequest struct {
	Refresh string `json:"refresh" binding:"required"`
}

func (h *AuthHandler) TokenRefresh(c *gin.Context) {
	var req TokenRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Refresh token is required")
		return
	}

	newAccess, err := h.authUC.RefreshToken(c.Request.Context(), req.Refresh)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"access": newAccess})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "old_password and new_password are required")
		return
	}

	if err := h.authUC.ChangePassword(c.Request.Context(), userUUID, req.OldPassword, req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"detail": "Password updated successfully."})
}

func (h *AuthHandler) PasswordResetRequest(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Valid email required")
		return
	}

	_ = h.authUC.PasswordResetRequest(c.Request.Context(), req.Email)
	response.JSON(c, http.StatusOK, gin.H{"detail": "If an account exists with this email, a reset OTP has been sent."})
}

func (h *AuthHandler) PasswordResetConfirm(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		OTPCode     string `json:"otp_code"`
		Token       string `json:"token"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload")
		return
	}

	code := req.OTPCode
	if code == "" {
		code = req.Token
	}

	if err := h.authUC.PasswordResetConfirm(c.Request.Context(), req.Email, code, req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"detail": "Password has been reset successfully."})
}

type socialLoginPayload struct {
	Token     string `json:"token"`
	IDToken   string `json:"id_token"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req socialLoginPayload
	_ = c.ShouldBindJSON(&req)

	token := req.Token
	if token == "" {
		token = req.IDToken
	}

	name := req.Name
	if name == "" {
		name = strings.TrimSpace(req.FirstName + " " + req.LastName)
	}

	result, err := h.authUC.SocialLogin(c.Request.Context(), "google", token, req.Email, name)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, result)
}

func (h *AuthHandler) AppleLogin(c *gin.Context) {
	var req socialLoginPayload
	_ = c.ShouldBindJSON(&req)

	token := req.Token
	if token == "" {
		token = req.IDToken
	}

	name := req.Name
	if name == "" {
		name = strings.TrimSpace(req.FirstName + " " + req.LastName)
	}

	result, err := h.authUC.SocialLogin(c.Request.Context(), "apple", token, req.Email, name)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	response.JSON(c, http.StatusOK, gin.H{"detail": "Successfully logged out."})
}

func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Invalid user authentication")
		return
	}

	if err := h.authUC.DeleteAccount(c.Request.Context(), userUUID); err != nil {
		response.InternalError(c, "Failed to delete account")
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "account_deleted"})
}

func (h *AuthHandler) ExportAccountData(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Invalid user authentication")
		return
	}

	data, err := h.authUC.ExportUserData(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, "Failed to export account data")
		return
	}

	response.JSON(c, http.StatusOK, data)
}

func (h *AuthHandler) RotateToken(c *gin.Context) {
	var req struct {
		Refresh string `json:"refresh" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Refresh token is required")
		return
	}

	result, err := h.authUC.RotateRefreshToken(c.Request.Context(), req.Refresh)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, result)
}

func (h *AuthHandler) AppleCallback(c *gin.Context) {
	response.JSON(c, http.StatusOK, gin.H{"status": "received"})
}

