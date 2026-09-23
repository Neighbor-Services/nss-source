package handler

import (
	"net/http"
	"os"
	"strings"
	"time"

	"backend-go/internal/domain/entity"
	domainUsecase "backend-go/internal/domain/usecase"
	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

type PublicHandler struct {
	publicUC domainUsecase.PublicUseCase
}

func NewPublicHandler(publicUC domainUsecase.PublicUseCase) *PublicHandler {
	return &PublicHandler{publicUC: publicUC}
}

func (h *PublicHandler) HealthCheck(c *gin.Context) {
	response.JSON(c, http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "neighbor-services-go",
		"version": "1.0.0",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *PublicHandler) ReadyCheck(c *gin.Context) {
	response.JSON(c, http.StatusOK, gin.H{
		"status":    "ready",
		"service":   "neighbor-services-go",
		"version":   "1.0.0",
		"database":  "connected",
		"timestamp": time.Now().UTC().Unix(),
	})
}

func (h *PublicHandler) GetCMSContent(c *gin.Context) {
	if h.publicUC == nil {
		response.JSON(c, http.StatusOK, gin.H{})
		return
	}

	content, err := h.publicUC.GetCMSContent(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to load CMS content")
		return
	}

	response.JSON(c, http.StatusOK, content)
}

func (h *PublicHandler) SubmitContactMessage(c *gin.Context) {
	var msg entity.ContactMessage

	// Handle both form POST and JSON body
	if c.ContentType() == "application/json" {
		if err := c.ShouldBindJSON(&msg); err != nil {
			response.BadRequest(c, "Invalid JSON payload")
			return
		}
	} else {
		msg.FirstName = c.PostForm("first_name")
		msg.LastName = c.PostForm("last_name")
		msg.Email = c.PostForm("email")
		msg.InquiryType = c.PostForm("inquiry_type")
		msg.Message = c.PostForm("message")
	}

	if msg.FirstName == "" || msg.Email == "" || msg.Message == "" {
		if strings.Contains(c.GetHeader("Accept"), "text/html") {
			c.Redirect(http.StatusSeeOther, "/contact/?error=missing_fields")
			return
		}
		response.BadRequest(c, "first_name, email, and message are required")
		return
	}

	if h.publicUC != nil {
		if err := h.publicUC.SubmitContactMessage(c.Request.Context(), &msg); err != nil {
			response.InternalError(c, "Failed to submit contact message")
			return
		}
	}

	if strings.Contains(c.GetHeader("Accept"), "text/html") {
		c.Redirect(http.StatusSeeOther, "/contact/?success=true")
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
		"status":  "success",
		"message": "Your message has been sent successfully. We will get back to you shortly.",
	})
}

func (h *PublicHandler) SubmitResolutionReport(c *gin.Context) {
	var report entity.ResolutionReport

	if c.ContentType() == "application/json" {
		if err := c.ShouldBindJSON(&report); err != nil {
			response.BadRequest(c, "Invalid resolution report payload")
			return
		}
	} else {
		report.Role = c.PostForm("role")
		report.IssueType = c.PostForm("issue_type")
		report.BookingRef = c.PostForm("booking_ref")
		report.OtherNeighbor = c.PostForm("other_neighbor")
		report.Description = c.PostForm("description")
		report.ExpectedOutcome = c.PostForm("expected_outcome")
	}

	if report.Role == "" || report.IssueType == "" || report.Description == "" {
		response.BadRequest(c, "role, issue_type, and description are required")
		return
	}

	if h.publicUC != nil {
		if err := h.publicUC.SubmitResolutionReport(c.Request.Context(), &report); err != nil {
			response.InternalError(c, "Failed to submit resolution report")
			return
		}
	}

	response.JSON(c, http.StatusOK, gin.H{
		"status":  "success",
		"message": "Resolution report submitted successfully.",
	})
}

func (h *PublicHandler) RenderStaticPage(pageName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Look in root directory or templates directory
		paths := []string{
			pageName + ".html",
			"../" + pageName + ".html",
			"backend/public_site/templates/public_site/" + pageName + ".html",
			"../backend/public_site/templates/public_site/" + pageName + ".html",
		}

		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				c.File(p)
				return
			}
		}

		// Fallback JSON if static file not found
		response.JSON(c, http.StatusOK, gin.H{
			"page":    pageName,
			"service": "Neighbor Service",
		})
	}
}
