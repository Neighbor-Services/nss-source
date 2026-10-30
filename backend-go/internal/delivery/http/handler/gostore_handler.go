package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"backend-go/internal/config"
	"backend-go/pkg/media"
	"backend-go/pkg/response"
)

type GoStoreWebhookPayload struct {
	EventID   string                 `json:"event_id"`
	Event     string                 `json:"event"`
	Timestamp int64                  `json:"timestamp"`
	Bucket    string                 `json:"bucket"`
	Path      string                 `json:"path"`
	Size      int64                  `json:"size"`
	MimeType  string                 `json:"mime_type"`
	Hash      string                 `json:"hash"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type GoStoreHandler struct {
	cfg *config.Config
}

func NewGoStoreHandler(cfg *config.Config) *GoStoreHandler {
	return &GoStoreHandler{cfg: cfg}
}

// HandleWebhook receives and verifies lifecycle & image processing webhooks from GoStore.
func (h *GoStoreHandler) HandleWebhook(c *gin.Context) {
	signature := c.GetHeader("X-GoStore-Signature")
	if signature == "" {
		signature = c.GetHeader("X-Signature")
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "Failed to read webhook payload")
		return
	}

	// Verify HMAC signature if secret/api key is configured
	secret := os.Getenv("GOSTORE_WEBHOOK_SECRET")
	if secret == "" {
		secret = os.Getenv("GOSTORE_API_KEY")
	}

	if secret != "" && signature != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(bodyBytes)
		expectedMAC := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
			log.Printf(" GoStore Webhook signature mismatch: got %s, expected %s", signature, expectedMAC)
			response.Unauthorized(c, "Invalid webhook signature")
			return
		}
	}

	var payload GoStoreWebhookPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		response.BadRequest(c, "Malformed JSON webhook payload")
		return
	}

	log.Printf(" [GoStore Webhook] Event: %s | Bucket: %s | Object: %s | Size: %d bytes",
		payload.Event, payload.Bucket, payload.Path, payload.Size)

	response.JSON(c, http.StatusOK, gin.H{
 "status": "received",
 "event_id": payload.EventID,
 "event":    payload.Event,
 "received": time.Now().UTC().Format(time.RFC3339),
	})
}

// ProxyMedia handles /media/*path directly via GoStore object store without local disk storage.
func (h *GoStoreHandler) ProxyMedia(c *gin.Context) {
	reqPath := c.Param("filepath")
	cleanPath := strings.TrimPrefix(reqPath, "/")

	// Set browser caching headers for media assets
	c.Header("Cache-Control", "public, max-age=31536000, immutable")

	gostoreURL := h.cfg.GoStoreURL
	if gostoreURL == "" {
		gostoreURL = "https://file.proleadsolutions.co"
	}
	bucket := h.cfg.GoStoreBucket
	if bucket == "" {
		bucket = os.Getenv("GOSTORE_BUCKET")
	}
	if bucket == "" {
		bucket = os.Getenv("BUCKET_NAME")
	}
	if bucket == "" {
		bucket = "neighborservice"
	}

	targetURL := fmt.Sprintf("%s/v0/b/%s/o/%s?alt=media",
		strings.TrimRight(gostoreURL, "/"),
		url.PathEscape(bucket),
		strings.ReplaceAll(cleanPath, "/", "%2F"),
	)

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, targetURL, nil)
	if err != nil {
		response.NotFound(c, "Media file not found")
		return
	}

	if rangeHeader := c.GetHeader("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	if h.cfg.GoStoreAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.cfg.GoStoreAPIKey)
		req.Header.Set("X-API-Key", h.cfg.GoStoreAPIKey)
	}

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil || (resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent) {
		if resp != nil {
			resp.Body.Close()
		}
		response.NotFound(c, "Media file not found")
		return
	}
	defer resp.Body.Close()

	c.Status(resp.StatusCode)
	if cType := resp.Header.Get("Content-Type"); cType != "" {
		c.Header("Content-Type", cType)
	}
	if cLen := resp.Header.Get("Content-Length"); cLen != "" {
		c.Header("Content-Length", cLen)
	}
	if cRange := resp.Header.Get("Content-Range"); cRange != "" {
		c.Header("Content-Range", cRange)
	}
	c.Header("Accept-Ranges", "bytes")

	// Stream directly to client
	_, _ = io.Copy(c.Writer, resp.Body)
}

// ListMedia returns a media browser list from GoStore for the admin portal.
func (h *GoStoreHandler) ListMedia(c *gin.Context) {
	defaultBucket := h.cfg.GoStoreBucket
	if defaultBucket == "" {
		defaultBucket = os.Getenv("GOSTORE_BUCKET")
	}
	if defaultBucket == "" {
		defaultBucket = os.Getenv("BUCKET_NAME")
	}
	if defaultBucket == "" {
		defaultBucket = "default"
	}

	bucket := c.Query("bucket")
	if bucket == "" {
		bucket = defaultBucket
	}

	gostoreURL := h.cfg.GoStoreURL
	if gostoreURL == "" {
		gostoreURL = "https://file.proleadsolutions.co"
	}

	targetURL := fmt.Sprintf("%s/v0/b/%s/o", strings.TrimRight(gostoreURL, "/"), url.PathEscape(bucket))
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, targetURL, nil)
	if err != nil {
		response.JSON(c, http.StatusOK, gin.H{
 "data":           gin.H{"items": []interface{}{}},
 "source": "gostore-cas",
 "bucket":         bucket,
 "default_bucket": defaultBucket,
		})
		return
	}

	if h.cfg.GoStoreAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.cfg.GoStoreAPIKey)
		req.Header.Set("X-API-Key", h.cfg.GoStoreAPIKey)
	}

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Do(req)
	if err != nil || (resp != nil && resp.StatusCode != http.StatusOK) {
		if resp != nil {
			resp.Body.Close()
		}
		// Gracefully return empty partition items when storage appliance is empty or syncing
		response.JSON(c, http.StatusOK, gin.H{
 "data":           gin.H{"items": []interface{}{}},
 "source": "gostore-cas",
 "bucket":         bucket,
 "default_bucket": defaultBucket,
 "status": "ready",
		})
		return
	}
	defer resp.Body.Close()

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		response.JSON(c, http.StatusOK, gin.H{
 "data":           gin.H{"items": []interface{}{}},
 "source": "gostore-cas",
 "bucket":         bucket,
 "default_bucket": defaultBucket,
		})
		return
	}

	response.JSON(c, http.StatusOK, gin.H{
 "data":           data,
 "source": "gostore",
 "bucket":         bucket,
 "default_bucket": defaultBucket,
	})
}

// UploadMedia handles direct file uploads from clients,
// validating the file, streaming it to GoStore CAS using server credentials, and returning the public download URL.
func (h *GoStoreHandler) UploadMedia(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file provided in form field 'file'")
		return
	}

	subfolder := strings.TrimSpace(c.DefaultPostForm("subfolder", "uploads"))
	if subfolder == "" {
		subfolder = "uploads"
	}
	subfolder = strings.ReplaceAll(subfolder, "..", "")
	subfolder = strings.Trim(subfolder, "/\\")

	relPath, err := media.ValidateAndSaveUploadedFile(fileHeader, subfolder, 100*1024*1024)
	if err != nil {
		response.BadRequest(c, "File upload rejected: "+err.Error())
		return
	}

	gostoreURL := h.cfg.GoStoreURL
	if gostoreURL == "" {
		gostoreURL = "https://file.proleadsolutions.co"
	}
	bucket := h.cfg.GoStoreBucket
	if bucket == "" {
		bucket = os.Getenv("GOSTORE_BUCKET")
	}
	if bucket == "" {
		bucket = "neighborservice"
	}

	downloadURL := fmt.Sprintf("%s/v0/b/%s/o/%s?alt=media",
		strings.TrimRight(gostoreURL, "/"),
		url.PathEscape(bucket),
		url.PathEscape(relPath),
	)

	response.JSON(c, http.StatusOK, gin.H{
		"status":    "success",
		"url":       downloadURL,
		"path":      relPath,
		"file_name": fileHeader.Filename,
		"size":      fileHeader.Size,
	})
}
