package handler

import (
	"fmt"
	"net/http"
	"strings"

	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/internal/middleware"
	"backend-go/pkg/logger"
	"backend-go/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	AdminOnly bool   `json:"admin_only"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Email and password are required")
		return
	}

	clientIP := c.ClientIP()
	blocker := middleware.GetDefaultLoginBlocker()

	// Check if IP or email is locked out due to brute force
	if locked, remaining, reason := blocker.IsLocked(clientIP, req.Email); locked {
		middleware.LogSecurityBruteForce(c, clientIP, req.Email, fmt.Sprintf("Blocked login attempt during active lockout (%s remaining)", remaining))
		middleware.RespondBruteForceLocked(c, remaining, reason)
		return
	}

	result, err := h.authUC.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if strings.HasPrefix(err.Error(), "ACCOUNT_SUSPENDED") {
			reason := strings.TrimPrefix(err.Error(), "ACCOUNT_SUSPENDED: ")
			c.JSON(http.StatusForbidden, gin.H{
 "code": "ACCOUNT_SUSPENDED",
 "error": "Account suspended",
 "message": "Your account has been suspended by an administrator.",
 "suspension_reason": reason,
 "support_email": "support@neighborservice.com",
			})
			return
		}

		newlyLocked, rem := blocker.RecordFailure(clientIP, req.Email)
		if newlyLocked {
			middleware.LogSecurityBruteForce(c, clientIP, req.Email, "Exceeded 5 failed login attempts; IP/account locked for 15 minutes")
			middleware.RespondBruteForceLocked(c, rem, "attempts_exceeded")
			return
		}

		response.Unauthorized(c, err.Error())
		return
	}

	// If admin-only authentication was requested, enforce staff/superuser status
	if req.AdminOnly || c.GetHeader("X-Portal-Type") == "admin" {
		if !result.User.IsStaff && !result.User.IsSuperuser {
			blocker.RecordFailure(clientIP, req.Email)
			logger.Warn(fmt.Sprintf("[SECURITY] Standard user %s attempted unauthorized admin portal login from IP %s", req.Email, clientIP))
			c.JSON(http.StatusForbidden, gin.H{
 "code": "ADMIN_ACCESS_REQUIRED",
 "error": "Forbidden",
 "message": "Access denied: You do not have permission to access the administrator portal.",
			})
			return
		}
	}

	// Login succeeded - clear failure counter
	blocker.RecordSuccess(clientIP, req.Email)

	// Returns SimpleJWT compatible access/refresh + user & profile metadata
	response.JSON(c, http.StatusOK, result)
}

// AdminLogin specifically authenticates administrator accounts for the Admin Portal.
// Standard users without is_staff or is_superuser permission are rejected with 403 Forbidden.
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Email and password are required")
		return
	}

	clientIP := c.ClientIP()
	blocker := middleware.GetDefaultLoginBlocker()

	// 1. Check Brute-force lockout
	if locked, remaining, reason := blocker.IsLocked(clientIP, req.Email); locked {
		middleware.LogSecurityBruteForce(c, clientIP, req.Email, fmt.Sprintf("Blocked admin login attempt during active lockout (%s remaining)", remaining))
		middleware.RespondBruteForceLocked(c, remaining, reason)
		return
	}

	// 2. Authenticate credentials
	result, err := h.authUC.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if strings.HasPrefix(err.Error(), "ACCOUNT_SUSPENDED") {
			reason := strings.TrimPrefix(err.Error(), "ACCOUNT_SUSPENDED: ")
			c.JSON(http.StatusForbidden, gin.H{
 "code": "ACCOUNT_SUSPENDED",
 "error": "Account suspended",
 "message": "Your administrator account has been suspended.",
 "suspension_reason": reason,
 "support_email": "support@neighborservice.com",
			})
			return
		}

		newlyLocked, rem := blocker.RecordFailure(clientIP, req.Email)
		if newlyLocked {
			middleware.LogSecurityBruteForce(c, clientIP, req.Email, "Exceeded 5 failed admin login attempts; IP/account locked for 15 minutes")
			middleware.RespondBruteForceLocked(c, rem, "attempts_exceeded")
			return
		}

		response.Unauthorized(c, "Invalid administrator credentials.")
		return
	}

	// 3. Strictly verify that the user is an Administrator (Staff or Superuser)
	if !result.User.IsStaff && !result.User.IsSuperuser {
		blocker.RecordFailure(clientIP, req.Email)
		logger.Warn(fmt.Sprintf("[SECURITY] Standard user %s attempted unauthorized admin portal login from IP %s", req.Email, clientIP))

		if client := logger.GetProleadClient(); client != nil {
			traceID, _ := c.Get("trace_id")
			tID, _ := traceID.(string)
			client.Log(&logger.LogPayload{
				Level: "WARN",
				Message: fmt.Sprintf("[SECURITY ALERT] Unauthorized admin login attempt by %s from IP %s", req.Email, clientIP),
				TraceID: tID,
				HTTP: &logger.HTTPMetadata{
					Method:     c.Request.Method,
					Path:       c.Request.URL.Path,
					StatusCode: http.StatusForbidden,
					ClientIP:   clientIP,
					UserAgent:  c.Request.UserAgent(),
				},
				Context: map[string]interface{}{
 "security_alert": "unauthorized_admin_portal_access",
 "client_ip":      clientIP,
 "user_email":     logger.MaskPII(req.Email),
 "user_id":        result.User.ID.String(),
				},
			})
		}

		c.JSON(http.StatusForbidden, gin.H{
 "code": "ADMIN_ACCESS_REQUIRED",
 "error": "Forbidden",
 "message": "Access denied: You do not have permission to access the administrator portal.",
		})
		return
	}

	// 4. Success: clear brute force counter and respond with auth tokens
	blocker.RecordSuccess(clientIP, req.Email)
	logger.Info(fmt.Sprintf("[ADMIN AUTH] Administrator %s authenticated successfully from IP %s", req.Email, clientIP))

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
		if strings.HasPrefix(err.Error(), "ACCOUNT_SUSPENDED") {
			reason := strings.TrimPrefix(err.Error(), "ACCOUNT_SUSPENDED: ")
			c.JSON(http.StatusForbidden, gin.H{
 "code": "ACCOUNT_SUSPENDED",
 "error": "Account suspended",
 "message": "Your account has been suspended by an administrator.",
 "suspension_reason": reason,
 "support_email": "support@neighborservice.com",
			})
			return
		}
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
		if strings.HasPrefix(err.Error(), "ACCOUNT_SUSPENDED") {
			reason := strings.TrimPrefix(err.Error(), "ACCOUNT_SUSPENDED: ")
			c.JSON(http.StatusForbidden, gin.H{
 "code": "ACCOUNT_SUSPENDED",
 "error": "Account suspended",
 "message": "Your account has been suspended by an administrator.",
 "suspension_reason": reason,
 "support_email": "support@neighborservice.com",
			})
			return
		}
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

	email := req.Email
	if email == "" && token != "" {
		// Extract email from Apple id_token JWT claims if not provided in payload
		parser := jwt.NewParser()
		var claims jwt.MapClaims
		if _, _, err := parser.ParseUnverified(token, &claims); err == nil {
			if em, ok := claims["email"].(string); ok && em != "" {
				email = em
			}
			if email == "" {
				if sub, ok := claims["sub"].(string); ok && sub != "" {
					email = fmt.Sprintf("apple_%s@privaterelay.appleid.com", sub)
				}
			}
		}
	}

	name := req.Name
	if name == "" {
		name = strings.TrimSpace(req.FirstName + " " + req.LastName)
	}

	result, err := h.authUC.SocialLogin(c.Request.Context(), "apple", token, email, name)
	if err != nil {
		if strings.HasPrefix(err.Error(), "ACCOUNT_SUSPENDED") {
			reason := strings.TrimPrefix(err.Error(), "ACCOUNT_SUSPENDED: ")
			c.JSON(http.StatusForbidden, gin.H{
 "code": "ACCOUNT_SUSPENDED",
 "error": "Account suspended",
 "message": "Your account has been suspended by an administrator.",
 "suspension_reason": reason,
 "support_email": "support@neighborservice.com",
			})
			return
		}
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

