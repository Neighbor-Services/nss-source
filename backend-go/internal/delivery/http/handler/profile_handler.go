package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"backend-go/internal/domain/entity"
	"backend-go/internal/domain/repository"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/pkg/media"
	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProfileHandler struct {
	profileUC domainUsecase.ProfileUseCase
}

func NewProfileHandler(profileUC domainUsecase.ProfileUseCase) *ProfileHandler {
	return &ProfileHandler{profileUC: profileUC}
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	idParam := c.Param("id")
	if idParam != "" && idParam != "me" && idParam != "popular" {
		if uid, err := uuid.Parse(idParam); err == nil {
			profile, err := h.profileUC.GetProfile(c.Request.Context(), uuid.Nil, &uid)
			if err != nil {
				response.NotFound(c, "Profile not found")
				return
			}
			response.JSON(c, http.StatusOK, profile)
			return
		}
	}

	// Check if querying by specific user ID (?user=UUID)
	userParam := c.Query("user")
	if userParam != "" {
		if uid, err := uuid.Parse(userParam); err == nil {
			profile, err := h.profileUC.GetProfile(c.Request.Context(), uid, nil)
			if err == nil && profile != nil {
				response.JSON(c, http.StatusOK, gin.H{"providers": []interface{}{profile}})
				return
			}
		}
		response.JSON(c, http.StatusOK, gin.H{"providers": []interface{}{}})
		return
	}

	// Popularity query
	popular := c.Query("popular") == "true" || idParam == "popular"
	userType := c.Query("user_type")
	search := c.Query("search")
	categorySlug := c.Query("category")
	categoryName := c.Query("category_name")
	serviceName := c.Query("service_name")
	serviceID := c.Query("service_id")
	city := c.Query("city")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var ratingMinVal, priceMinVal, priceMaxVal *float64
	if rStr := c.Query("rating_min"); rStr != "" {
		if r, err := strconv.ParseFloat(rStr, 64); err == nil {
			ratingMinVal = &r
		}
	}
	if pStr := c.Query("price_min"); pStr != "" {
		if p, err := strconv.ParseFloat(pStr, 64); err == nil {
			priceMinVal = &p
		}
	}
	if pStr := c.Query("price_max"); pStr != "" {
		if p, err := strconv.ParseFloat(pStr, 64); err == nil {
			priceMaxVal = &p
		}
	}

	var latVal, lngVal, radiusVal *float64
	if latStr := c.Query("lat"); latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil && lat != 0 {
			latVal = &lat
		}
	} else if latStr := c.Query("latitude"); latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil && lat != 0 {
			latVal = &lat
		}
	}

	if lngStr := c.Query("lng"); lngStr != "" {
		if lng, err := strconv.ParseFloat(lngStr, 64); err == nil && lng != 0 {
			lngVal = &lng
		}
	} else if lngStr := c.Query("longitude"); lngStr != "" {
		if lng, err := strconv.ParseFloat(lngStr, 64); err == nil && lng != 0 {
			lngVal = &lng
		}
	}

	defaultRadius := 25.0
	if radStr := c.Query("radius"); radStr != "" {
		if rad, err := strconv.ParseFloat(radStr, 64); err == nil && rad > 0 {
			radiusVal = &rad
		}
	} else if radStr := c.Query("radius_km"); radStr != "" {
		if rad, err := strconv.ParseFloat(radStr, 64); err == nil && rad > 0 {
			radiusVal = &rad
		}
	} else if radStr := c.Query("max_distance"); radStr != "" {
		if rad, err := strconv.ParseFloat(radStr, 64); err == nil && rad > 0 {
			radiusVal = &rad
		}
	}

	if radiusVal == nil {
		radiusVal = &defaultRadius
	}

	// If seeker / user is logged in and didn't provide lat/lng in query, check their profile lat/lng!
	if latVal == nil || lngVal == nil {
		userIDStr := c.GetString("userID")
		if userUUID, err := uuid.Parse(userIDStr); err == nil && userUUID != uuid.Nil {
			if userProfile, err := h.profileUC.GetProfile(c.Request.Context(), userUUID, nil); err == nil && userProfile != nil {
				if userProfile.Latitude != 0 || userProfile.Longitude != 0 {
					lat := userProfile.Latitude
					lng := userProfile.Longitude
					latVal = &lat
					lngVal = &lng
				}
			}
		}
	}

	if popular || search != "" || userType != "" || categorySlug != "" || categoryName != "" || serviceName != "" || serviceID != "" || city != "" || ratingMinVal != nil || latVal != nil {
		profiles, err := h.profileUC.ListProfiles(c.Request.Context(), repository.ProfileFilterParams{
			UserType:     userType,
			Popular:      popular,
			Search:       search,
			CategorySlug: categorySlug,
			CategoryName: categoryName,
			ServiceName:  serviceName,
			ServiceID:    serviceID,
			RatingMin:    ratingMinVal,
			PriceMin:     priceMinVal,
			PriceMax:     priceMaxVal,
			City:         city,
			Latitude:     latVal,
			Longitude:    lngVal,
			RadiusKm:     radiusVal,
			Limit:        limit,
			Offset:       offset,
		})
		if err != nil {
			response.InternalError(c, "Failed to load profiles")
			return
		}
		if popular || userType == "PROVIDER" {
			response.JSON(c, http.StatusOK, gin.H{"providers": profiles})
			return
		}
		response.JSON(c, http.StatusOK, gin.H{"providers": profiles})
		return
	}

	// Default: Current authenticated user profile
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	profile, err := h.profileUC.GetProfile(c.Request.Context(), userUUID, nil)
	if err != nil {
		response.NotFound(c, "Profile not found")
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"profile": profile})
}

func (h *ProfileHandler) GetProfileMe(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	profile, err := h.profileUC.GetProfile(c.Request.Context(), userUUID, nil)
	if err != nil {
		response.NotFound(c, "Profile not found")
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"profile": profile})
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "Invalid JSON body")
		return
	}

	profile, err := h.profileUC.UpdateProfile(c.Request.Context(), userUUID, updates)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"profile": profile})
}

func (h *ProfileHandler) UploadPicture(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var imageURL string
	file, err := c.FormFile("image")
	if err == nil {
		mediaDir := media.ResolveMediaDir("profiles")
		filename := fmt.Sprintf("profile_%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
		savePath := filepath.Join(mediaDir, filename)
		if err := c.SaveUploadedFile(file, savePath); err == nil {
			imageURL = "/media/profiles/" + filename
		}
	} else {
		var body struct {
			Image string `json:"image"`
		}
		if err := c.ShouldBindJSON(&body); err == nil && body.Image != "" {
			imageURL = body.Image
		}
	}

	if imageURL == "" {
		response.BadRequest(c, "No image provided")
		return
	}

	profile, err := h.profileUC.UpdateProfilePicture(c.Request.Context(), userUUID, imageURL)
	if err != nil {
		response.InternalError(c, "Failed to update profile picture")
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"profile": profile})
}

func (h *ProfileHandler) GetAbout(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	about, err := h.profileUC.GetAbout(c.Request.Context(), userUUID)
	if err != nil || about == nil {
		response.JSON(c, http.StatusOK, gin.H{})
		return
	}

	response.JSON(c, http.StatusOK, about)
}

func (h *ProfileHandler) GetAboutUser(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		userIDStr = c.GetString("userID")
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid user_id")
		return
	}

	about, _ := h.profileUC.GetAbout(c.Request.Context(), userUUID)
	profile, _ := h.profileUC.GetProfile(c.Request.Context(), userUUID, nil)

	response.JSON(c, http.StatusOK, gin.H{
		"about": about,
		"user":  profile,
	})
}

func (h *ProfileHandler) UpdateAbout(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "Invalid JSON body")
		return
	}

	about, err := h.profileUC.UpdateAbout(c.Request.Context(), userUUID, updates)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, about)
}

func (h *ProfileHandler) GetPortfolios(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	portfolios, err := h.profileUC.GetPortfolios(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, "Failed to load portfolios")
		return
	}

	response.JSON(c, http.StatusOK, portfolios)
}

func (h *ProfileHandler) CreatePortfolio(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var item entity.Portfolio
	file, err := c.FormFile("image")
	if err == nil {
		mediaDir := media.ResolveMediaDir("portfolio")
		filename := fmt.Sprintf("portfolio_%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
		savePath := filepath.Join(mediaDir, filename)
		if err := c.SaveUploadedFile(file, savePath); err == nil {
			item.Image = "/media/portfolio/" + filename
		}
		item.Description = c.PostForm("description")
	} else {
		if err := c.ShouldBindJSON(&item); err != nil {
			response.BadRequest(c, "Invalid portfolio payload")
			return
		}
	}

	created, err := h.profileUC.CreatePortfolio(c.Request.Context(), userUUID, &item)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, created)
}

func (h *ProfileHandler) DeletePortfolio(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	itemID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid portfolio item ID")
		return
	}

	if err := h.profileUC.DeletePortfolio(c.Request.Context(), userUUID, itemID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "deleted"})
}

func (h *ProfileHandler) GetServicePackages(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	pkgs, err := h.profileUC.GetServicePackages(c.Request.Context(), userUUID)
	if err != nil {
		response.InternalError(c, "Failed to load service packages")
		return
	}

	response.JSON(c, http.StatusOK, pkgs)
}

func (h *ProfileHandler) CreateServicePackage(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	var pkg entity.ServicePackage
	if err := c.ShouldBindJSON(&pkg); err != nil {
		response.BadRequest(c, "Invalid package payload")
		return
	}

	created, err := h.profileUC.CreateServicePackage(c.Request.Context(), userUUID, &pkg)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, created)
}

func (h *ProfileHandler) DeleteServicePackage(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	pkgID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid package ID")
		return
	}

	if err := h.profileUC.DeleteServicePackage(c.Request.Context(), userUUID, pkgID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"status": "deleted"})
}

func (h *ProfileHandler) GetLegalDocuments(c *gin.Context) {
	docType := c.Query("type")
	docs, err := h.profileUC.GetLegalDocuments(c.Request.Context(), docType)
	if err != nil {
		response.InternalError(c, "Failed to load legal documents")
		return
	}

	response.JSON(c, http.StatusOK, docs)
}
