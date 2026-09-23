package handler

import (
	"net/http"

	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

type DocsHandler struct{}

func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// GetOpenAPISpec returns the comprehensive OpenAPI 3.0.3 specification for ALL endpoints across the Go backend.
func (h *DocsHandler) GetOpenAPISpec(c *gin.Context) {
	spec := gin.H{
		"openapi": "3.0.3",
		"info": gin.H{
			"title":       "Neighbor Services API - Complete OpenAPI Specification",
			"version":     "2.0.0",
			"description": "Complete, interactive Swagger documentation for the entire Neighbor Services platform. Includes all Seeker, Provider, Payments, Subscriptions, Moderation, WebSockets, and Superuser Admin endpoints.",
			"contact": gin.H{
				"name":  "Neighbor Services Platform Team",
				"email": "support@neighborservice.com",
			},
		},
		"servers": []gin.H{
			{"url": "/", "description": "Current Host / Development Server"},
		},
		"tags": []gin.H{
			{"name": "Auth & Accounts", "description": "Registration, authentication, OTP verification, password resets, and OAuth login"},
			{"name": "Profiles & About", "description": "User profile management, bio, hourly rate, and portfolio assets"},
			{"name": "Service Packages", "description": "Provider catalog service packages and offerings"},
			{"name": "Categories & Services", "description": "Service categories, catalog lookup, and service discovery"},
			{"name": "Service Requests & Proposals", "description": "Seeker requests and provider proposals"},
			{"name": "Appointments & Reviews", "description": "Appointment scheduling, arrival verification, completion, and ratings"},
			{"name": "Disputes", "description": "Appointment dispute filing and evidence uploads"},
			{"name": "Chat & Messaging", "description": "Conversations, real-time messaging, and chat blocking"},
			{"name": "Notifications", "description": "In-app notifications and push notification device tokens"},
			{"name": "Consultations (Agora RTC/RTM)", "description": "Video/audio consultation RTC and RTM token generation"},
			{"name": "Payments & Wallet", "description": "Stripe Connect onboarding, wallet balance, payouts, and payment sheets"},
			{"name": "In-App Purchases & Subscriptions", "description": "Apple App Store and Google Play subscription validation"},
			{"name": "Moderation & Background Checks", "description": "Identity verification, Checkr background checks, and user reporting"},
			{"name": "Admin: Dashboard & Users", "description": "Administrative KPI dashboard, user search, updates, restore, and GDPR exports"},
			{"name": "Admin: Verifications & Background Checks", "description": "Admin ID verification review, batch processing, and Checkr overrides"},
			{"name": "Admin: Moderation, Reports & Disputes", "description": "Admin dispute resolution, report adjudication, and notes"},
			{"name": "Admin: Financials, Payouts & Wallets", "description": "Payout approvals/rejections, wallet adjustments, and financial reports"},
			{"name": "Admin: Categories, Services & Settings", "description": "Catalog management, moderation settings, and audit logs"},
			{"name": "Admin: Operations & Feature Flags", "description": "Broadcast notifications, runtime feature flags, and cache clearing"},
			{"name": "Admin: Enterprise Security & 2FA", "description": "Admin TOTP 2FA, staff impersonation, and RBAC roles"},
			{"name": "Admin: Fraud & Risk Scoring", "description": "Automated fraud scoring, risk detection, and alert resolution"},
			{"name": "Admin: Templates & System Backups", "description": "Notification templates, database backups, and health telemetry"},
			{"name": "Admin: Data Exports (CSV)", "description": "Streaming CSV exports for users, payouts, and disputes"},
			{"name": "WebSockets", "description": "Real-time presence, chat, tracking, notifications, and admin events"},
		},
		"components": gin.H{
			"securitySchemes": gin.H{
				"bearerAuth": gin.H{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
					"description":  "Enter your JWT Access Token (without 'Bearer ' prefix).",
				},
			},
			"schemas": gin.H{
				"StandardResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"status":  gin.H{"type": "string"},
						"message": gin.H{"type": "string"},
					},
				},
				"ErrorResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"error":   gin.H{"type": "string"},
						"message": gin.H{"type": "string"},
					},
				},
			},
		},
		"security": []gin.H{
			{"bearerAuth": []string{}},
		},
		"paths": h.buildPaths(),
	}

	response.JSON(c, http.StatusOK, spec)
}

// GetSwaggerUI returns the embedded, interactive Swagger UI HTML page.
func (h *DocsHandler) GetSwaggerUI(c *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Neighbor Services - Interactive Swagger API Documentation</title>
	<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css">
	<style>
		html { box-sizing: border-box; overflow-y: scroll; }
		*, *:before, *:after { box-sizing: inherit; }
		body { margin: 0; background: #f8fafc; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; }
		.topbar { display: none !important; }
		.swagger-ui .info { margin: 25px 0; }
		.swagger-ui .info .title { font-size: 32px; color: #0f172a; font-weight: 700; }
		.swagger-ui .scheme-container { background: #ffffff; padding: 16px; border-radius: 8px; margin-bottom: 24px; box-shadow: 0 1px 3px rgba(0,0,0,0.08); }
		.swagger-ui .opblock-tag { font-size: 18px; border-bottom: 2px solid #e2e8f0; }
	</style>
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
	<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js"></script>
	<script>
		window.onload = function() {
			window.ui = SwaggerUIBundle({
				url: "/openapi.json",
				dom_id: '#swagger-ui',
				deepLinking: true,
				presets: [
					SwaggerUIBundle.presets.apis,
					SwaggerUIStandalonePreset
				],
				plugins: [
					SwaggerUIBundle.plugins.DownloadUrl
				],
				layout: "StandaloneLayout",
				persistAuthorization: true,
				displayRequestDuration: true,
				docExpansion: "none",
				filter: true,
				showExtensions: true,
				showCommonExtensions: true,
				tryItOutEnabled: true
			});
		}
	</script>
</body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func (h *DocsHandler) buildPaths() gin.H {
	return gin.H{
		// ─── HEALTH & SYSTEM ────────────────────────────────────────────────
		"/health": gin.H{
			"get": gin.H{
				"tags":        []string{"Auth & Accounts"},
				"summary":     "System Health Check",
				"description": "Returns operational status of the Go service.",
				"responses":   gin.H{"200": gin.H{"description": "Service is healthy."}},
			},
		},
		"/api/health": gin.H{
			"get": gin.H{
				"tags":        []string{"Auth & Accounts"},
				"summary":     "API Health Check",
				"responses":   gin.H{"200": gin.H{"description": "API Gateway is operational."}},
			},
		},

		// ─── WEBHOOKS & CALLBACKS ───────────────────────────────────────────
		"/callbacks/apple": gin.H{
			"post": gin.H{
				"tags":        []string{"Auth & Accounts"},
				"summary":     "Apple OAuth Callback",
				"description": "Receives Apple OAuth POST callback data.",
				"responses":   gin.H{"200": gin.H{"description": "Callback handled."}},
			},
		},
		"/payments/webhook/apple-s2s/": gin.H{
			"post": gin.H{
				"tags":        []string{"In-App Purchases & Subscriptions"},
				"summary":     "Apple Server-to-Server (S2S) V2 Webhook",
				"description": "Receives subscription renewal, cancellation, and refund notifications from Apple StoreKit.",
				"responses":   gin.H{"200": gin.H{"description": "Webhook received."}},
			},
		},
		"/api/v1/payments/webhook/": gin.H{
			"post": gin.H{
				"tags":        []string{"Payments & Wallet"},
				"summary":     "Stripe Webhook Listener",
				"description": "Receives and validates asynchronous Stripe events (payment_intent.succeeded, transfer.created, account.updated).",
				"responses":   gin.H{"200": gin.H{"description": "Stripe event processed."}},
			},
		},
		"/api/v1/payments/webhook/google-pubsub/": gin.H{
			"post": gin.H{
				"tags":        []string{"In-App Purchases & Subscriptions"},
				"summary":     "Google Play Real-Time Developer Notifications (Pub/Sub)",
				"description": "Receives Google Cloud Pub/Sub push messages for subscription lifecycle events.",
				"responses":   gin.H{"200": gin.H{"description": "Pub/Sub acknowledged."}},
			},
		},
		"/api/v1/moderation/checkr-webhook/": gin.H{
			"post": gin.H{
				"tags":        []string{"Moderation & Background Checks"},
				"summary":     "Checkr Background Check Webhook",
				"description": "Receives background check report completion and status updates from Checkr.",
				"responses":   gin.H{"200": gin.H{"description": "Checkr update processed."}},
			},
		},

		// ─── AUTH & ACCOUNTS ────────────────────────────────────────────────
		"/api/v1/accounts/register/": gin.H{
			"post": gin.H{
				"tags":        []string{"Auth & Accounts"},
				"summary":     "User Registration",
				"description": "Registers a new user (Seeker or Provider) and sends an OTP verification code.",
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"email", "password"},
								"properties": gin.H{
									"email":      gin.H{"type": "string", "example": "user@example.com"},
									"password":   gin.H{"type": "string", "example": "SecurePass123!"},
									"user_type":  gin.H{"type": "string", "enum": []string{"SEEKER", "PROVIDER"}, "description": "Optional: defaults to SEEKER"},
									"first_name": gin.H{"type": "string", "example": "Alex", "description": "Optional: can be added later in profile"},
									"last_name":  gin.H{"type": "string", "example": "Smith", "description": "Optional: can be added later in profile"},
									"phone":      gin.H{"type": "string", "example": "+15551234567", "description": "Optional: can be added later in profile"},
								},
							},
						},
					},
				},
				"responses": gin.H{
					"201": gin.H{"description": "User created. OTP code dispatched to email."},
					"400": gin.H{"description": "Validation error or email already in use."},
				},
			},
		},
		"/api/v1/accounts/login/": gin.H{
			"post": gin.H{
				"tags":        []string{"Auth & Accounts"},
				"summary":     "User Login",
				"description": "Authenticates user and returns JWT access and refresh tokens.",
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"email", "password"},
								"properties": gin.H{
									"email":    gin.H{"type": "string", "example": "user@example.com"},
									"password": gin.H{"type": "string", "example": "SecurePass123!"},
								},
							},
						},
					},
				},
				"responses": gin.H{
					"200": gin.H{"description": "Login successful. JWT token pair returned."},
					"401": gin.H{"description": "Invalid email or password."},
				},
			},
		},
		"/api/v1/accounts/token/refresh/": gin.H{
			"post": gin.H{
				"tags":        []string{"Auth & Accounts"},
				"summary":     "Refresh Access Token",
				"description": "Generates a new access token using a valid refresh token.",
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"refresh"},
								"properties": gin.H{
									"refresh": gin.H{"type": "string", "example": "eyJhbGciOiJIUzI1Ni..."},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "New access token returned."}},
			},
		},
		"/api/v1/accounts/token/rotate/": gin.H{
			"post": gin.H{
				"tags":        []string{"Auth & Accounts"},
				"summary":     "JWT Refresh Token Rotation",
				"description": "Invalidates previous refresh token and returns a new access + refresh token pair.",
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"refresh"},
								"properties": gin.H{
									"refresh": gin.H{"type": "string", "example": "eyJhbGciOiJIUzI1Ni..."},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Rotated token pair returned."}},
			},
		},
		"/api/v1/accounts/export/": gin.H{
			"get": gin.H{
				"tags":        []string{"Auth & Accounts"},
				"summary":     "GDPR User Data Export",
				"description": "Exports complete personal profile, wallet history, bookings, and disputes for the authenticated user.",
				"security":    []gin.H{{"bearerAuth": []string{}}},
				"responses":   gin.H{"200": gin.H{"description": "Full user data structure returned."}},
			},
		},
		"/api/v1/accounts/verify-otp/": gin.H{
			"post": gin.H{
				"tags":        []string{"Auth & Accounts"},
				"summary":     "Verify Account OTP",
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"email", "otp"},
								"properties": gin.H{
									"email": gin.H{"type": "string", "example": "user@example.com"},
									"otp":   gin.H{"type": "string", "example": "123456"},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Email verified."}},
			},
		},
		"/api/v1/accounts/resend-otp/": gin.H{
			"post": gin.H{
				"tags":        []string{"Auth & Accounts"},
				"summary":     "Resend OTP Code",
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"email"},
								"properties": gin.H{"email": gin.H{"type": "string", "example": "user@example.com"}},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "New OTP dispatched."}},
			},
		},
		"/api/v1/accounts/change-password/": gin.H{
			"post": gin.H{
				"tags":     []string{"Auth & Accounts"},
				"summary":  "Change Password",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"old_password", "new_password"},
								"properties": gin.H{
									"old_password": gin.H{"type": "string"},
									"new_password": gin.H{"type": "string"},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Password updated."}},
			},
		},
		"/api/v1/accounts/password-reset/": gin.H{
			"post": gin.H{
				"tags":        []string{"Auth & Accounts"},
				"summary":     "Request Password Reset",
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"email"},
								"properties": gin.H{"email": gin.H{"type": "string", "example": "user@example.com"}},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Password reset code sent."}},
			},
		},
		"/api/v1/accounts/password-reset-confirm/": gin.H{
			"post": gin.H{
				"tags":        []string{"Auth & Accounts"},
				"summary":     "Confirm Password Reset with OTP",
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"email", "otp", "new_password"},
								"properties": gin.H{
									"email":        gin.H{"type": "string"},
									"otp":          gin.H{"type": "string"},
									"new_password": gin.H{"type": "string"},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Password reset successful."}},
			},
		},
		"/api/v1/accounts/login-google/": gin.H{
			"post": gin.H{
				"tags":        []string{"Auth & Accounts"},
				"summary":     "Google OAuth Sign-In",
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"id_token"},
								"properties": gin.H{
									"id_token":  gin.H{"type": "string"},
									"user_type": gin.H{"type": "string", "enum": []string{"SEEKER", "PROVIDER"}},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Authentication tokens returned."}},
			},
		},
		"/api/v1/accounts/login-apple/": gin.H{
			"post": gin.H{
				"tags":        []string{"Auth & Accounts"},
				"summary":     "Apple OAuth Sign-In",
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"id_token"},
								"properties": gin.H{
									"id_token":  gin.H{"type": "string"},
									"user_type": gin.H{"type": "string"},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Authentication tokens returned."}},
			},
		},
		"/api/v1/accounts/logout/": gin.H{
			"post": gin.H{
				"tags":      []string{"Auth & Accounts"},
				"summary":   "User Logout",
				"security":  []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Logged out."}},
			},
		},
		"/api/v1/accounts/delete-account/": gin.H{
			"post": gin.H{
				"tags":      []string{"Auth & Accounts"},
				"summary":   "Delete User Account",
				"security":  []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Account deactivated/deleted."}},
			},
		},
		"/api/v1/accounts/legal/": gin.H{
			"get": gin.H{
				"tags":      []string{"Auth & Accounts"},
				"summary":   "Get Terms of Service & Privacy Policy",
				"responses": gin.H{"200": gin.H{"description": "Legal documents payload."}},
			},
		},

		// ─── PROFILES & ABOUT ───────────────────────────────────────────────
		"/api/v1/accounts/profile/me/": gin.H{
			"get": gin.H{
				"tags":      []string{"Profiles & About"},
				"summary":   "Get Current User Profile",
				"security":  []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Profile details."}},
			},
		},
		"/api/v1/accounts/profile/popular/": gin.H{
			"get": gin.H{
				"tags":      []string{"Profiles & About"},
				"summary":   "Get Popular Top-Rated Providers",
				"responses": gin.H{"200": gin.H{"description": "List of popular provider profiles."}},
			},
		},
		"/api/v1/accounts/profile/update_me/": gin.H{
			"patch": gin.H{
				"tags":     []string{"Profiles & About"},
				"summary":  "Update Current Profile",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type": "object",
								"properties": gin.H{
									"first_name":  gin.H{"type": "string"},
									"last_name":   gin.H{"type": "string"},
									"bio":         gin.H{"type": "string"},
									"hourly_rate": gin.H{"type": "number"},
									"address":     gin.H{"type": "string"},
									"city":        gin.H{"type": "string"},
									"state":       gin.H{"type": "string"},
									"zip_code":    gin.H{"type": "string"},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Updated profile object."}},
			},
		},
		"/api/v1/accounts/profile/picture/": gin.H{
			"patch": gin.H{
				"tags":     []string{"Profiles & About"},
				"summary":  "Upload Profile Picture (Magic-byte verified)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"content": gin.H{
						"multipart/form-data": gin.H{
							"schema": gin.H{
								"type": "object",
								"properties": gin.H{
									"profile_picture": gin.H{"type": "string", "format": "binary"},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Profile picture updated."}},
			},
		},
		"/api/v1/accounts/profile/{id}/": gin.H{
			"get": gin.H{
				"tags":     []string{"Profiles & About"},
				"summary":  "Get Profile by ID",
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Profile object."}},
			},
		},
		"/api/v1/accounts/about/": gin.H{
			"get": gin.H{
				"tags":     []string{"Profiles & About"},
				"summary":  "Get About Information",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "About section details."}},
			},
			"post": gin.H{
				"tags":     []string{"Profiles & About"},
				"summary":  "Create / Update About Section",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type": "object",
								"properties": gin.H{
									"description": gin.H{"type": "string"},
									"skills":      gin.H{"type": "array", "items": gin.H{"type": "string"}},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "About updated."}},
			},
		},
		"/api/v1/accounts/portfolio/": gin.H{
			"get": gin.H{
				"tags":     []string{"Profiles & About"},
				"summary":  "List Portfolio Items",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "List of portfolio items."}},
			},
			"post": gin.H{
				"tags":     []string{"Profiles & About"},
				"summary":  "Create Portfolio Item with Image",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"content": gin.H{
						"multipart/form-data": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"title", "image"},
								"properties": gin.H{
									"title":       gin.H{"type": "string"},
									"description": gin.H{"type": "string"},
									"image":       gin.H{"type": "string", "format": "binary"},
								},
							},
						},
					},
				},
				"responses": gin.H{"201": gin.H{"description": "Portfolio item created."}},
			},
		},
		"/api/v1/accounts/portfolio/{id}/": gin.H{
			"delete": gin.H{
				"tags":     []string{"Profiles & About"},
				"summary":  "Delete Portfolio Item",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Deleted."}},
			},
		},

		// ─── SERVICE PACKAGES ───────────────────────────────────────────────
		"/api/v1/accounts/service-packages/": gin.H{
			"get": gin.H{
				"tags":     []string{"Service Packages"},
				"summary":  "List Provider Service Packages",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Service packages list."}},
			},
			"post": gin.H{
				"tags":     []string{"Service Packages"},
				"summary":  "Create Service Package",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"title", "price", "duration_minutes"},
								"properties": gin.H{
									"title":            gin.H{"type": "string"},
									"description":      gin.H{"type": "string"},
									"price":            gin.H{"type": "number"},
									"duration_minutes": gin.H{"type": "integer"},
								},
							},
						},
					},
				},
				"responses": gin.H{"201": gin.H{"description": "Package created."}},
			},
		},
		"/api/v1/accounts/service-packages/{id}/": gin.H{
			"delete": gin.H{
				"tags":     []string{"Service Packages"},
				"summary":  "Delete Service Package",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Package deleted."}},
			},
		},

		// ─── CATEGORIES & CATALOG SERVICES ──────────────────────────────────
		"/api/v1/services/categories/": gin.H{
			"get": gin.H{
				"tags":      []string{"Categories & Services"},
				"summary":   "List Service Categories",
				"responses": gin.H{"200": gin.H{"description": "Categories list."}},
			},
		},
		"/api/v1/services/catalog-services/": gin.H{
			"get": gin.H{
				"tags":    []string{"Categories & Services"},
				"summary": "Search Catalog Services",
				"parameters": []gin.H{
					{"name": "category", "in": "query", "schema": gin.H{"type": "string"}},
					{"name": "search", "in": "query", "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Catalog services list."}},
			},
		},
		"/api/v1/services/match-providers/": gin.H{
			"post": gin.H{
				"tags":    []string{"Categories & Services"},
				"summary": "Match Providers by Location & Category",
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"category_id", "latitude", "longitude"},
								"properties": gin.H{
									"category_id": gin.H{"type": "string"},
									"latitude":    gin.H{"type": "number"},
									"longitude":   gin.H{"type": "number"},
									"radius_km":   gin.H{"type": "number", "default": 25},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Matched providers list."}},
			},
		},

		// ─── SERVICE REQUESTS & PROPOSALS ───────────────────────────────────
		"/api/v1/services/requests/": gin.H{
			"get": gin.H{
				"tags":     []string{"Service Requests & Proposals"},
				"summary":  "List Service Requests",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Requests list."}},
			},
			"post": gin.H{
				"tags":     []string{"Service Requests & Proposals"},
				"summary":  "Create Service Request",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"category_id", "title", "description"},
								"properties": gin.H{
									"category_id": gin.H{"type": "string"},
									"title":       gin.H{"type": "string"},
									"description": gin.H{"type": "string"},
									"budget":      gin.H{"type": "number"},
									"latitude":    gin.H{"type": "number"},
									"longitude":   gin.H{"type": "number"},
								},
							},
						},
					},
				},
				"responses": gin.H{"201": gin.H{"description": "Request created."}},
			},
		},
		"/api/v1/services/requests/{id}/": gin.H{
			"get": gin.H{
				"tags":     []string{"Service Requests & Proposals"},
				"summary":  "Get Request by ID",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Request details."}},
			},
			"patch": gin.H{
				"tags":     []string{"Service Requests & Proposals"},
				"summary":  "Update Request",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Updated."}},
			},
			"delete": gin.H{
				"tags":     []string{"Service Requests & Proposals"},
				"summary":  "Cancel / Delete Request",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Deleted."}},
			},
		},
		"/api/v1/services/requests/{id}/approve_proposal/": gin.H{
			"post": gin.H{
				"tags":     []string{"Service Requests & Proposals"},
				"summary":  "Approve / Accept Provider Proposal",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"proposal_id"},
								"properties": gin.H{"proposal_id": gin.H{"type": "string"}},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Proposal accepted, appointment booked."}},
			},
		},
		"/api/v1/services/proposals/": gin.H{
			"get": gin.H{
				"tags":     []string{"Service Requests & Proposals"},
				"summary":  "List Proposals",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Proposals list."}},
			},
			"post": gin.H{
				"tags":     []string{"Service Requests & Proposals"},
				"summary":  "Submit Proposal to Request",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"service_request_id", "price", "cover_letter"},
								"properties": gin.H{
									"service_request_id": gin.H{"type": "string"},
									"price":              gin.H{"type": "number"},
									"cover_letter":       gin.H{"type": "string"},
								},
							},
						},
					},
				},
				"responses": gin.H{"201": gin.H{"description": "Proposal submitted."}},
			},
		},

		// ─── APPOINTMENTS & REVIEWS ─────────────────────────────────────────
		"/api/v1/interactions/appointments/": gin.H{
			"get": gin.H{
				"tags":     []string{"Appointments & Reviews"},
				"summary":  "List Appointments",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Appointments list."}},
			},
			"post": gin.H{
				"tags":     []string{"Appointments & Reviews"},
				"summary":  "Create Appointment",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"201": gin.H{"description": "Appointment scheduled."}},
			},
		},
		"/api/v1/interactions/appointments/{id}/verify-code/": gin.H{
			"post": gin.H{
				"tags":     []string{"Appointments & Reviews"},
				"summary":  "Verify Provider Arrival 4-Digit Code",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"code"},
								"properties": gin.H{"code": gin.H{"type": "string", "example": "4821"}},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Arrival verified."}},
			},
		},
		"/api/v1/interactions/appointments/{id}/complete/": gin.H{
			"post": gin.H{
				"tags":     []string{"Appointments & Reviews"},
				"summary":  "Complete Appointment & Release Payment",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Completed."}},
			},
		},
		"/api/v1/interactions/appointments/{id}/cancel/": gin.H{
			"post": gin.H{
				"tags":     []string{"Appointments & Reviews"},
				"summary":  "Cancel Appointment",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Cancelled."}},
			},
		},
		"/api/v1/interactions/favorites/": gin.H{
			"get": gin.H{
				"tags":     []string{"Appointments & Reviews"},
				"summary":  "List Favorite Providers",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Favorites list."}},
			},
			"post": gin.H{
				"tags":     []string{"Appointments & Reviews"},
				"summary":  "Toggle / Add Favorite Provider",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Favorited."}},
			},
		},
		"/api/v1/interactions/reviews/": gin.H{
			"get": gin.H{
				"tags":     []string{"Appointments & Reviews"},
				"summary":  "List Provider Reviews",
				"parameters": []gin.H{
					{"name": "provider_id", "in": "query", "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Reviews list."}},
			},
			"post": gin.H{
				"tags":     []string{"Appointments & Reviews"},
				"summary":  "Submit Rating & Review",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"appointment_id", "rating", "comment"},
								"properties": gin.H{
									"appointment_id": gin.H{"type": "string"},
									"rating":         gin.H{"type": "integer", "minimum": 1, "maximum": 5},
									"comment":        gin.H{"type": "string"},
								},
							},
						},
					},
				},
				"responses": gin.H{"201": gin.H{"description": "Review published."}},
			},
		},

		// ─── DISPUTES ───────────────────────────────────────────────────────
		"/api/v1/interactions/disputes/": gin.H{
			"get": gin.H{
				"tags":     []string{"Disputes"},
				"summary":  "List Disputes",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Disputes list."}},
			},
			"post": gin.H{
				"tags":     []string{"Disputes"},
				"summary":  "File Appointment Dispute",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"appointment_id", "reason"},
								"properties": gin.H{
									"appointment_id": gin.H{"type": "string"},
									"reason":         gin.H{"type": "string"},
									"description":    gin.H{"type": "string"},
								},
							},
						},
					},
				},
				"responses": gin.H{"201": gin.H{"description": "Dispute opened."}},
			},
		},
		"/api/v1/interactions/disputes/{id}/upload-evidence/": gin.H{
			"post": gin.H{
				"tags":     []string{"Disputes"},
				"summary":  "Upload Dispute Evidence Attachment",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"requestBody": gin.H{
					"content": gin.H{
						"multipart/form-data": gin.H{
							"schema": gin.H{
								"type": "object",
								"properties": gin.H{"file": gin.H{"type": "string", "format": "binary"}},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Evidence attached."}},
			},
		},

		// ─── CHAT & MESSAGING ───────────────────────────────────────────────
		"/api/v1/chat/conversations/": gin.H{
			"get": gin.H{
				"tags":     []string{"Chat & Messaging"},
				"summary":  "List Active Conversations",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Conversations list."}},
			},
			"post": gin.H{
				"tags":     []string{"Chat & Messaging"},
				"summary":  "Start Conversation",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"other_user_id"},
								"properties": gin.H{"other_user_id": gin.H{"type": "string"}},
							},
						},
					},
				},
				"responses": gin.H{"201": gin.H{"description": "Conversation created."}},
			},
		},
		"/api/v1/chat/conversations/block_chat/": gin.H{
			"post": gin.H{
				"tags":     []string{"Chat & Messaging"},
				"summary":  "Block User in Chat",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "User blocked."}},
			},
		},
		"/api/v1/chat/messages/": gin.H{
			"get": gin.H{
				"tags":     []string{"Chat & Messaging"},
				"summary":  "Get Conversation Messages",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "conversation_id", "in": "query", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Messages list."}},
			},
			"post": gin.H{
				"tags":     []string{"Chat & Messaging"},
				"summary":  "Send Chat Message",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"conversation_id", "content"},
								"properties": gin.H{
									"conversation_id": gin.H{"type": "string"},
									"content":         gin.H{"type": "string"},
								},
							},
						},
					},
				},
				"responses": gin.H{"201": gin.H{"description": "Message sent."}},
			},
		},

		// ─── NOTIFICATIONS ──────────────────────────────────────────────────
		"/api/v1/notifications/": gin.H{
			"get": gin.H{
				"tags":     []string{"Notifications"},
				"summary":  "Get Notification Inbox",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Notifications list."}},
			},
		},
		"/api/v1/notifications/{id}/read/": gin.H{
			"patch": gin.H{
				"tags":     []string{"Notifications"},
				"summary":  "Mark Notification as Read",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Marked as read."}},
			},
		},
		"/api/v1/notifications/mark_all_as_read/": gin.H{
			"post": gin.H{
				"tags":      []string{"Notifications"},
				"summary":   "Mark All Notifications as Read",
				"security":  []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "All read."}},
			},
		},
		"/api/v1/notifications/device-token/": gin.H{
			"post": gin.H{
				"tags":     []string{"Notifications"},
				"summary":  "Register APNS / FCM Push Device Token",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"token"},
								"properties": gin.H{
									"token":    gin.H{"type": "string"},
									"platform": gin.H{"type": "string", "enum": []string{"ios", "android", "web"}},
								},
							},
						},
					},
				},
				"responses": gin.H{"201": gin.H{"description": "Token registered."}},
			},
		},

		// ─── CONSULTATIONS (AGORA) ──────────────────────────────────────────
		"/api/v1/consultations/rtc-token/": gin.H{
			"get": gin.H{
				"tags":     []string{"Consultations (Agora RTC/RTM)"},
				"summary":  "Generate Agora Audio/Video RTC Token (GET)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "channel_name", "in": "query", "required": true, "schema": gin.H{"type": "string"}},
					{"name": "uid", "in": "query", "schema": gin.H{"type": "integer"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Agora RTC Token returned."}},
			},
			"post": gin.H{
				"tags":     []string{"Consultations (Agora RTC/RTM)"},
				"summary":  "Generate Agora Audio/Video RTC Token (POST)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"channel_name"},
								"properties": gin.H{
									"channel_name": gin.H{"type": "string"},
									"uid":          gin.H{"type": "integer"},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Agora RTC Token returned."}},
			},
		},
		"/api/v1/consultations/rtm-token/": gin.H{
			"get": gin.H{
				"tags":     []string{"Consultations (Agora RTC/RTM)"},
				"summary":  "Generate Agora Real-Time Messaging RTM Token (GET)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Agora RTM Token returned."}},
			},
			"post": gin.H{
				"tags":     []string{"Consultations (Agora RTC/RTM)"},
				"summary":  "Generate Agora Real-Time Messaging RTM Token (POST)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Agora RTM Token returned."}},
			},
		},

		// ─── PAYMENTS, WALLET & SUBSCRIPTIONS ───────────────────────────────
		"/api/v1/payments/subscription-plans/": gin.H{
			"get": gin.H{
				"tags":      []string{"In-App Purchases & Subscriptions"},
				"summary":   "List Available Subscription Plans",
				"responses": gin.H{"200": gin.H{"description": "Subscription plans list."}},
			},
		},
		"/api/v1/payments/validate/apple/": gin.H{
			"post": gin.H{
				"tags":     []string{"In-App Purchases & Subscriptions"},
				"summary":  "Validate Apple App Store Receipt",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Subscription activated."}},
			},
		},
		"/api/v1/payments/validate/google/": gin.H{
			"post": gin.H{
				"tags":     []string{"In-App Purchases & Subscriptions"},
				"summary":  "Validate Google Play Purchase Token",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Google Play subscription confirmed."}},
			},
		},
		"/api/v1/payments/wallet/my_wallet/": gin.H{
			"get": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Get Wallet Balance",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Balance and currency."}},
			},
		},
		"/api/v1/payments/wallet/transactions/": gin.H{
			"get": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "List Wallet Ledger Transactions",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Transactions history."}},
			},
		},
		"/api/v1/payments/wallet/request_payout/": gin.H{
			"post": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Request Payout to Bank Account",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"amount"},
								"properties": gin.H{"amount": gin.H{"type": "number", "example": 100.0}},
							},
						},
					},
				},
				"responses": gin.H{"201": gin.H{"description": "Payout request submitted."}},
			},
		},
		"/api/v1/payments/wallet/onboard/": gin.H{
			"post": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Initiate Stripe Connect Express Onboarding",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Stripe Connect onboarding URL."}},
			},
		},
		"/api/v1/payments/payment-sheet/": gin.H{
			"post": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Generate Mobile Stripe PaymentSheet Params",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Client secret & ephemeral key returned."}},
			},
		},
		"/api/v1/payments/tip/": gin.H{
			"post": gin.H{
				"tags":        []string{"Payments & Wallet"},
				"summary":     "Tip Service Provider",
				"description": "Sends a tip directly to a provider from the seeker's wallet balance post-appointment.",
				"security":    []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"provider_id", "appointment_id", "amount"},
								"properties": gin.H{
									"provider_id":    gin.H{"type": "string", "format": "uuid"},
									"appointment_id": gin.H{"type": "string", "format": "uuid"},
									"amount":         gin.H{"type": "number", "example": 15.00},
								},
							},
						},
					},
				},
				"responses": gin.H{
					"200": gin.H{"description": "Tip sent successfully."},
					"400": gin.H{"description": "Insufficient wallet balance or invalid amount."},
				},
			},
		},

		// ─── MODERATION, AUDIT & BACKGROUND CHECKS ──────────────────────────
		"/api/v1/moderation/reports/": gin.H{
			"post": gin.H{
				"tags":     []string{"Moderation & Background Checks"},
				"summary":  "Submit Moderation Report",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"resource_type", "resource_id", "reason"},
								"properties": gin.H{
									"resource_type":    gin.H{"type": "string"},
									"resource_id":      gin.H{"type": "string"},
									"reported_user_id": gin.H{"type": "string"},
									"reason":           gin.H{"type": "string"},
								},
							},
						},
					},
				},
				"responses": gin.H{"201": gin.H{"description": "Report submitted."}},
			},
		},
		"/api/v1/moderation/background-checks/initiate/": gin.H{
			"post": gin.H{
				"tags":     []string{"Moderation & Background Checks"},
				"summary":  "Initiate Checkr Background Check",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Checkr candidate invitation created."}},
			},
		},
		"/api/v1/audit/logs/": gin.H{
			"get": gin.H{
				"tags":     []string{"Moderation & Background Checks"},
				"summary":  "Get Audit Trail Logs",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Audit log entries."}},
			},
		},

		// ─── SUPERUSER & ADMIN MANAGEMENT (ALL ENDPOINTS) ───────────────────
		"/api/v1/admin/dashboard/stats": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Dashboard & Users"},
				"summary":  "Admin Dashboard KPI Analytics",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Dashboard KPI metrics."}},
			},
		},
		"/api/v1/admin/stats": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Dashboard & Users"},
				"summary":  "Admin Dashboard Summary Stats",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Dashboard summary statistics."}},
			},
		},
		"/api/v1/admin/users": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Dashboard & Users"},
				"summary":  "List & Filter Users",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "search", "in": "query", "schema": gin.H{"type": "string"}},
					{"name": "user_type", "in": "query", "schema": gin.H{"type": "string"}},
					{"name": "is_active", "in": "query", "schema": gin.H{"type": "boolean"}},
					{"name": "is_staff", "in": "query", "schema": gin.H{"type": "boolean"}},
					{"name": "page", "in": "query", "schema": gin.H{"type": "integer"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Paginated user entities."}},
			},
		},
		"/api/v1/admin/users/{id}": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Dashboard & Users"},
				"summary":  "Get User Entity Details",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "User entity with profiles."}},
			},
			"patch": gin.H{
				"tags":     []string{"Admin: Dashboard & Users"},
				"summary":  "Update User Flags / Subscription Tier (Partial)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "User updated."}},
			},
			"put": gin.H{
				"tags":     []string{"Admin: Dashboard & Users"},
				"summary":  "Update User Details (Full)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "User updated."}},
			},
			"delete": gin.H{
				"tags":     []string{"Admin: Dashboard & Users"},
				"summary":  "Soft-Delete User Account",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Soft-deleted."}},
			},
		},
		"/api/v1/admin/users/{id}/restore": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Dashboard & Users"},
				"summary":  "Restore Soft-Deleted User",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Account restored."}},
			},
		},
		"/api/v1/admin/users/{id}/gdpr-export": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Dashboard & Users"},
				"summary":  "Export Full GDPR / CCPA User Data Bundle",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "GDPR export archive."}},
			},
		},

		// ─── ADMIN: VERIFICATIONS & BACKGROUND CHECKS ───────────────────────
		"/api/v1/admin/verifications": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Verifications & Background Checks"},
				"summary":  "List Provider Identity Verifications",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Verifications list."}},
			},
		},
		"/api/v1/admin/verifications/{id}/approve": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Verifications & Background Checks"},
				"summary":  "Approve Provider Verification",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Approved."}},
			},
		},
		"/api/v1/admin/verifications/{id}/reject": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Verifications & Background Checks"},
				"summary":  "Reject Provider Verification",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Rejected."}},
			},
		},
		"/api/v1/admin/verifications/batch": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Verifications & Background Checks"},
				"summary":  "Batch Review Verifications",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"ids", "action"},
								"properties": gin.H{
									"ids":    gin.H{"type": "array", "items": gin.H{"type": "string"}},
									"action": gin.H{"type": "string", "enum": []string{"APPROVE", "REJECT"}},
									"notes":  gin.H{"type": "string"},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Batch operation executed."}},
			},
		},
		"/api/v1/admin/background-checks": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Verifications & Background Checks"},
				"summary":  "List Background Checks",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Background checks list."}},
			},
		},
		"/api/v1/admin/background-checks/{id}/override": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Verifications & Background Checks"},
				"summary":  "Override Background Check Status",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Status overridden."}},
			},
		},

		// ─── ADMIN: REPORTS & DISPUTES ──────────────────────────────────────
		"/api/v1/admin/reports": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Moderation, Reports & Disputes"},
				"summary":  "List User Moderation Reports",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Reports list."}},
			},
		},
		"/api/v1/admin/reports/{id}/resolve": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Moderation, Reports & Disputes"},
				"summary":  "Resolve Moderation Report",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Resolved."}},
			},
		},
		"/api/v1/admin/disputes": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Moderation, Reports & Disputes"},
				"summary":  "List Appointment Disputes",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Disputes list."}},
			},
		},
		"/api/v1/admin/disputes/{id}/resolve": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Moderation, Reports & Disputes"},
				"summary":  "Resolve Appointment Dispute",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Dispute resolved."}},
			},
		},
		"/api/v1/admin/disputes/{id}/reject": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Moderation, Reports & Disputes"},
				"summary":  "Reject Appointment Dispute",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Dispute rejected."}},
			},
		},

		// ─── ADMIN: FINANCIALS, PAYOUTS & SUBSCRIPTIONS ────────────────────
		"/api/v1/admin/payouts": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Financials, Payouts & Wallets"},
				"summary":  "List Payout Requests",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Payout requests list."}},
			},
		},
		"/api/v1/admin/payouts/{id}/approve": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Financials, Payouts & Wallets"},
				"summary":  "Approve Payout Request",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Payout approved."}},
			},
		},
		"/api/v1/admin/payouts/{id}/reject": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Financials, Payouts & Wallets"},
				"summary":  "Reject Payout & Refund Wallet Balance",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Payout rejected & balance refunded."}},
			},
		},
		"/api/v1/admin/wallets": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Financials, Payouts & Wallets"},
				"summary":  "List All Wallets",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Wallets list."}},
			},
		},
		"/api/v1/admin/wallets/{id}/adjust": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Financials, Payouts & Wallets"},
				"summary":  "Manual Wallet Credit / Debit Adjustment",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"amount", "reason"},
								"properties": gin.H{
									"amount": gin.H{"type": "number"},
									"reason": gin.H{"type": "string"},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Wallet adjusted."}},
			},
		},
		"/api/v1/admin/subscriptions": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Financials, Payouts & Wallets"},
				"summary":  "List Platform Subscriptions",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Subscriptions list."}},
			},
		},
		"/api/v1/admin/subscriptions/{id}/toggle": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Financials, Payouts & Wallets"},
				"summary":  "Toggle Subscription Active Status",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Subscription toggled."}},
			},
		},
		"/api/v1/admin/reports/financial": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Financials, Payouts & Wallets"},
				"summary":  "Generate Aggregate Financial & GMV Report",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "start_date", "in": "query", "schema": gin.H{"type": "string"}},
					{"name": "end_date", "in": "query", "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "GMV & platform fee summary."}},
			},
		},

		// ─── ADMIN: CATEGORIES, SERVICES & SETTINGS ─────────────────────────
		"/api/v1/admin/categories": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Categories, Services & Settings"},
				"summary":  "Create Service Category",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"name"},
								"properties": gin.H{
									"name":        gin.H{"type": "string"},
									"description": gin.H{"type": "string"},
									"image":       gin.H{"type": "string"},
								},
							},
						},
					},
				},
				"responses": gin.H{"201": gin.H{"description": "Category created."}},
			},
		},
		"/api/v1/admin/categories/{id}": gin.H{
			"put": gin.H{
				"tags":     []string{"Admin: Categories, Services & Settings"},
				"summary":  "Update Service Category (Full)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Category updated."}},
			},
			"patch": gin.H{
				"tags":     []string{"Admin: Categories, Services & Settings"},
				"summary":  "Update Service Category (Partial)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Category updated."}},
			},
			"delete": gin.H{
				"tags":     []string{"Admin: Categories, Services & Settings"},
				"summary":  "Delete Service Category",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Category deleted."}},
			},
		},
		"/api/v1/admin/catalog-services": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Categories, Services & Settings"},
				"summary":  "Create Catalog Service Offering",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"required": true,
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"category_id", "name"},
								"properties": gin.H{
									"category_id": gin.H{"type": "string"},
									"name":        gin.H{"type": "string"},
									"description": gin.H{"type": "string"},
									"base_price":  gin.H{"type": "number"},
								},
							},
						},
					},
				},
				"responses": gin.H{"201": gin.H{"description": "Catalog service created."}},
			},
		},
		"/api/v1/admin/catalog-services/{id}": gin.H{
			"put": gin.H{
				"tags":     []string{"Admin: Categories, Services & Settings"},
				"summary":  "Update Catalog Service (Full)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Catalog service updated."}},
			},
			"patch": gin.H{
				"tags":     []string{"Admin: Categories, Services & Settings"},
				"summary":  "Update Catalog Service (Partial)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Catalog service updated."}},
			},
			"delete": gin.H{
				"tags":     []string{"Admin: Categories, Services & Settings"},
				"summary":  "Delete Catalog Service",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Catalog service deleted."}},
			},
		},
		"/api/v1/admin/settings": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Categories, Services & Settings"},
				"summary":  "Get Platform Moderation Settings",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Platform settings."}},
			},
			"put": gin.H{
				"tags":     []string{"Admin: Categories, Services & Settings"},
				"summary":  "Update Platform Moderation Settings (Full)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Settings updated."}},
			},
			"patch": gin.H{
				"tags":     []string{"Admin: Categories, Services & Settings"},
				"summary":  "Update Platform Moderation Settings (Partial)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Settings updated."}},
			},
		},
		"/api/v1/admin/audit-logs": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Categories, Services & Settings"},
				"summary":  "List Admin Audit Trail",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Audit trail entries."}},
			},
		},

		// ─── ADMIN: OPERATIONS & FEATURE FLAGS ──────────────────────────────
		"/api/v1/admin/notifications/broadcast": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Operations & Feature Flags"},
				"summary":  "Broadcast Notification to User Segments",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"title", "message"},
								"properties": gin.H{
									"title":            gin.H{"type": "string"},
									"message":          gin.H{"type": "string"},
									"target_user_type": gin.H{"type": "string", "enum": []string{"ALL", "SEEKER", "PROVIDER"}},
								},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "Broadcast dispatched."}},
			},
		},
		"/api/v1/admin/feature-flags": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Operations & Feature Flags"},
				"summary":  "List Runtime Feature Flags",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Feature flags list."}},
			},
		},
		"/api/v1/admin/feature-flags/{key}": gin.H{
			"put": gin.H{
				"tags":     []string{"Admin: Operations & Feature Flags"},
				"summary":  "Toggle Feature Flag (Full)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "key", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Flag state updated."}},
			},
			"patch": gin.H{
				"tags":     []string{"Admin: Operations & Feature Flags"},
				"summary":  "Toggle Feature Flag (Partial)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "key", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Flag state updated."}},
			},
		},
		"/api/v1/admin/cache/clear": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Operations & Feature Flags"},
				"summary":  "Flush Application Memory Cache",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Cache flushed."}},
			},
		},
		"/api/v1/admin/system/health": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Operations & Feature Flags"},
				"summary":  "Get Database & Telemetry Health Stats",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Telemetry status."}},
			},
		},

		// ─── ADMIN: ENTERPRISE SECURITY & 2FA ───────────────────────────────
		"/api/v1/admin/2fa/setup": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Enterprise Security & 2FA"},
				"summary":  "Initiate TOTP 2FA Setup",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "TOTP secret & provisioning URI."}},
			},
		},
		"/api/v1/admin/2fa/verify": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Enterprise Security & 2FA"},
				"summary":  "Verify & Enable TOTP 2FA",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"requestBody": gin.H{
					"content": gin.H{
						"application/json": gin.H{
							"schema": gin.H{
								"type":     "object",
								"required": []string{"code"},
								"properties": gin.H{"code": gin.H{"type": "string", "example": "123456"}},
							},
						},
					},
				},
				"responses": gin.H{"200": gin.H{"description": "2FA activated."}},
			},
		},
		"/api/v1/admin/2fa/disable": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Enterprise Security & 2FA"},
				"summary":  "Disable TOTP 2FA",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "2FA disabled."}},
			},
		},
		"/api/v1/admin/users/{id}/impersonate": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Enterprise Security & 2FA"},
				"summary":  "Staff User Impersonation",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Short-lived JWT impersonation token."}},
			},
		},
		"/api/v1/admin/roles": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Enterprise Security & 2FA"},
				"summary":  "List RBAC Roles",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Roles list."}},
			},
			"post": gin.H{
				"tags":     []string{"Admin: Enterprise Security & 2FA"},
				"summary":  "Create Custom RBAC Role",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"201": gin.H{"description": "Role created."}},
			},
		},
		"/api/v1/admin/users/{id}/role": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Enterprise Security & 2FA"},
				"summary":  "Assign Role to User",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Role assigned."}},
			},
		},

		// ─── ADMIN: FRAUD & RISK SCORING ────────────────────────────────────
		"/api/v1/admin/fraud/risk-alerts": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Fraud & Risk Scoring"},
				"summary":  "List Fraud Risk Alerts",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Fraud alerts list."}},
			},
		},
		"/api/v1/admin/fraud/evaluate/{id}": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Fraud & Risk Scoring"},
				"summary":  "Evaluate User Fraud Risk Score",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Calculated risk score & flags."}},
			},
		},
		"/api/v1/admin/fraud/risk-alerts/{id}/resolve": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Fraud & Risk Scoring"},
				"summary":  "Resolve Risk Alert",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Risk alert resolved."}},
			},
		},

		// ─── ADMIN: TEMPLATES & SYSTEM BACKUPS ──────────────────────────────
		"/api/v1/admin/templates/emails": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Templates & System Backups"},
				"summary":  "List Notification Templates",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Notification templates list."}},
			},
		},
		"/api/v1/admin/templates/emails/{key}": gin.H{
			"put": gin.H{
				"tags":     []string{"Admin: Templates & System Backups"},
				"summary":  "Update Email Notification Template (Full)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "key", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Template updated."}},
			},
			"patch": gin.H{
				"tags":     []string{"Admin: Templates & System Backups"},
				"summary":  "Update Email Notification Template (Partial)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "key", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Template updated."}},
			},
		},
		"/api/v1/admin/templates/emails/test-send": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Templates & System Backups"},
				"summary":  "Test Dispatch Email Template",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Test email sent."}},
			},
		},
		"/api/v1/admin/system/backup": gin.H{
			"post": gin.H{
				"tags":     []string{"Admin: Templates & System Backups"},
				"summary":  "Trigger Database Backup Snapshot",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Backup triggered with checksum."}},
			},
		},
		"/api/v1/admin/system/backups": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Templates & System Backups"},
				"summary":  "List Historical Database Backup Snapshots",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Snapshots list."}},
			},
		},

		// ─── ADMIN: CSV EXPORTS ─────────────────────────────────────────────
		"/api/v1/admin/export/users": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Data Exports (CSV)"},
				"summary":  "Export Users as CSV",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "CSV stream returned."}},
			},
		},
		"/api/v1/admin/export/payouts": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Data Exports (CSV)"},
				"summary":  "Export Payouts as CSV",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "CSV stream returned."}},
			},
		},
		"/api/v1/admin/export/disputes": gin.H{
			"get": gin.H{
				"tags":     []string{"Admin: Data Exports (CSV)"},
				"summary":  "Export Disputes as CSV",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "CSV stream returned."}},
			},
		},

		// ─── WEBSOCKETS ─────────────────────────────────────────────────────
		"/ws/presence": gin.H{
			"get": gin.H{
				"tags":        []string{"WebSockets"},
				"summary":     "Real-Time User Online Presence Stream",
				"description": "WebSocket endpoint for real-time presence heartbeats. Pass JWT via `token` query param or `Sec-WebSocket-Protocol` header.",
				"responses":   gin.H{"101": gin.H{"description": "Switching Protocols to WebSocket."}},
			},
		},
		"/ws/chat/{conversation_id}": gin.H{
			"get": gin.H{
				"tags":        []string{"WebSockets"},
				"summary":     "Real-Time Chat & Direct Messaging Stream",
				"description": "WebSocket stream for sending and receiving instant chat messages.",
				"parameters": []gin.H{
					{"name": "conversation_id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"101": gin.H{"description": "Switching Protocols to WebSocket."}},
			},
		},
		"/ws/notifications": gin.H{
			"get": gin.H{
				"tags":        []string{"WebSockets"},
				"summary":     "Real-Time Push Notification Stream",
				"description": "WebSocket stream delivering instantaneous in-app notification alerts.",
				"responses":   gin.H{"101": gin.H{"description": "Switching Protocols to WebSocket."}},
			},
		},
		"/ws/tracking/{appointment_id}": gin.H{
			"get": gin.H{
				"tags":        []string{"WebSockets"},
				"summary":     "Real-Time Provider Location & Appointment Tracking",
				"parameters": []gin.H{
					{"name": "appointment_id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"101": gin.H{"description": "Switching Protocols to WebSocket."}},
			},
		},
		"/ws/admin/events": gin.H{
			"get": gin.H{
				"tags":        []string{"WebSockets"},
				"summary":     "Superuser Real-Time Administrative Platform Event Stream",
				"description": "Restricted WebSocket stream broadcasting platform-wide administrative events. Requires Staff or Superuser credentials.",
				"responses":   gin.H{"101": gin.H{"description": "Switching Protocols to WebSocket."}},
			},
		},

		// ─── ALTERNATE & MOBILE CLIENT ROUTE COMPATIBILITY ──────────────────
		"/api/v1/subscription/create/": gin.H{
			"post": gin.H{
				"tags":     []string{"In-App Purchases & Subscriptions"},
				"summary":  "Create / Validate Apple Subscription",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Subscription processed."}},
			},
		},
		"/api/v1/subscription/user/get/": gin.H{
			"get": gin.H{
				"tags":     []string{"In-App Purchases & Subscriptions"},
				"summary":  "Get User Active Subscription",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "User subscription details."}},
			},
		},
		"/api/v1/subscription/user/delete/": gin.H{
			"delete": gin.H{
				"tags":     []string{"In-App Purchases & Subscriptions"},
				"summary":  "Cancel User Subscription",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Subscription cancelled."}},
			},
		},
		"/api/v1/subscription/subscription-plans/": gin.H{
			"get": gin.H{
				"tags":      []string{"In-App Purchases & Subscriptions"},
				"summary":   "List Subscription Plans",
				"responses": gin.H{"200": gin.H{"description": "Plans list."}},
			},
		},
		"/api/v1/wallet/": gin.H{
			"get": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Get User Wallet (Root Mount)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Wallet data."}},
			},
		},
		"/api/v1/wallet/my_wallet/": gin.H{
			"get": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Get User Wallet Details",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Wallet balance."}},
			},
		},
		"/api/v1/wallet/transactions/": gin.H{
			"get": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "List Wallet Transactions",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Transactions list."}},
			},
		},
		"/api/v1/wallet/request_payout/": gin.H{
			"post": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Request Payout",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"201": gin.H{"description": "Payout requested."}},
			},
		},
		"/api/v1/wallet/onboard/": gin.H{
			"post": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Stripe Connect Onboarding",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Onboarding URL."}},
			},
		},
		"/api/v1/wallet/onboarding-status/": gin.H{
			"get": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Check Stripe Connect Onboarding Status",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Onboarding status."}},
			},
		},
		"/api/v1/wallet/stripe-dashboard/": gin.H{
			"get": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Get Express Login Link for Stripe Dashboard",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Express dashboard URL."}},
			},
		},
		"/api/v1/customer/user/": gin.H{
			"get": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Get Stripe Customer Details",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Customer details."}},
			},
		},
		"/api/v1/customer/create/": gin.H{
			"post": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Create Stripe Customer",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"201": gin.H{"description": "Stripe customer created."}},
			},
		},
		"/api/v1/customer/ephmeral/": gin.H{
			"patch": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Create Stripe Ephemeral Key",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Ephemeral key."}},
			},
		},
		"/api/v1/customer/paymentmethod/update/": gin.H{
			"patch": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Update Default Payment Method",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Payment method set."}},
			},
		},
		"/api/v1/customer/account/connect/": gin.H{
			"post": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Connect Account",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Connected."}},
			},
		},
		"/api/v1/customer/transfer/": gin.H{
			"post": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Transfer Funds",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Transfer initiated."}},
			},
		},
		"/api/v1/customer/fund-appointment/": gin.H{
			"post": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Fund / Escrow Appointment Payment",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Escrow held."}},
			},
		},
		"/api/v1/customer/fund-background-check/": gin.H{
			"post": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Pay Checkr Background Check Fee",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Payment confirmed."}},
			},
		},
		"/api/v1/customer/payment-sheet/": gin.H{
			"post": gin.H{
				"tags":     []string{"Payments & Wallet"},
				"summary":  "Generate Mobile PaymentSheet",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "PaymentSheet params."}},
			},
		},
		"/api/v1/service-packages/": gin.H{
			"get": gin.H{
				"tags":     []string{"Service Packages"},
				"summary":  "List Service Packages",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Packages list."}},
			},
			"post": gin.H{
				"tags":     []string{"Service Packages"},
				"summary":  "Create Service Package",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"201": gin.H{"description": "Package created."}},
			},
		},
		"/api/v1/service-packages/{id}/": gin.H{
			"delete": gin.H{
				"tags":     []string{"Service Packages"},
				"summary":  "Delete Service Package",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Package deleted."}},
			},
		},
		"/api/v1/services/categories/{id}/": gin.H{
			"get": gin.H{
				"tags":     []string{"Categories & Services"},
				"summary":  "Get Category by ID",
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Category details."}},
			},
		},
		"/api/v1/services/catalog-services/{id}/": gin.H{
			"get": gin.H{
				"tags":     []string{"Categories & Services"},
				"summary":  "Get Catalog Service by ID",
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Catalog service details."}},
			},
		},
		"/api/v1/services/requests/{id}/accept_proposal/": gin.H{
			"post": gin.H{
				"tags":     []string{"Service Requests & Proposals"},
				"summary":  "Accept Provider Proposal",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Proposal accepted."}},
			},
		},
		"/api/v1/services/requests/{id}/cancel_approval/": gin.H{
			"post": gin.H{
				"tags":     []string{"Service Requests & Proposals"},
				"summary":  "Cancel Proposal Approval",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Approval cancelled."}},
			},
		},
		"/api/v1/services/requests/image/": gin.H{
			"patch": gin.H{
				"tags":     []string{"Service Requests & Proposals"},
				"summary":  "Upload Request Image (Multipart)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Image attached."}},
			},
		},
		"/api/v1/services/requests/{id}/upload_image/": gin.H{
			"post": gin.H{
				"tags":     []string{"Service Requests & Proposals"},
				"summary":  "Upload Request Image by Request ID",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Image uploaded."}},
			},
		},
		"/api/v1/services/proposals/{id}/": gin.H{
			"delete": gin.H{
				"tags":     []string{"Service Requests & Proposals"},
				"summary":  "Withdraw / Delete Proposal",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Proposal deleted."}},
			},
		},
		"/api/v1/interactions/favorites/{provider_id}/": gin.H{
			"delete": gin.H{
				"tags":     []string{"Appointments & Reviews"},
				"summary":  "Remove Provider from Favorites",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "provider_id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Removed from favorites."}},
			},
		},
		"/api/v1/interactions/appointments/{id}/": gin.H{
			"delete": gin.H{
				"tags":     []string{"Appointments & Reviews"},
				"summary":  "Delete / Cancel Appointment",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Appointment deleted."}},
			},
		},
		"/api/v1/interactions/appointments/{id}/verify_code/": gin.H{
			"post": gin.H{
				"tags":     []string{"Appointments & Reviews"},
				"summary":  "Verify Provider Arrival (Snake Case Route)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Arrival verified."}},
			},
		},
		"/api/v1/chat/conversations/set_seen/": gin.H{
			"post": gin.H{
				"tags":     []string{"Chat & Messaging"},
				"summary":  "Mark Conversation Messages as Seen",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Marked seen."}},
			},
		},
		"/api/v1/chat/conversations/unblock_chat/": gin.H{
			"post": gin.H{
				"tags":     []string{"Chat & Messaging"},
				"summary":  "Unblock User in Chat",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "User unblocked."}},
			},
		},
		"/api/v1/chat/conversations/blocked_users/": gin.H{
			"get": gin.H{
				"tags":     []string{"Chat & Messaging"},
				"summary":  "List Blocked Chat Users",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Blocked users list."}},
			},
		},
		"/api/v1/chat/messages/set_seen/": gin.H{
			"post": gin.H{
				"tags":     []string{"Chat & Messaging"},
				"summary":  "Mark Single Message as Seen",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Message marked seen."}},
			},
		},
		"/api/v1/notifications/{id}/mark_as_read/": gin.H{
			"post": gin.H{
				"tags":     []string{"Notifications"},
				"summary":  "Mark Notification Read (POST variant)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Marked as read."}},
			},
		},
		"/api/v1/notifications/tokens/": gin.H{
			"get": gin.H{
				"tags":     []string{"Notifications"},
				"summary":  "List Registered Device Tokens",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Device tokens list."}},
			},
			"post": gin.H{
				"tags":     []string{"Notifications"},
				"summary":  "Register Device Token",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"201": gin.H{"description": "Token registered."}},
			},
		},
		"/api/v1/moderation/verifications/": gin.H{
			"get": gin.H{
				"tags":     []string{"Moderation & Background Checks"},
				"summary":  "Get Verification Status for Current User",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Verification status details."}},
			},
		},
		"/api/v1/moderation/background-checks/": gin.H{
			"get": gin.H{
				"tags":     []string{"Moderation & Background Checks"},
				"summary":  "Get Background Check Status for Current User",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Background check status details."}},
			},
		},
		"/api/v1/moderation/background-checks/config/": gin.H{
			"get": gin.H{
				"tags":     []string{"Moderation & Background Checks"},
				"summary":  "Get Checkr Payment & Configuration Mode",
				"responses": gin.H{"200": gin.H{"description": "Checkr configuration."}},
			},
		},
		"/api/v1/moderation/background-checks/{id}/resync/": gin.H{
			"post": gin.H{
				"tags":     []string{"Moderation & Background Checks"},
				"summary":  "Resync Background Check Status from Checkr API",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"parameters": []gin.H{
					{"name": "id", "in": "path", "required": true, "schema": gin.H{"type": "string"}},
				},
				"responses": gin.H{"200": gin.H{"description": "Resynced status."}},
			},
		},
		"/api/v1/audit/": gin.H{
			"get": gin.H{
				"tags":     []string{"Moderation & Background Checks"},
				"summary":  "Get Audit Trail Logs (Root alias)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Audit trail."}},
			},
		},
		"/api/notifications/": gin.H{
			"get": gin.H{
				"tags":     []string{"Notifications"},
				"summary":  "Get Notifications (Direct Flutter route)",
				"security": []gin.H{{"bearerAuth": []string{}}},
				"responses": gin.H{"200": gin.H{"description": "Notifications list."}},
			},
		},
	}
}
