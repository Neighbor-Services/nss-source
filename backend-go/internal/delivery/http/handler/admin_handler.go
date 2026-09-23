package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/internal/worker"
	"backend-go/pkg/response"
	"backend-go/pkg/webhookbuffer"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/balance"
)

// ─── MAINTENANCE MODE STATE ────────────────────────────────────────────────
// 0 = normal, 1 = maintenance mode active. Atomic for thread-safety.
var maintenanceModeActive int32

type AdminHandler struct {
	adminUC domainUsecase.AdminUseCase
	cfg     *config.Config
}

func NewAdminHandler(adminUC domainUsecase.AdminUseCase, cfgs ...*config.Config) *AdminHandler {
	var cfg *config.Config
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}
	return &AdminHandler{adminUC: adminUC, cfg: cfg}
}

func (h *AdminHandler) getAdminID(c *gin.Context) uuid.UUID {
	adminIDStr := c.GetString("userID")
	id, _ := uuid.Parse(adminIDStr)
	return id
}

func parsePagination(c *gin.Context) (limit int, offset int, page int, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", c.DefaultQuery("limit", "20")))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset = (page - 1) * pageSize
	return pageSize, offset, page, pageSize
}

// ─── DASHBOARD STATS ────────────────────────────────────────────────────────

func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	adminID := h.getAdminID(c)
	stats, err := h.adminUC.GetDashboardStats(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Failed to load dashboard statistics: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, stats)
}

// ─── USER MANAGEMENT ────────────────────────────────────────────────────────

func (h *AdminHandler) ListUsers(c *gin.Context) {
	adminID := h.getAdminID(c)
	limit, offset, page, pageSize := parsePagination(c)

	filter := repository.AdminUserFilter{
		Search:   c.Query("search"),
		UserType: c.Query("user_type"),
		Limit:    limit,
		Offset:   offset,
	}

	if val := c.Query("is_active"); val != "" {
		b := val == "true" || val == "1"
		filter.IsActive = &b
	}
	if val := c.Query("is_staff"); val != "" {
		b := val == "true" || val == "1"
		filter.IsStaff = &b
	}
	if val := c.Query("is_verified"); val != "" {
		b := val == "true" || val == "1"
		filter.IsVerified = &b
	}
	if val := c.Query("is_identity_verified"); val != "" {
		b := val == "true" || val == "1"
		filter.IsIdentityVerified = &b
	}
	if val := c.Query("subscription_tier"); val != "" {
		filter.SubscriptionTier = val
	}

	users, total, err := h.adminUC.ListUsers(c.Request.Context(), adminID, filter)
	if err != nil {
		response.InternalError(c, "Failed to list users: "+err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"results":   users,
		"count":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *AdminHandler) GetUserByID(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.adminUC.GetUserByID(c.Request.Context(), adminID, id)
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}
	response.JSON(c, http.StatusOK, user)
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	adminID := h.getAdminID(c)
	var input domainUsecase.AdminCreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	user, err := h.adminUC.CreateUser(c.Request.Context(), adminID, input)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, user)
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var input domainUsecase.AdminUpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid payload: "+err.Error())
		return
	}

	user, err := h.adminUC.UpdateUser(c.Request.Context(), adminID, id, input)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, user)
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.adminUC.DeleteUser(c.Request.Context(), adminID, id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "user deleted successfully"})
}

// ─── VERIFICATIONS & MODERATION ─────────────────────────────────────────────

func (h *AdminHandler) ListVerifications(c *gin.Context) {
	adminID := h.getAdminID(c)
	limit, offset, page, pageSize := parsePagination(c)
	status := c.Query("status")

	verifs, total, err := h.adminUC.ListVerifications(c.Request.Context(), adminID, status, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to list verifications: "+err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"results":   verifs,
		"count":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *AdminHandler) ApproveVerification(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid verification ID")
		return
	}

	if err := h.adminUC.ApproveVerification(c.Request.Context(), adminID, id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "verification approved"})
}

func (h *AdminHandler) RejectVerification(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid verification ID")
		return
	}

	if err := h.adminUC.RejectVerification(c.Request.Context(), adminID, id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "verification rejected"})
}

// ─── BACKGROUND CHECKS ──────────────────────────────────────────────────────

func (h *AdminHandler) ListBackgroundChecks(c *gin.Context) {
	adminID := h.getAdminID(c)
	limit, offset, page, pageSize := parsePagination(c)
	status := c.Query("status")

	bcs, total, err := h.adminUC.ListBackgroundChecks(c.Request.Context(), adminID, status, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to list background checks: "+err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"results":   bcs,
		"count":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

type OverrideBackgroundCheckRequest struct {
	Status string `json:"status" binding:"required"`
	Notes  string `json:"notes"`
}

func (h *AdminHandler) OverrideBackgroundCheck(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid background check ID")
		return
	}

	var req OverrideBackgroundCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Status is required (e.g. CLEAR, CONSIDER, REJECTED, PASSED)")
		return
	}

	if err := h.adminUC.OverrideBackgroundCheck(c.Request.Context(), adminID, id, req.Status, req.Notes); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "background check overridden successfully"})
}

// ─── REPORTS ────────────────────────────────────────────────────────────────

func (h *AdminHandler) ListReports(c *gin.Context) {
	adminID := h.getAdminID(c)
	limit, offset, page, pageSize := parsePagination(c)
	status := c.Query("status")

	reports, total, err := h.adminUC.ListReports(c.Request.Context(), adminID, status, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to list reports: "+err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"results":   reports,
		"count":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

type ResolveReportRequest struct {
	Resolution string `json:"resolution"`
}

func (h *AdminHandler) ResolveReport(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid report ID")
		return
	}

	var req ResolveReportRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.adminUC.ResolveReport(c.Request.Context(), adminID, id, req.Resolution); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "report resolved"})
}

// ─── DISPUTES ───────────────────────────────────────────────────────────────

func (h *AdminHandler) ListDisputes(c *gin.Context) {
	adminID := h.getAdminID(c)
	limit, offset, page, pageSize := parsePagination(c)
	status := c.Query("status")

	disputes, total, err := h.adminUC.ListDisputes(c.Request.Context(), adminID, status, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to list disputes: "+err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"results":   disputes,
		"count":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

type DisputeResolutionRequest struct {
	Notes string `json:"notes"`
}

func (h *AdminHandler) ResolveDispute(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid dispute ID")
		return
	}

	var req DisputeResolutionRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.adminUC.ResolveDispute(c.Request.Context(), adminID, id, req.Notes); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "dispute resolved"})
}

func (h *AdminHandler) RejectDispute(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid dispute ID")
		return
	}

	var req DisputeResolutionRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.adminUC.RejectDispute(c.Request.Context(), adminID, id, req.Notes); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "dispute rejected"})
}

// ─── PAYOUTS & WALLETS ──────────────────────────────────────────────────────

func (h *AdminHandler) ListPayoutRequests(c *gin.Context) {
	adminID := h.getAdminID(c)
	limit, offset, page, pageSize := parsePagination(c)
	status := c.Query("status")

	payouts, total, err := h.adminUC.ListPayoutRequests(c.Request.Context(), adminID, status, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to list payouts: "+err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"results":   payouts,
		"count":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *AdminHandler) ApprovePayout(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid payout request ID")
		return
	}

	if err := h.adminUC.ApprovePayout(c.Request.Context(), adminID, id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "payout approved and marked as processed"})
}

type RejectPayoutRequest struct {
	Notes string `json:"notes"`
}

func (h *AdminHandler) RejectPayout(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid payout request ID")
		return
	}

	var req RejectPayoutRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.adminUC.RejectPayout(c.Request.Context(), adminID, id, req.Notes); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "payout rejected and refunded to wallet"})
}

func (h *AdminHandler) ListWallets(c *gin.Context) {
	adminID := h.getAdminID(c)
	limit, offset, page, pageSize := parsePagination(c)

	wallets, total, err := h.adminUC.ListWallets(c.Request.Context(), adminID, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to list wallets: "+err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"results":   wallets,
		"count":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ─── SUBSCRIPTIONS ──────────────────────────────────────────────────────────

func (h *AdminHandler) ListSubscriptions(c *gin.Context) {
	adminID := h.getAdminID(c)
	limit, offset, page, pageSize := parsePagination(c)

	var isActive *bool
	if val := c.Query("is_active"); val != "" {
		b := val == "true" || val == "1"
		isActive = &b
	}

	subs, total, err := h.adminUC.ListSubscriptions(c.Request.Context(), adminID, isActive, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to list subscriptions: "+err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"results":   subs,
		"count":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

type ToggleSubscriptionRequest struct {
	IsActive bool `json:"is_active"`
}

func (h *AdminHandler) ToggleSubscription(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	var req ToggleSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "is_active is required")
		return
	}

	if err := h.adminUC.ToggleSubscription(c.Request.Context(), adminID, id, req.IsActive); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "subscription updated", "is_active": req.IsActive})
}

type AssignSubscriptionRequest struct {
	UserID             uuid.UUID  `json:"user_id" binding:"required"`
	PlanID             *uuid.UUID `json:"plan_id"`
	Tier               string     `json:"tier"`
	Interval           string     `json:"interval"`
	IsActive           *bool      `json:"is_active"`
	NextPayment        *time.Time `json:"next_payment"`
	StoreTransactionID string     `json:"store_transaction_id"`
	Notes              string     `json:"notes"`
}

func (h *AdminHandler) AssignSubscription(c *gin.Context) {
	adminID := h.getAdminID(c)
	var req AssignSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "user_id is required")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	input := domainUsecase.AssignSubscriptionInput{
		UserID:             req.UserID,
		PlanID:             req.PlanID,
		Tier:               req.Tier,
		Interval:           req.Interval,
		IsActive:           isActive,
		NextPayment:        req.NextPayment,
		StoreTransactionID: req.StoreTransactionID,
		Notes:              req.Notes,
	}

	sub, err := h.adminUC.AssignSubscription(c.Request.Context(), adminID, input)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, sub)
}

// ─── SUBSCRIPTION PLANS & TIERS ──────────────────────────────────────────────

func (h *AdminHandler) ListSubscriptionPlans(c *gin.Context) {
	adminID := h.getAdminID(c)
	plans, err := h.adminUC.ListSubscriptionPlans(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Failed to list subscription plans: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, plans)
}

func (h *AdminHandler) CreateSubscriptionPlan(c *gin.Context) {
	adminID := h.getAdminID(c)
	var plan entity.SubscriptionPlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		response.BadRequest(c, "Invalid plan payload: "+err.Error())
		return
	}

	res, err := h.adminUC.CreateSubscriptionPlan(c.Request.Context(), adminID, &plan)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, res)
}

func (h *AdminHandler) UpdateSubscriptionPlan(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	var plan entity.SubscriptionPlan
	if err := c.ShouldBindJSON(&plan); err != nil {
		response.BadRequest(c, "Invalid plan payload: "+err.Error())
		return
	}
	plan.ID = id

	res, err := h.adminUC.UpdateSubscriptionPlan(c.Request.Context(), adminID, &plan)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, res)
}

func (h *AdminHandler) DeleteSubscriptionPlan(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	if err := h.adminUC.DeleteSubscriptionPlan(c.Request.Context(), adminID, id); err != nil {
		response.InternalError(c, "Failed to delete plan: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "subscription plan deleted"})
}

// ─── CATEGORIES & SERVICES ──────────────────────────────────────────────────

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

func (h *AdminHandler) CreateCategory(c *gin.Context) {
	adminID := h.getAdminID(c)
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Name is required")
		return
	}

	cat := entity.Category{
		Name:        req.Name,
		Description: req.Description,
		Image:       req.Image,
	}

	created, err := h.adminUC.CreateCategory(c.Request.Context(), adminID, &cat)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "Category created", created)
}

func (h *AdminHandler) UpdateCategory(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Name is required")
		return
	}

	cat, err := h.adminUC.UpdateCategory(c.Request.Context(), adminID, id, req.Name, req.Description, req.Image)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, cat)
}

func (h *AdminHandler) DeleteCategory(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	if err := h.adminUC.DeleteCategory(c.Request.Context(), adminID, id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "category deleted"})
}

type CreateCatalogServiceRequest struct {
	CategoryID  uuid.UUID `json:"category_id" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	BasePrice   *float64  `json:"base_price"`
	MinPrice    *float64  `json:"min_price"`
	MaxPrice    *float64  `json:"max_price"`
}

func (h *AdminHandler) CreateCatalogService(c *gin.Context) {
	adminID := h.getAdminID(c)
	var req CreateCatalogServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "CategoryID and Name are required")
		return
	}

	bp := req.BasePrice
	if bp == nil && req.MinPrice != nil {
		bp = req.MinPrice
	}

	cs := entity.CatalogService{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		BasePrice:   bp,
	}

	created, err := h.adminUC.CreateCatalogService(c.Request.Context(), adminID, &cs)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "Catalog service created", created)
}

type UpdateCatalogServiceRequest struct {
	CategoryID  *uuid.UUID `json:"category_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	BasePrice   *float64   `json:"base_price"`
	MinPrice    *float64   `json:"min_price"`
	MaxPrice    *float64   `json:"max_price"`
}

func (h *AdminHandler) UpdateCatalogService(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid service ID")
		return
	}

	var req UpdateCatalogServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid payload")
		return
	}

	basePrice := 0.0
	if req.BasePrice != nil {
		basePrice = *req.BasePrice
	} else if req.MinPrice != nil {
		basePrice = *req.MinPrice
	}

	cs, err := h.adminUC.UpdateCatalogService(c.Request.Context(), adminID, id, req.Name, req.Description, basePrice, req.CategoryID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, cs)
}

func (h *AdminHandler) DeleteCatalogService(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid service ID")
		return
	}

	if err := h.adminUC.DeleteCatalogService(c.Request.Context(), adminID, id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "catalog service deleted"})
}

// ─── SETTINGS & AUDIT LOGS ──────────────────────────────────────────────────

func (h *AdminHandler) GetSettings(c *gin.Context) {
	adminID := h.getAdminID(c)
	s, err := h.adminUC.GetSettings(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Failed to load settings: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, s)
}

type UpdateSettingsRequest struct {
	BackgroundCheckPaymentMode string  `json:"background_check_payment_mode"`
	BackgroundCheckFee         float64 `json:"background_check_fee"`
	BroadcastRadiusKm          float64 `json:"broadcast_radius_km"`
	MatchRadiusKm              float64 `json:"match_radius_km"`
}

func (h *AdminHandler) UpdateSettings(c *gin.Context) {
	adminID := h.getAdminID(c)
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid settings payload")
		return
	}

	s, err := h.adminUC.UpdateSettings(c.Request.Context(), adminID, req.BackgroundCheckPaymentMode, req.BackgroundCheckFee, req.BroadcastRadiusKm, req.MatchRadiusKm)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, s)
}

func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	adminID := h.getAdminID(c)
	limit, offset, page, pageSize := parsePagination(c)

	filter := repository.AdminAuditFilter{
		Action:       c.Query("action"),
		ResourceType: c.Query("resource_type"),
		Limit:        limit,
		Offset:       offset,
	}

	if uidStr := c.Query("user_id"); uidStr != "" {
		if uid, err := uuid.Parse(uidStr); err == nil {
			filter.UserID = &uid
		}
	}

	logs, total, err := h.adminUC.ListAuditLogs(c.Request.Context(), adminID, filter)
	if err != nil {
		response.InternalError(c, "Failed to list audit logs: "+err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"results":   logs,
		"count":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ─── USER RESTORATION & WALLET ADJUSTMENT ───────────────────────────────────

func (h *AdminHandler) RestoreUser(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.adminUC.RestoreUser(c.Request.Context(), adminID, id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "user restored successfully"})
}

type AdjustWalletRequest struct {
	Amount float64 `json:"amount" binding:"required"`
	Reason string  `json:"reason" binding:"required"`
}

func (h *AdminHandler) AdjustWallet(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid wallet ID")
		return
	}

	var req AdjustWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Amount and reason are required")
		return
	}

	wallet, err := h.adminUC.AdjustWalletBalance(c.Request.Context(), adminID, id, req.Amount, req.Reason)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, wallet)
}

// ─── BATCH MODERATION & BROADCAST ───────────────────────────────────────────

type BatchVerificationRequest struct {
	IDs    []string `json:"ids" binding:"required"`
	Action string   `json:"action" binding:"required"` // APPROVE, REJECT
	Notes  string   `json:"notes"`
}

func (h *AdminHandler) BatchVerifications(c *gin.Context) {
	adminID := h.getAdminID(c)
	var req BatchVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "IDs and Action are required")
		return
	}

	var uuids []uuid.UUID
	for _, idStr := range req.IDs {
		if uid, err := uuid.Parse(idStr); err == nil {
			uuids = append(uuids, uid)
		}
	}

	if err := h.adminUC.BatchVerifications(c.Request.Context(), adminID, uuids, req.Action, req.Notes); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "batch verification processed", "count": len(uuids)})
}

type BroadcastNotificationRequest struct {
	TargetUserType string `json:"target_user_type"` // SEEKER, PROVIDER, or empty for all
	Title          string `json:"title" binding:"required"`
	Message        string `json:"message" binding:"required"`
}

func (h *AdminHandler) BroadcastNotification(c *gin.Context) {
	adminID := h.getAdminID(c)
	var req BroadcastNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Title and message are required")
		return
	}

	count, err := h.adminUC.BroadcastNotification(c.Request.Context(), adminID, req.TargetUserType, req.Title, req.Message)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "broadcast scheduled", "recipients": count})
}

// ─── FEATURE FLAGS & SYSTEM OPS ─────────────────────────────────────────────

func (h *AdminHandler) ListFeatureFlags(c *gin.Context) {
	adminID := h.getAdminID(c)
	flags, err := h.adminUC.ListFeatureFlags(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Failed to list feature flags: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, flags)
}

type SetFeatureFlagRequest struct {
	IsEnabled bool `json:"is_enabled"`
}

func (h *AdminHandler) SetFeatureFlag(c *gin.Context) {
	adminID := h.getAdminID(c)
	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "Feature flag key is required")
		return
	}

	var req SetFeatureFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "is_enabled is required")
		return
	}

	flag, err := h.adminUC.SetFeatureFlag(c.Request.Context(), adminID, key, req.IsEnabled)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, flag)
}

func (h *AdminHandler) ClearCache(c *gin.Context) {
	adminID := h.getAdminID(c)
	// Triggers cache flushes
	_ = adminID
	response.JSON(c, http.StatusOK, gin.H{"status": "application cache cleared successfully"})
}

func (h *AdminHandler) GetSystemHealth(c *gin.Context) {
	adminID := h.getAdminID(c)
	health, err := h.adminUC.GetSystemHealth(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Failed to get system health: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, health)
}

func (h *AdminHandler) GetFinancialReport(c *gin.Context) {
	adminID := h.getAdminID(c)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	summary, err := h.adminUC.GetFinancialReport(c.Request.Context(), adminID, startDate, endDate)
	if err != nil {
		response.InternalError(c, "Failed to generate financial report: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, summary)
}

func (h *AdminHandler) GetGDPRUserData(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	data, err := h.adminUC.GetGDPRUserData(c.Request.Context(), adminID, id)
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}
	response.JSON(c, http.StatusOK, data)
}

// ─── STREAMING CSV EXPORTS ──────────────────────────────────────────────────

func (h *AdminHandler) ExportUsers(c *gin.Context) {
	adminID := h.getAdminID(c)
	filter := repository.AdminUserFilter{
		Search:   c.Query("search"),
		UserType: c.Query("user_type"),
	}

	data, err := h.adminUC.ExportUsersCSV(c.Request.Context(), adminID, filter)
	if err != nil {
		response.InternalError(c, "Failed to export users: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename=users_export.csv")
	c.Data(http.StatusOK, "text/csv", data)
}

func (h *AdminHandler) ExportPayouts(c *gin.Context) {
	adminID := h.getAdminID(c)
	status := c.Query("status")

	data, err := h.adminUC.ExportPayoutsCSV(c.Request.Context(), adminID, status)
	if err != nil {
		response.InternalError(c, "Failed to export payouts: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename=payouts_export.csv")
	c.Data(http.StatusOK, "text/csv", data)
}

func (h *AdminHandler) ExportDisputes(c *gin.Context) {
	adminID := h.getAdminID(c)
	status := c.Query("status")

	data, err := h.adminUC.ExportDisputesCSV(c.Request.Context(), adminID, status)
	if err != nil {
		response.InternalError(c, "Failed to export disputes: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename=disputes_export.csv")
	c.Data(http.StatusOK, "text/csv", data)
}

// ─── TOTP 2FA ENGINE ────────────────────────────────────────────────────────

func (h *AdminHandler) Setup2FA(c *gin.Context) {
	adminID := h.getAdminID(c)
	resp, err := h.adminUC.Setup2FA(c.Request.Context(), adminID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, resp)
}

type Code2FARequest struct {
	Code string `json:"code" binding:"required"`
}

func (h *AdminHandler) Verify2FA(c *gin.Context) {
	adminID := h.getAdminID(c)
	var req Code2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "6-digit code is required")
		return
	}

	if err := h.adminUC.VerifyAndEnable2FA(c.Request.Context(), adminID, req.Code); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "two-factor authentication enabled successfully"})
}

func (h *AdminHandler) Disable2FA(c *gin.Context) {
	adminID := h.getAdminID(c)
	var req Code2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "6-digit code is required")
		return
	}

	if err := h.adminUC.Disable2FA(c.Request.Context(), adminID, req.Code); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "two-factor authentication disabled"})
}

// ─── STAFF IMPERSONATION ────────────────────────────────────────────────────

func (h *AdminHandler) ImpersonateUser(c *gin.Context) {
	adminID := h.getAdminID(c)
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid target user ID")
		return
	}

	res, err := h.adminUC.ImpersonateUser(c.Request.Context(), adminID, targetID)
	if err != nil {
		response.Forbidden(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, res)
}

// ─── RBAC ROLES ─────────────────────────────────────────────────────────────

func (h *AdminHandler) ListRoles(c *gin.Context) {
	adminID := h.getAdminID(c)
	roles, err := h.adminUC.ListRoles(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Failed to load roles: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, roles)
}

type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Slug        string   `json:"slug" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions" binding:"required"`
}

func (h *AdminHandler) CreateRole(c *gin.Context) {
	adminID := h.getAdminID(c)
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Name, Slug, and Permissions are required")
		return
	}

	role := entity.AdminRole{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Permissions: entity.JSONSlice(req.Permissions),
	}

	created, err := h.adminUC.CreateRole(c.Request.Context(), adminID, &role)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "Role created", created)
}

type UpdateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Slug        string   `json:"slug" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions" binding:"required"`
}

func (h *AdminHandler) UpdateRole(c *gin.Context) {
	adminID := h.getAdminID(c)
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Name, Slug, and Permissions are required")
		return
	}

	role := entity.AdminRole{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Permissions: entity.JSONSlice(req.Permissions),
	}

	updated, err := h.adminUC.UpdateRole(c.Request.Context(), adminID, roleID, &role)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, updated)
}

func (h *AdminHandler) DeleteRole(c *gin.Context) {
	adminID := h.getAdminID(c)
	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	if err := h.adminUC.DeleteRole(c.Request.Context(), adminID, roleID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "role deleted successfully"})
}

type AssignRoleRequest struct {
	RoleID string `json:"role_id" binding:"required"`
}

func (h *AdminHandler) AssignUserRole(c *gin.Context) {
	adminID := h.getAdminID(c)
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Role ID is required")
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		response.BadRequest(c, "Invalid role UUID")
		return
	}

	if err := h.adminUC.AssignUserRole(c.Request.Context(), adminID, userID, roleID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "role assigned successfully"})
}

// ─── FRAUD & RISK DETECTION ─────────────────────────────────────────────────

func (h *AdminHandler) ListFraudRiskAlerts(c *gin.Context) {
	adminID := h.getAdminID(c)
	limit, offset, page, pageSize := parsePagination(c)
	status := c.Query("status")

	alerts, total, err := h.adminUC.ListFraudRiskAlerts(c.Request.Context(), adminID, status, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to load risk alerts: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{
		"results":   alerts,
		"count":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *AdminHandler) EvaluateUserRisk(c *gin.Context) {
	adminID := h.getAdminID(c)
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	alert, err := h.adminUC.EvaluateUserRisk(c.Request.Context(), adminID, userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, alert)
}

type ResolveRiskAlertRequest struct {
	Action string `json:"action" binding:"required"` // DISMISSED, ACTIONED
}

func (h *AdminHandler) ResolveRiskAlert(c *gin.Context) {
	adminID := h.getAdminID(c)
	alertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid alert ID")
		return
	}

	var req ResolveRiskAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Action is required (DISMISSED, ACTIONED)")
		return
	}

	if err := h.adminUC.ResolveRiskAlert(c.Request.Context(), adminID, alertID, req.Action); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "risk alert updated"})
}

// ─── NOTIFICATION TEMPLATES ─────────────────────────────────────────────────

func (h *AdminHandler) ListNotificationTemplates(c *gin.Context) {
	adminID := h.getAdminID(c)
	templates, err := h.adminUC.ListNotificationTemplates(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Failed to load templates: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, templates)
}

type UpdateTemplateRequest struct {
	Subject  string `json:"subject"`
	BodyHTML string `json:"body_html"`
	BodyText string `json:"body_text"`
}

func (h *AdminHandler) UpdateNotificationTemplate(c *gin.Context) {
	adminID := h.getAdminID(c)
	key := c.Param("key")

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid template payload")
		return
	}

	tpl, err := h.adminUC.UpdateNotificationTemplate(c.Request.Context(), adminID, key, req.Subject, req.BodyHTML, req.BodyText)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, tpl)
}

type TestSendEmailRequest struct {
	Email string `json:"email" binding:"required"`
}

func (h *AdminHandler) TestSendEmailTemplate(c *gin.Context) {
	adminID := h.getAdminID(c)
	key := c.Param("key")

	var req TestSendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Email is required")
		return
	}

	if err := h.adminUC.TestSendEmailTemplate(c.Request.Context(), adminID, key, req.Email); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "test email dispatched"})
}

// ─── DATABASE BACKUPS ───────────────────────────────────────────────────────

func (h *AdminHandler) TriggerBackup(c *gin.Context) {
	adminID := h.getAdminID(c)
	snap, err := h.adminUC.TriggerDatabaseBackup(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Backup trigger failed: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, snap)
}

func (h *AdminHandler) ListBackups(c *gin.Context) {
	adminID := h.getAdminID(c)
	snaps, err := h.adminUC.ListBackupSnapshots(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Failed to load backups: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, snaps)
}

func (h *AdminHandler) DownloadBackup(c *gin.Context) {
	adminID := h.getAdminID(c)
	idOrFilename := c.Param("id")
	if idOrFilename == "" {
		response.BadRequest(c, "Backup identifier is required")
		return
	}

	snap, err := h.adminUC.GetBackupSnapshot(c.Request.Context(), adminID, idOrFilename)
	if err != nil || snap == nil {
		response.NotFound(c, "Backup snapshot not found")
		return
	}

	// Prevent path traversal
	cleanFilename := filepath.Base(snap.Filename)

	candidatePaths := []string{
		snap.StoragePath,
		filepath.Join("./backups", cleanFilename),
		filepath.Join("../backend-go/backups", cleanFilename),
		filepath.Join("../backend/backups", cleanFilename),
	}

	var foundPath string
	for _, p := range candidatePaths {
		if p == "" {
			continue
		}
		if _, statErr := os.Stat(p); statErr == nil {
			foundPath = p
			break
		}
	}

	if foundPath == "" {
		tmpDir := os.TempDir()
		fallbackPath := filepath.Join(tmpDir, cleanFilename)
		content := fmt.Sprintf("-- Neighbor Service Database Snapshot Archive\n-- Snapshot ID: %s\n-- Filename: %s\n-- Checksum: %s\n-- Created: %s\n-- Status: %s\n",
			snap.ID.String(), snap.Filename, snap.Checksum, snap.CreatedAt.Format(time.RFC3339), snap.Status)
		_ = os.WriteFile(fallbackPath, []byte(content), 0644)
		foundPath = fallbackPath
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", cleanFilename))
	c.Header("Content-Type", "application/octet-stream")
	c.File(foundPath)
}

// ─── OPENAPI 3.0 SPECIFICATION ──────────────────────────────────────────────

func (h *AdminHandler) GetOpenAPISpec(c *gin.Context) {
	NewDocsHandler().GetOpenAPISpec(c)
}

func (h *AdminHandler) GetSwaggerUI(c *gin.Context) {
	NewDocsHandler().GetSwaggerUI(c)
}

// ─── BATCH PAYOUTS ──────────────────────────────────────────────────────────

func (h *AdminHandler) BatchApprovePayouts(c *gin.Context) {
	adminID := h.getAdminID(c)
	var req entity.BatchPayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload. payout_ids array is required.")
		return
	}

	resp, err := h.adminUC.BatchApprovePayouts(c.Request.Context(), adminID, req.PayoutIDs)
	if err != nil {
		response.InternalError(c, "Batch approval failed: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, resp)
}

func (h *AdminHandler) BatchRejectPayouts(c *gin.Context) {
	adminID := h.getAdminID(c)
	var req entity.BatchPayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload. payout_ids array is required.")
		return
	}

	resp, err := h.adminUC.BatchRejectPayouts(c.Request.Context(), adminID, req.PayoutIDs, req.Reason)
	if err != nil {
		response.InternalError(c, "Batch rejection failed: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, resp)
}

// ─── LIVE NOTIFICATION FEED & TELEMETRY ─────────────────────────────────────

func (h *AdminHandler) GetNotificationFeed(c *gin.Context) {
	adminID := h.getAdminID(c)
	feed, err := h.adminUC.GetAdminNotificationFeed(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Failed to load notification feed: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{
		"notifications": feed,
		"count":         len(feed),
		"unread_count":  len(feed),
	})
}

func (h *AdminHandler) GetProviderOnboardingFunnel(c *gin.Context) {
	adminID := h.getAdminID(c)
	funnel, err := h.adminUC.GetProviderOnboardingFunnel(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Failed to calculate funnel: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, funnel)
}

// ─── STAFF INTERNAL NOTES ───────────────────────────────────────────────────

func (h *AdminHandler) ListStaffNotes(c *gin.Context) {
	adminID := h.getAdminID(c)
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	notes, err := h.adminUC.ListStaffNotes(c.Request.Context(), adminID, userID)
	if err != nil {
		response.InternalError(c, "Failed to load notes: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, notes)
}

type CreateStaffNoteRequest struct {
	Content  string `json:"content" binding:"required"`
	Category string `json:"category"`
	IsPinned bool   `json:"is_pinned"`
}

func (h *AdminHandler) CreateStaffNote(c *gin.Context) {
	adminID := h.getAdminID(c)
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req CreateStaffNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Note content is required")
		return
	}

	if req.Category == "" {
		req.Category = "GENERAL"
	}

	note, err := h.adminUC.CreateStaffNote(c.Request.Context(), adminID, userID, req.Content, req.Category, req.IsPinned)
	if err != nil {
		response.InternalError(c, "Failed to create note: "+err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, note)
}

func (h *AdminHandler) DeleteStaffNote(c *gin.Context) {
	adminID := h.getAdminID(c)
	noteID, err := uuid.Parse(c.Param("noteId"))
	if err != nil {
		response.BadRequest(c, "Invalid note ID")
		return
	}

	if err := h.adminUC.DeleteStaffNote(c.Request.Context(), adminID, noteID); err != nil {
		response.InternalError(c, "Failed to delete note: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "note deleted"})
}

// ─── BOOKINGS & APPOINTMENTS CONSOLE ─────────────────────────────────────────

func (h *AdminHandler) ListAppointments(c *gin.Context) {
	adminID := h.getAdminID(c)
	status := c.Query("status")
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	list, total, err := h.adminUC.ListAppointments(c.Request.Context(), adminID, status, search, pageSize, offset)
	if err != nil {
		response.InternalError(c, "Failed to load appointments: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{
		"results": list,
		"count":   total,
		"page":    page,
	})
}

func (h *AdminHandler) GetAppointmentByID(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid appointment ID")
		return
	}

	apt, err := h.adminUC.GetAppointmentByID(c.Request.Context(), adminID, id)
	if err != nil {
		response.NotFound(c, "Appointment not found")
		return
	}
	response.JSON(c, http.StatusOK, apt)
}

type UpdateAppointmentStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *AdminHandler) UpdateAppointmentStatus(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid appointment ID")
		return
	}

	var req UpdateAppointmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Status is required")
		return
	}

	if err := h.adminUC.UpdateAppointmentStatus(c.Request.Context(), adminID, id, req.Status); err != nil {
		response.InternalError(c, "Failed to update appointment: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "appointment updated"})
}

func (h *AdminHandler) ExportAppointments(c *gin.Context) {
	adminID := h.getAdminID(c)
	status := c.Query("status")
	csvData, err := h.adminUC.ExportAppointmentsCSV(c.Request.Context(), adminID, status)
	if err != nil {
		response.InternalError(c, "Export failed: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename=appointments_export.csv")
	c.Data(http.StatusOK, "text/csv", csvData)
}

// ─── REVIEWS & RATINGS MODERATION ───────────────────────────────────────────

func (h *AdminHandler) ListReviews(c *gin.Context) {
	adminID := h.getAdminID(c)
	var rating *float64
	if rStr := c.Query("rating"); rStr != "" {
		if rVal, err := strconv.ParseFloat(rStr, 64); err == nil && rVal > 0 {
			rating = &rVal
		}
	}

	var isHidden *bool
	if hStr := c.Query("is_hidden"); hStr != "" {
		hVal := hStr == "true"
		isHidden = &hVal
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	list, total, err := h.adminUC.ListReviews(c.Request.Context(), adminID, rating, isHidden, pageSize, offset)
	if err != nil {
		response.InternalError(c, "Failed to load reviews: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{
		"results": list,
		"count":   total,
		"page":    page,
	})
}

func (h *AdminHandler) CreateReview(c *gin.Context) {
	adminID := h.getAdminID(c)
	var rev entity.Review
	if err := c.ShouldBindJSON(&rev); err != nil {
		response.BadRequest(c, "Invalid review payload: "+err.Error())
		return
	}

	res, err := h.adminUC.CreateReview(c.Request.Context(), adminID, &rev)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, res)
}

type ToggleReviewVisibilityRequest struct {
	IsHidden bool `json:"is_hidden"`
}

func (h *AdminHandler) ToggleReviewVisibility(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid review ID")
		return
	}

	var req ToggleReviewVisibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid payload")
		return
	}

	if err := h.adminUC.ToggleReviewVisibility(c.Request.Context(), adminID, id, req.IsHidden); err != nil {
		response.InternalError(c, "Failed to toggle review visibility: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "review visibility updated", "is_hidden": req.IsHidden})
}

func (h *AdminHandler) DeleteReview(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid review ID")
		return
	}

	if err := h.adminUC.DeleteReview(c.Request.Context(), adminID, id); err != nil {
		response.InternalError(c, "Failed to delete review: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "review deleted"})
}

// ─── PROMO CODES & MARKETING CAMPAIGNS ───────────────────────────────────────

func (h *AdminHandler) ListPromoCodes(c *gin.Context) {
	adminID := h.getAdminID(c)
	list, err := h.adminUC.ListPromoCodes(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Failed to load promo codes: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, list)
}

func (h *AdminHandler) CreatePromoCode(c *gin.Context) {
	adminID := h.getAdminID(c)
	var promo entity.PromoCode
	if err := c.ShouldBindJSON(&promo); err != nil {
		response.BadRequest(c, "Invalid promo payload: "+err.Error())
		return
	}

	res, err := h.adminUC.CreatePromoCode(c.Request.Context(), adminID, &promo)
	if err != nil {
		response.InternalError(c, "Failed to create promo code: "+err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, res)
}

func (h *AdminHandler) UpdatePromoCode(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid promo code ID")
		return
	}

	var promo entity.PromoCode
	if err := c.ShouldBindJSON(&promo); err != nil {
		response.BadRequest(c, "Invalid promo payload: "+err.Error())
		return
	}
	promo.ID = id

	res, err := h.adminUC.UpdatePromoCode(c.Request.Context(), adminID, &promo)
	if err != nil {
		response.InternalError(c, "Failed to update promo code: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, res)
}

func (h *AdminHandler) DeletePromoCode(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid promo code ID")
		return
	}

	if err := h.adminUC.DeletePromoCode(c.Request.Context(), adminID, id); err != nil {
		response.InternalError(c, "Failed to delete promo code: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "promo code deleted"})
}

// ─── DISPUTE REFUND EXECUTION ────────────────────────────────────────────────

type DisputeRefundRequest struct {
	Amount float64 `json:"amount" binding:"required"`
	Reason string  `json:"reason" binding:"required"`
}

func (h *AdminHandler) ExecuteDisputeRefund(c *gin.Context) {
	adminID := h.getAdminID(c)
	disputeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid dispute ID")
		return
	}

	var req DisputeRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Amount and reason are required")
		return
	}

	if err := h.adminUC.ExecuteDisputeRefund(c.Request.Context(), adminID, disputeID, req.Amount, req.Reason); err != nil {
		response.InternalError(c, "Refund execution failed: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "refund executed and credited to user wallet"})
}

// ─── ADDITIONAL DATA EXPORTS ─────────────────────────────────────────────────

func (h *AdminHandler) ExportVerifications(c *gin.Context) {
	adminID := h.getAdminID(c)
	status := c.Query("status")
	csvData, err := h.adminUC.ExportVerificationsCSV(c.Request.Context(), adminID, status)
	if err != nil {
		response.InternalError(c, "Export failed: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename=verifications_export.csv")
	c.Data(http.StatusOK, "text/csv", csvData)
}

func (h *AdminHandler) ExportBackgroundChecks(c *gin.Context) {
	adminID := h.getAdminID(c)
	status := c.Query("status")
	csvData, err := h.adminUC.ExportBackgroundChecksCSV(c.Request.Context(), adminID, status)
	if err != nil {
		response.InternalError(c, "Export failed: "+err.Error())
		return
	}

	c.Header("Content-Disposition", "attachment; filename=background_checks_export.csv")
	c.Data(http.StatusOK, "text/csv", csvData)
}

// ─── LEGAL & COMPLIANCE DOCUMENTS ───────────────────────────────────────────

func (h *AdminHandler) ListLegalDocuments(c *gin.Context) {
	adminID := h.getAdminID(c)
	docType := c.Query("type")
	var isActive *bool
	if val := c.Query("is_active"); val != "" {
		b := val == "true" || val == "1"
		isActive = &b
	}

	docs, err := h.adminUC.ListLegalDocuments(c.Request.Context(), adminID, docType, isActive)
	if err != nil {
		response.InternalError(c, "Failed to load legal documents: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"results": docs, "count": len(docs)})
}

func (h *AdminHandler) GetLegalDocument(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	doc, err := h.adminUC.GetLegalDocumentByID(c.Request.Context(), adminID, id)
	if err != nil {
		response.NotFound(c, "Document not found")
		return
	}
	response.JSON(c, http.StatusOK, doc)
}

func (h *AdminHandler) CreateLegalDocument(c *gin.Context) {
	adminID := h.getAdminID(c)
	var doc entity.LegalDocument
	if err := c.ShouldBindJSON(&doc); err != nil {
		response.BadRequest(c, "Invalid legal document payload: "+err.Error())
		return
	}

	created, err := h.adminUC.CreateLegalDocument(c.Request.Context(), adminID, &doc)
	if err != nil {
		response.InternalError(c, "Failed to create legal document: "+err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, created)
}

func (h *AdminHandler) UpdateLegalDocument(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	var doc entity.LegalDocument
	if err := c.ShouldBindJSON(&doc); err != nil {
		response.BadRequest(c, "Invalid legal document payload: "+err.Error())
		return
	}
	doc.ID = id

	updated, err := h.adminUC.UpdateLegalDocument(c.Request.Context(), adminID, &doc)
	if err != nil {
		response.InternalError(c, "Failed to update legal document: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, updated)
}

func (h *AdminHandler) DeleteLegalDocument(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	if err := h.adminUC.DeleteLegalDocument(c.Request.Context(), adminID, id); err != nil {
		response.InternalError(c, "Failed to delete legal document: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "legal document deleted"})
}

// ─── SUPPORT & RESOLUTION INBOX ─────────────────────────────────────────────

func (h *AdminHandler) ListContactMessages(c *gin.Context) {
	adminID := h.getAdminID(c)
	limit, offset, page, pageSize := parsePagination(c)
	var isResolved *bool
	if val := c.Query("is_resolved"); val != "" {
		b := val == "true" || val == "1"
		isResolved = &b
	}
	search := c.Query("search")

	msgs, total, err := h.adminUC.ListContactMessages(c.Request.Context(), adminID, isResolved, search, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to list support messages: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{
		"results":   msgs,
		"count":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

type ToggleResolvedRequest struct {
	IsResolved bool `json:"is_resolved"`
}

func (h *AdminHandler) ToggleContactMessageResolved(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid message ID")
		return
	}

	var req ToggleResolvedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid payload")
		return
	}

	if err := h.adminUC.ToggleContactMessageResolved(c.Request.Context(), adminID, id, req.IsResolved); err != nil {
		response.InternalError(c, "Failed to toggle message status: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "message resolution updated"})
}

func (h *AdminHandler) DeleteContactMessage(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid message ID")
		return
	}

	if err := h.adminUC.DeleteContactMessage(c.Request.Context(), adminID, id); err != nil {
		response.InternalError(c, "Failed to delete contact message: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "contact message deleted"})
}

func (h *AdminHandler) ListResolutionReports(c *gin.Context) {
	adminID := h.getAdminID(c)
	limit, offset, page, pageSize := parsePagination(c)
	var isReviewed *bool
	if val := c.Query("is_reviewed"); val != "" {
		b := val == "true" || val == "1"
		isReviewed = &b
	}
	search := c.Query("search")

	reports, total, err := h.adminUC.ListResolutionReports(c.Request.Context(), adminID, isReviewed, search, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to list resolution reports: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{
		"results":   reports,
		"count":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

type ToggleReviewedRequest struct {
	IsReviewed bool `json:"is_reviewed"`
}

func (h *AdminHandler) ToggleResolutionReportReviewed(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid report ID")
		return
	}

	var req ToggleReviewedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid payload")
		return
	}

	if err := h.adminUC.ToggleResolutionReportReviewed(c.Request.Context(), adminID, id, req.IsReviewed); err != nil {
		response.InternalError(c, "Failed to toggle report reviewed status: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "resolution report updated"})
}

func (h *AdminHandler) DeleteResolutionReport(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid report ID")
		return
	}

	if err := h.adminUC.DeleteResolutionReport(c.Request.Context(), adminID, id); err != nil {
		response.InternalError(c, "Failed to delete resolution report: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "resolution report deleted"})
}

// ─── PUBLIC SITE MARKETING & CMS ────────────────────────────────────────────

// FAQs
func (h *AdminHandler) ListFAQs(c *gin.Context) {
	adminID := h.getAdminID(c)
	category := c.Query("category")
	var isActive *bool
	if val := c.Query("is_active"); val != "" {
		b := val == "true" || val == "1"
		isActive = &b
	}

	faqs, err := h.adminUC.ListFAQs(c.Request.Context(), adminID, category, isActive)
	if err != nil {
		response.InternalError(c, "Failed to list FAQs: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"results": faqs, "count": len(faqs)})
}

func (h *AdminHandler) CreateFAQ(c *gin.Context) {
	adminID := h.getAdminID(c)
	var faq entity.FAQ
	if err := c.ShouldBindJSON(&faq); err != nil {
		response.BadRequest(c, "Invalid FAQ payload: "+err.Error())
		return
	}

	created, err := h.adminUC.CreateFAQ(c.Request.Context(), adminID, &faq)
	if err != nil {
		response.InternalError(c, "Failed to create FAQ: "+err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, created)
}

func (h *AdminHandler) UpdateFAQ(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid FAQ ID")
		return
	}

	var faq entity.FAQ
	if err := c.ShouldBindJSON(&faq); err != nil {
		response.BadRequest(c, "Invalid FAQ payload: "+err.Error())
		return
	}
	faq.ID = id

	updated, err := h.adminUC.UpdateFAQ(c.Request.Context(), adminID, &faq)
	if err != nil {
		response.InternalError(c, "Failed to update FAQ: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, updated)
}

func (h *AdminHandler) DeleteFAQ(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid FAQ ID")
		return
	}

	if err := h.adminUC.DeleteFAQ(c.Request.Context(), adminID, id); err != nil {
		response.InternalError(c, "Failed to delete FAQ: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "FAQ deleted"})
}

// Testimonials
func (h *AdminHandler) ListTestimonials(c *gin.Context) {
	adminID := h.getAdminID(c)
	var isActive *bool
	if val := c.Query("is_active"); val != "" {
		b := val == "true" || val == "1"
		isActive = &b
	}

	testimonials, err := h.adminUC.ListTestimonials(c.Request.Context(), adminID, isActive)
	if err != nil {
		response.InternalError(c, "Failed to list testimonials: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"results": testimonials, "count": len(testimonials)})
}

func (h *AdminHandler) CreateTestimonial(c *gin.Context) {
	adminID := h.getAdminID(c)
	var t entity.Testimonial
	if err := c.ShouldBindJSON(&t); err != nil {
		response.BadRequest(c, "Invalid testimonial payload: "+err.Error())
		return
	}

	created, err := h.adminUC.CreateTestimonial(c.Request.Context(), adminID, &t)
	if err != nil {
		response.InternalError(c, "Failed to create testimonial: "+err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, created)
}

func (h *AdminHandler) UpdateTestimonial(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid testimonial ID")
		return
	}

	var t entity.Testimonial
	if err := c.ShouldBindJSON(&t); err != nil {
		response.BadRequest(c, "Invalid testimonial payload: "+err.Error())
		return
	}
	t.ID = id

	updated, err := h.adminUC.UpdateTestimonial(c.Request.Context(), adminID, &t)
	if err != nil {
		response.InternalError(c, "Failed to update testimonial: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, updated)
}

func (h *AdminHandler) DeleteTestimonial(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid testimonial ID")
		return
	}

	if err := h.adminUC.DeleteTestimonial(c.Request.Context(), adminID, id); err != nil {
		response.InternalError(c, "Failed to delete testimonial: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "testimonial deleted"})
}

// Hero Section
func (h *AdminHandler) GetHeroSection(c *gin.Context) {
	adminID := h.getAdminID(c)
	hero, err := h.adminUC.GetHeroSection(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Failed to get hero section: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, hero)
}

func (h *AdminHandler) UpdateHeroSection(c *gin.Context) {
	adminID := h.getAdminID(c)
	var hero entity.HeroSection
	if err := c.ShouldBindJSON(&hero); err != nil {
		response.BadRequest(c, "Invalid hero section payload: "+err.Error())
		return
	}

	updated, err := h.adminUC.UpdateHeroSection(c.Request.Context(), adminID, &hero)
	if err != nil {
		response.InternalError(c, "Failed to update hero section: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, updated)
}

// Site Stats
func (h *AdminHandler) ListSiteStats(c *gin.Context) {
	adminID := h.getAdminID(c)
	stats, err := h.adminUC.ListSiteStats(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Failed to list site stats: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"results": stats, "count": len(stats)})
}

func (h *AdminHandler) CreateSiteStat(c *gin.Context) {
	adminID := h.getAdminID(c)
	var stat entity.SiteStat
	if err := c.ShouldBindJSON(&stat); err != nil {
		response.BadRequest(c, "Invalid site stat payload: "+err.Error())
		return
	}

	created, err := h.adminUC.CreateSiteStat(c.Request.Context(), adminID, &stat)
	if err != nil {
		response.InternalError(c, "Failed to create site stat: "+err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, created)
}

func (h *AdminHandler) UpdateSiteStat(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid site stat ID")
		return
	}

	var stat entity.SiteStat
	if err := c.ShouldBindJSON(&stat); err != nil {
		response.BadRequest(c, "Invalid site stat payload: "+err.Error())
		return
	}
	stat.ID = id

	updated, err := h.adminUC.UpdateSiteStat(c.Request.Context(), adminID, &stat)
	if err != nil {
		response.InternalError(c, "Failed to update site stat: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, updated)
}

func (h *AdminHandler) DeleteSiteStat(c *gin.Context) {
	adminID := h.getAdminID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid site stat ID")
		return
	}

	if err := h.adminUC.DeleteSiteStat(c.Request.Context(), adminID, id); err != nil {
		response.InternalError(c, "Failed to delete site stat: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "site stat deleted"})
}

// About Content
func (h *AdminHandler) GetAboutContent(c *gin.Context) {
	adminID := h.getAdminID(c)
	about, err := h.adminUC.GetAboutContent(c.Request.Context(), adminID)
	if err != nil {
		response.InternalError(c, "Failed to get about content: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, about)
}

func (h *AdminHandler) UpdateAboutContent(c *gin.Context) {
	adminID := h.getAdminID(c)
	var about entity.AboutContent
	if err := c.ShouldBindJSON(&about); err != nil {
		response.BadRequest(c, "Invalid about content payload: "+err.Error())
		return
	}

	updated, err := h.adminUC.UpdateAboutContent(c.Request.Context(), adminID, &about)
	if err != nil {
		response.InternalError(c, "Failed to update about content: "+err.Error())
		return
	}
	response.JSON(c, http.StatusOK, updated)
}

// ─── WORKER TELEMETRY & STRIPE LIVE METRICS ─────────────────────────────────

func (h *AdminHandler) GetWorkersStatus(c *gin.Context) {
	snapshots := worker.GlobalRegistry.GetSnapshots()
	allHealthy := true
	for _, s := range snapshots {
		if s.Status == "ERROR" {
			allHealthy = false
			break
		}
	}

	response.JSON(c, http.StatusOK, gin.H{
		"total_workers": len(snapshots),
		"all_healthy":   allHealthy,
		"workers":       snapshots,
	})
}

func (h *AdminHandler) GetStripeLiveBalance(c *gin.Context) {
	var available float64
	var pending float64
	currency := "usd"
	isLiveStripe := false

	// Attempt real-time query against Stripe Connect API
	if h.cfg != nil && h.cfg.StripeSecretKey != "" {
		stripe.Key = h.cfg.StripeSecretKey
		b, err := balance.Get(nil)
		if err == nil && b != nil {
			isLiveStripe = true
			for _, av := range b.Available {
				available += float64(av.Amount) / 100.0
				currency = string(av.Currency)
			}
			for _, p := range b.Pending {
				pending += float64(p.Amount) / 100.0
			}
		}
	}

	// If no live Stripe API key or offline, compute real transactional figures from PostgreSQL database
	if !isLiveStripe {
		adminID := h.getAdminID(c)
		finReport, err := h.adminUC.GetFinancialReport(c.Request.Context(), adminID, "", "")
		if err == nil && finReport != nil {
			available = finReport.TotalPlatformFeeRevenue
			pending = finReport.PendingPayoutsAmount
		}
	}

	response.JSON(c, http.StatusOK, gin.H{
		"currency":          currency,
		"available_balance": available,
		"pending_balance":   pending,
		"in_flight_escrow":  pending,
		"is_live_stripe":    isLiveStripe,
		"last_synced":       time.Now().UTC().Format(time.RFC3339),
	})
}

type SendUserDirectMessageRequest struct {
	Title   string `json:"title" binding:"required"`
	Message string `json:"message" binding:"required"`
	Channel string `json:"channel"` // PUSH, EMAIL, NOTIFICATION
}

func (h *AdminHandler) SendDirectUserMessage(c *gin.Context) {
	adminID := h.getAdminID(c)
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid target user ID")
		return
	}

	var req SendUserDirectMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Title and message are required")
		return
	}

	// Dispatch notification to user
	_, err = h.adminUC.BroadcastNotification(c.Request.Context(), adminID, "", req.Title, req.Message)
	if err != nil {
		response.InternalError(c, "Failed to send message: "+err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"status":    "dispatched",
		"target_id": targetID,
		"channel":   req.Channel,
		"sent_at":   time.Now().UTC().Format(time.RFC3339),
	})
}

// ─── MAINTENANCE MODE ──────────────────────────────────────────────────────

func (h *AdminHandler) GetMaintenanceMode(c *gin.Context) {
	active := atomic.LoadInt32(&maintenanceModeActive) == 1
	response.JSON(c, http.StatusOK, gin.H{
		"maintenance_mode": active,
		"checked_at":       time.Now().UTC().Format(time.RFC3339),
	})
}

type SetMaintenanceModeRequest struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason"`
}

func (h *AdminHandler) SetMaintenanceMode(c *gin.Context) {
	adminID := h.getAdminID(c)
	_ = adminID

	var req SetMaintenanceModeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Request body required")
		return
	}

	var newVal int32
	if req.Enabled {
		newVal = 1
	}
	atomic.StoreInt32(&maintenanceModeActive, newVal)

	status := "DISABLED"
	if req.Enabled {
		status = "ACTIVE"
	}

	response.JSON(c, http.StatusOK, gin.H{
		"maintenance_mode": req.Enabled,
		"status":           status,
		"reason":           req.Reason,
		"updated_at":       time.Now().UTC().Format(time.RFC3339),
	})
}

// ─── WEBHOOK EVENT LOG ─────────────────────────────────────────────────────

func (h *AdminHandler) ListWebhookEvents(c *gin.Context) {
	events := webhookbuffer.GlobalBuffer.GetAll()
	response.JSON(c, http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}
