package handler

import (
	"fmt"
	"net/http"
	"path/filepath"

	"backend-go/internal/domain/entity"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/pkg/media"
	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InteractionHandler struct {
	interactionUC domainUsecase.InteractionUseCase
}

func NewInteractionHandler(interactionUC domainUsecase.InteractionUseCase) *InteractionHandler {
	return &InteractionHandler{interactionUC: interactionUC}
}

func (h *InteractionHandler) wrapAppointments(appointments []entity.Appointment, currentUserID uuid.UUID) []gin.H {
	wrapped := make([]gin.H, 0, len(appointments))
	for _, apt := range appointments {
		var userProfile interface{}
		role := "provider"

		if apt.SeekerID == currentUserID {
			role = "seeker"
			if apt.Provider != nil && apt.Provider.Profile != nil {
				userProfile = apt.Provider.Profile
			} else if apt.ProviderProfile != nil {
				userProfile = apt.ProviderProfile
			}
		} else {
			if apt.Seeker != nil && apt.Seeker.Profile != nil {
				userProfile = apt.Seeker.Profile
			} else if apt.SeekerProfile != nil {
				userProfile = apt.SeekerProfile
			}
		}

		wrapped = append(wrapped, gin.H{
			"appointment": apt,
			"user":        userProfile,
			"role":        role,
		})
	}
	return wrapped
}

func (h *InteractionHandler) GetFavorites(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	favorites, err := h.interactionUC.GetFavorites(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, "Failed to load favorites")
		return
	}
	response.JSON(c, http.StatusOK, favorites)
}

func (h *InteractionHandler) CreateFavorite(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := cleanUUID(userIDStr)

	var req struct {
		FavoriteUserID  string `json:"favorite_user"`
		FavoriteUserAlt string `json:"favorite_user_id"`
		Provider        string `json:"provider"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "favorite_user or provider is required")
		return
	}

	target := req.FavoriteUserID
	if target == "" {
		target = req.FavoriteUserAlt
	}
	if target == "" {
		target = req.Provider
	}

	favUUID, err := cleanUUID(target)
	if err != nil {
		response.BadRequest(c, "invalid provider/favorite_user ID")
		return
	}

	fav, err := h.interactionUC.CreateFavorite(c.Request.Context(), userUUID, favUUID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, fav)
}

func (h *InteractionHandler) DeleteFavorite(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := cleanUUID(userIDStr)

	providerIDStr := c.Param("provider_id")
	if providerIDStr == "" {
		providerIDStr = c.Param("id")
	}
	providerUUID, err := cleanUUID(providerIDStr)
	if err != nil {
		response.BadRequest(c, "invalid provider ID")
		return
	}

	if err := h.interactionUC.DeleteFavorite(c.Request.Context(), userUUID, providerUUID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "removed"})
}

func (h *InteractionHandler) GetReviews(c *gin.Context) {
	providerIDStr := c.Query("provider")
	if providerIDStr == "" {
		providerIDStr = c.Param("provider_id")
	}
	providerUUID, err := cleanUUID(providerIDStr)
	if err != nil {
		response.BadRequest(c, "valid provider ID is required")
		return
	}

	reviews, err := h.interactionUC.GetReviews(c.Request.Context(), providerUUID)
	if err != nil {
		response.InternalError(c, "Failed to load reviews")
		return
	}
	response.JSON(c, http.StatusOK, reviews)
}

func (h *InteractionHandler) CreateReview(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := cleanUUID(userIDStr)

	var req struct {
		ProviderID    string  `json:"provider"`
		ProviderIDAlt string  `json:"provider_id"`
		ProfileID     string  `json:"profile_id"`
		Rating        float64 `json:"rating"`
		Comment       string  `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid review payload")
		return
	}

	target := req.ProviderID
	if target == "" {
		target = req.ProviderIDAlt
	}
	if target == "" {
		target = req.ProfileID
	}

	provUUID, err := cleanUUID(target)
	if err != nil {
		response.BadRequest(c, "valid provider ID is required")
		return
	}

	if req.Rating <= 0 || req.Rating > 5 {
		response.BadRequest(c, "Rating must be between 1 and 5")
		return
	}

	rev := entity.Review{
		ProviderID: provUUID,
		Rating:     req.Rating,
		Comment:    req.Comment,
	}

	created, err := h.interactionUC.CreateReview(c.Request.Context(), userUUID, &rev)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, created)
}

func (h *InteractionHandler) GetAppointments(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userType := c.GetString("userType")
	userUUID, _ := uuid.Parse(userIDStr)
	status := c.Query("status")

	appointments, err := h.interactionUC.GetAppointments(c.Request.Context(), userUUID, userType, status)
	if err != nil {
		response.InternalError(c, "Failed to load appointments")
		return
	}
	response.JSON(c, http.StatusOK, h.wrapAppointments(appointments, userUUID))
}

func (h *InteractionHandler) CreateAppointment(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := uuid.Parse(userIDStr)

	var apt entity.Appointment
	if err := c.ShouldBindJSON(&apt); err != nil {
		response.BadRequest(c, "Invalid appointment payload")
		return
	}

	created, err := h.interactionUC.CreateAppointment(c.Request.Context(), userUUID, &apt)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	wrapped := h.wrapAppointments([]entity.Appointment{*created}, userUUID)
	if len(wrapped) > 0 {
		response.JSON(c, http.StatusCreated, wrapped[0])
		return
	}
	response.Created(c, "Appointment scheduled", created)
}

func (h *InteractionHandler) VerifyArrivalCode(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := cleanUUID(userIDStr)

	idStr := c.Param("id")
	aptUUID, err := cleanUUID(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid appointment ID")
		return
	}

	var body struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "code is required")
		return
	}

	apt, err := h.interactionUC.VerifyArrivalCode(c.Request.Context(), userUUID, aptUUID, body.Code)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	wrapped := h.wrapAppointments([]entity.Appointment{*apt}, userUUID)
	if len(wrapped) > 0 {
		response.JSON(c, http.StatusOK, gin.H{"status": "verified", "data": wrapped[0]})
		return
	}
	response.JSON(c, http.StatusOK, gin.H{"status": "verified", "data": apt})
}

func (h *InteractionHandler) CompleteAppointment(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := cleanUUID(userIDStr)

	idStr := c.Param("id")
	aptUUID, err := cleanUUID(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid appointment ID")
		return
	}

	var body struct {
		Amount float64 `json:"amount"`
	}
	_ = c.ShouldBindJSON(&body)

	released, commission, err := h.interactionUC.CompleteAppointment(c.Request.Context(), userUUID, aptUUID, body.Amount)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"status":              "appointment completed",
		"funds_released":      released,
		"commission_deducted": commission,
	})
}

func (h *InteractionHandler) CancelAppointment(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := cleanUUID(userIDStr)

	idStr := c.Param("id")
	aptUUID, err := cleanUUID(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid appointment ID")
		return
	}

	if err := h.interactionUC.CancelAppointment(c.Request.Context(), userUUID, aptUUID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "appointment cancelled"})
}

func (h *InteractionHandler) GetDisputes(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := cleanUUID(userIDStr)

	disputes, err := h.interactionUC.GetDisputes(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, "Failed to load disputes")
		return
	}
	response.JSON(c, http.StatusOK, disputes)
}

func (h *InteractionHandler) CreateDispute(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := cleanUUID(userIDStr)

	var dispute entity.Dispute
	if err := c.ShouldBindJSON(&dispute); err != nil {
		response.BadRequest(c, "Invalid dispute payload")
		return
	}

	created, err := h.interactionUC.CreateDispute(c.Request.Context(), userUUID, &dispute)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, created)
}

func (h *InteractionHandler) UploadDisputeEvidence(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, _ := cleanUUID(userIDStr)

	idStr := c.Param("id")
	disputeUUID, err := cleanUUID(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid dispute ID")
		return
	}

	file, err := c.FormFile("evidence")
	if err != nil {
		response.BadRequest(c, "No evidence file provided")
		return
	}

	mediaDir := media.ResolveMediaDir("evidence")
	filename := fmt.Sprintf("evidence_%s_%s", uuid.New().String(), filepath.Base(file.Filename))
	savePath := filepath.Join(mediaDir, filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.InternalError(c, "Failed to save evidence file")
		return
	}

	evidenceURL := "/media/evidence/" + filename

	if err := h.interactionUC.UploadDisputeEvidence(c.Request.Context(), userUUID, disputeUUID, evidenceURL); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "evidence uploaded successfully", "evidence_url": evidenceURL})
}
