package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	"backend-go/pkg/auth"
	"backend-go/pkg/response"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CORS(cfg *config.Config) gin.HandlerFunc {
	// Parse allowed origins from env or default to local/standard production domains
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsEnv != "" {
		for _, o := range strings.Split(allowedOriginsEnv, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	} else {
		allowedOrigins = []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8080",
			"https://admin.neighborservice.com",
			"https://app.neighborservice.com",
			"https://neighborservice.com",
			"https://admin.neighborservices.com",
			"https://app.neighborservices.com",
			"https://neighborservices.com",

			"https://www.neighborservice.com",

			"https://api.neighborservice.com",
		}
	}

	corsConfig := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowOriginFunc: func(origin string) bool {
			for _, o := range allowedOrigins {
				if o == origin || o == "*" {
					return true
				}
				// Allow subdomains or wildcard domain match
				if strings.HasPrefix(o, "*.") && strings.HasSuffix(origin, o[1:]) {
					return true
				}
			}
			// Automatically permit any secure HTTPS subdomains of neighborservice.com
			if strings.HasSuffix(origin, ".neighborservice.com") || origin == "https://neighborservice.com" ||
				strings.HasSuffix(origin, ".neighborservices.com") || origin == "https://neighborservices.com" {
				return true
			}
			// In development mode without ALLOWED_ORIGINS, permit localhost with any port
			if os.Getenv("ENV") == "development" || os.Getenv("ENVIRONMENT") == "development" || allowedOriginsEnv == "" {
				if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
					return true
				}
			}
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Admin-Key"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition", "Retry-After"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	return cors.New(corsConfig)
}

// SecurityHeaders injects standard enterprise HTTP security headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data: https: blob:; script-src 'self' 'unsafe-inline' https://unpkg.com; style-src 'self' 'unsafe-inline' https://unpkg.com; font-src 'self' https: data:; connect-src 'self' ws: wss: https:;")
		c.Next()
	}
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if raw != "" {
			path = path + "?" + raw
		}

		fmt.Printf("[GIN] %v | %3d | %13v | %15s | %-7s %#v\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}

func AuthRequired(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || (parts[0] != "Bearer" && parts[0] != "JWT") {
			response.Unauthorized(c, "Invalid Authorization header format. Format: Bearer <token>")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := auth.ValidateToken(tokenString, cfg.JWTSecret)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("userType", claims.UserType)
		c.Set("claims", claims)

		c.Next()
	}
}

// AuthOptional extracts JWT claims if an Authorization header is present and valid,
// but does NOT reject the request if no token or an invalid token is provided.
// Handlers can check c.Get("userID") to determine if a user is authenticated.
func AuthOptional(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || (parts[0] != "Bearer" && parts[0] != "JWT") {
			c.Next()
			return
		}

		claims, err := auth.ValidateToken(parts[1], cfg.JWTSecret)
		if err != nil {
			c.Next()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("userType", claims.UserType)
		c.Set("claims", claims)

		c.Next()
	}
}

func ProviderOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userType, exists := c.Get("userType")
		if !exists || userType != "PROVIDER" {
			response.Forbidden(c, "Access restricted to service providers only")
			c.Abort()
			return
		}
		c.Next()
	}
}

func AdminRequired(cfg *config.Config, userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("userID")
		if !exists {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				response.Unauthorized(c, "Authorization header is required")
				c.Abort()
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || (parts[0] != "Bearer" && parts[0] != "JWT") {
				response.Unauthorized(c, "Invalid Authorization header format. Format: Bearer <token>")
				c.Abort()
				return
			}

			claims, err := auth.ValidateToken(parts[1], cfg.JWTSecret)
			if err != nil {
				response.Unauthorized(c, "Invalid or expired token")
				c.Abort()
				return
			}

			c.Set("userID", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("userType", claims.UserType)
			c.Set("claims", claims)
			userIDVal = claims.UserID
		}

		uidStr, ok := userIDVal.(string)
		if !ok {
			response.Unauthorized(c, "Invalid user credentials")
			c.Abort()
			return
		}
		uid, err := uuid.Parse(uidStr)
		if err != nil {
			response.Unauthorized(c, "Invalid user ID")
			c.Abort()
			return
		}

		user, err := userRepo.GetByID(c.Request.Context(), uid)
		if err != nil || user == nil {
			response.Unauthorized(c, "User not found")
			c.Abort()
			return
		}

		if !user.IsActive {
			response.Forbidden(c, "Account is disabled")
			c.Abort()
			return
		}

		if !user.IsStaff && !user.IsSuperuser {
			response.Forbidden(c, "Admin access required. User must be staff or superuser.")
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("isStaff", user.IsStaff)
		c.Set("isSuperuser", user.IsSuperuser)
		c.Next()
	}
}

func SuperuserOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		isSuperuser, exists := c.Get("isSuperuser")
		if !exists || isSuperuser != true {
			response.Forbidden(c, "Superuser authorization required for this action")
			c.Abort()
			return
		}
		c.Next()
	}
}

func AdminIPWhitelist() gin.HandlerFunc {
	allowedStr := os.Getenv("ADMIN_ALLOWED_IPS")
	if allowedStr == "" {
		return func(c *gin.Context) { c.Next() }
	}

	allowedIPs := strings.Split(allowedStr, ",")
	for i := range allowedIPs {
		allowedIPs[i] = strings.TrimSpace(allowedIPs[i])
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		allowed := false
		for _, ip := range allowedIPs {
			if ip == "" {
				continue
			}
			if ip == clientIP {
				allowed = true
				break
			}
			// Check CIDR block
			if strings.Contains(ip, "/") {
				if _, ipNet, err := net.ParseCIDR(ip); err == nil {
					if parsedClientIP := net.ParseIP(clientIP); parsedClientIP != nil && ipNet.Contains(parsedClientIP) {
						allowed = true
						break
					}
				}
			}
		}

		if !allowed {
			response.Forbidden(c, "Access denied: IP not whitelisted for admin panel")
			c.Abort()
			return
		}
		c.Next()
	}
}

func ErrorRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		errStr := fmt.Sprintf("Internal Server Error: %v", recovered)
		response.InternalError(c, errStr)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// AdminAuditLogger automatically persists mutating administrative requests to the audit trail.
func AdminAuditLogger(adminRepo repository.AdminRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		if status >= 400 {
			return
		}

		var adminID *uuid.UUID
		if uVal, exists := c.Get("user"); exists {
			if u, ok := uVal.(*entity.User); ok && u != nil {
				adminID = &u.ID
			}
		}

		if adminID == nil {
			if uidVal, exists := c.Get("userID"); exists {
				if uidStr, ok := uidVal.(string); ok {
					if parsed, err := uuid.Parse(uidStr); err == nil {
						adminID = &parsed
					}
				}
			}
		}

		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		durationMs := time.Since(start).Milliseconds()

		go func(aid *uuid.UUID, m, p, ip, ua string, dur int64, statusCode int) {
			details := entity.JSONMap{
				"method":      m,
				"path":        p,
				"status_code": statusCode,
				"duration_ms": dur,
				"user_agent":  ua,
			}
			logEntry := entity.AuditLog{
				ID:           uuid.New(),
				UserID:       aid,
				Action:       fmt.Sprintf("ADMIN_%s", m),
				ResourceType: "ADMIN_API",
				ResourceID:   p,
				Details:      details,
				IPAddress:    ip,
				CreatedAt:    time.Now(),
			}
			_ = adminRepo.CreateAuditLog(c.Request.Context(), &logEntry)
		}(adminID, method, path, clientIP, userAgent, durationMs, status)
	}
}

// UserAuditLogger automatically captures user actions and security events to the audit trail.
func UserAuditLogger(auditRepo repository.AuditLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path

		c.Next()

		status := c.Writer.Status()
		if status >= 400 {
			return
		}

		var action, resourceType, resourceID string
		switch {
		case strings.Contains(path, "/auth/login") || strings.Contains(path, "/auth/google") || strings.Contains(path, "/auth/apple"):
			action = "LOGIN"
			resourceType = "AUTH"
			resourceID = "SESSION"
		case strings.Contains(path, "/auth/register"):
			action = "REGISTER"
			resourceType = "AUTH"
			resourceID = "USER"
		case strings.Contains(path, "/auth/logout"):
			action = "LOGOUT"
			resourceType = "AUTH"
			resourceID = "SESSION"
		case (strings.Contains(path, "/profile") || strings.Contains(path, "/accounts/profile")) && (method == "POST" || method == "PUT" || method == "PATCH"):
			action = "UPDATE_PROFILE"
			resourceType = "PROFILE"
			resourceID = "PROFILE_DATA"
		case strings.Contains(path, "/requests/proposals") || strings.Contains(path, "/proposals"):
			if method == "POST" {
				action = "ACCEPT_REQUEST"
				resourceType = "PROPOSAL"
			}
		case strings.Contains(path, "/requests") && method == "POST":
			action = "CREATE_REQUEST"
			resourceType = "SERVICE_REQUEST"
		case strings.Contains(path, "/appointments"):
			if method == "POST" || method == "PUT" {
				action = "APPROVE_REQUEST"
				resourceType = "APPOINTMENT"
			}
		case strings.Contains(path, "/reviews") && method == "POST":
			action = "CREATE_REVIEW"
			resourceType = "REVIEW"
		case strings.Contains(path, "/disputes") && method == "POST":
			action = "RAISE_DISPUTE"
			resourceType = "DISPUTE"
		case strings.Contains(path, "/favorites") && method == "POST":
			action = "ADD_FAVORITE"
			resourceType = "FAVORITE"
		case strings.Contains(path, "/payouts") && method == "POST":
			action = "REQUEST_PAYOUT"
			resourceType = "PAYOUT"
		case strings.Contains(path, "/subscriptions") && method == "POST":
			action = "CREATE_SUBSCRIPTION"
			resourceType = "SUBSCRIPTION"
		case strings.Contains(path, "/moderation/background-checks") && method == "POST":
			action = "FUND_BACKGROUND_CHECK"
			resourceType = "BACKGROUND_CHECK"
		}

		if action == "" {
			return
		}

		var targetUserID *uuid.UUID
		if uidVal, exists := c.Get("userID"); exists {
			if uidStr, ok := uidVal.(string); ok && uidStr != "" {
				if parsed, err := uuid.Parse(uidStr); err == nil {
					targetUserID = &parsed
				}
			}
		}

		if targetUserID == nil {
			if uVal, exists := c.Get("user"); exists {
				if u, ok := uVal.(*entity.User); ok && u != nil {
					targetUserID = &u.ID
				}
			}
		}

		if targetUserID == nil {
			return
		}

		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		go func(uid *uuid.UUID, act, rType, rID, ip, ua, p, m string, sc int) {
			details := entity.JSONMap{
				"path":        p,
				"method":      m,
				"status_code": sc,
				"user_agent":  ua,
			}
			logEntry := entity.AuditLog{
				ID:           uuid.New(),
				UserID:       uid,
				Action:       act,
				ResourceType: rType,
				ResourceID:   rID,
				Details:      details,
				IPAddress:    ip,
				CreatedAt:    time.Now(),
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = auditRepo.Create(ctx, &logEntry)
		}(targetUserID, action, resourceType, resourceID, clientIP, userAgent, path, method, status)
	}
}


