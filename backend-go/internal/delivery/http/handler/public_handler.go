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
	"gorm.io/gorm"
)

type PublicHandler struct {
	publicUC  domainUsecase.PublicUseCase
	db        *gorm.DB
	startTime time.Time
}

func NewPublicHandler(publicUC domainUsecase.PublicUseCase, db ...*gorm.DB) *PublicHandler {
	var gormDB *gorm.DB
	if len(db) > 0 {
		gormDB = db[0]
	}
	return &PublicHandler{
		publicUC:  publicUC,
		db:        gormDB,
		startTime: time.Now(),
	}
}

func (h *PublicHandler) HealthCheck(c *gin.Context) {
	dbStatus := "connected"
	status := "healthy"
	statusCode := http.StatusOK

	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "disconnected"
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}
	}

	uptime := int64(0)
	if !h.startTime.IsZero() {
		uptime = int64(time.Since(h.startTime).Seconds())
	}

	response.JSON(c, statusCode, gin.H{
 "status":         status,
 "database":       dbStatus,
 "service": "neighbor-services-go",
 "version": "1.0.0",
 "uptime_seconds": uptime,
 "time":           time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *PublicHandler) ReadyCheck(c *gin.Context) {
	dbStatus := "connected"
	status := "ready"
	statusCode := http.StatusOK

	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "disconnected"
			status = "not_ready"
			statusCode = http.StatusServiceUnavailable
		}
	}

	uptime := int64(0)
	if !h.startTime.IsZero() {
		uptime = int64(time.Since(h.startTime).Seconds())
	}

	response.JSON(c, statusCode, gin.H{
 "status":         status,
 "service": "neighbor-services-go",
 "version": "1.0.0",
 "database":       dbStatus,
 "uptime_seconds": uptime,
 "timestamp":      time.Now().UTC().Unix(),
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
 "status": "success",
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
 "status": "success",
 "message": "Resolution report submitted successfully.",
	})
}

func (h *PublicHandler) RenderStaticPage(pageName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Look in root directory or templates directory
		paths := []string{
			pageName + ".html",
 "../" + pageName + ".html",
 "templates/" + pageName + ".html",
 "templates/public_site/" + pageName + ".html",
 "/app/templates/" + pageName + ".html",
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
 "service": "Neighbor Services",
		})
	}
}

func (h *PublicHandler) AssetLinks(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, `[
  {
 "relation": ["delegate_permission/common.handle_all_urls"],
 "target": {
 "namespace": "android_app",
 "package_name": "com.neighborservicesolutionsllc.nsapp",
 "sha256_cert_fingerprints": [
 "A6:ED:80:4D:3E:A4:63:7D:EB:81:09:FA:72:E1:BC:EA:65:56:64:D5:7D:3F:A5:85:1D:55:D6:7D:F1:06:1E:07",
 "DF:A8:0E:3E:9D:BA:CC:31:C2:0F:34:9F:C3:06:E8:A6:79:22:AF:D1:CA:8C:56:0E:3D:7E:0F:7E:08:FD:19:2F"
      ]
    }
  }
]`)
}

func (h *PublicHandler) AppleAppSiteAssociation(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, `{
 "applinks": {
 "apps": [],
 "details": [
      {
 "appID": "875W85H2Q9.com.neighborservicesolutionsllc.nsapp",
 "paths": ["*"]
      }
    ]
  },
 "webcredentials": {
 "apps": [
 "875W85H2Q9.com.neighborservicesolutionsllc.nsapp"
    ]
  }
}`)
}
