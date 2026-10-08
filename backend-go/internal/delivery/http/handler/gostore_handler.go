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
	"path/filepath"
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
			log.Printf("⚠️ GoStore Webhook signature mismatch: got %s, expected %s", signature, expectedMAC)
			response.Unauthorized(c, "Invalid webhook signature")
			return
		}
	}

	var payload GoStoreWebhookPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		response.BadRequest(c, "Malformed JSON webhook payload")
		return
	}

	log.Printf("✅ [GoStore Webhook] Event: %s | Bucket: %s | Object: %s | Size: %d bytes",
		payload.Event, payload.Bucket, payload.Path, payload.Size)

	response.JSON(c, http.StatusOK, gin.H{
		"status":   "received",
		"event_id": payload.EventID,
		"event":    payload.Event,
		"received": time.Now().UTC().Format(time.RFC3339),
	})
}

// ProxyMedia handles /media/*path. It serves locally if available or pulls from GoStore.
func (h *GoStoreHandler) ProxyMedia(c *gin.Context) {
	reqPath := c.Param("filepath")
	cleanPath := strings.TrimPrefix(reqPath, "/")

	baseDir := media.GetBaseMediaDir()
	fullLocalPath, found := media.FindMediaFile(baseDir, cleanPath)

	// Set browser caching headers for media assets
	c.Header("Cache-Control", "public, max-age=31536000, immutable")

	if found && fullLocalPath != "" {
		c.File(fullLocalPath)
		return
	}

	// File not on local disk; stream/proxy from GoStore object store
	gostoreURL := h.cfg.GoStoreURL
	if gostoreURL == "" {
		gostoreURL = "https://file.proleadsolutions.co"
	}
	bucket := h.cfg.GoStoreBucket
	if bucket == "" {
		bucket = "neighborservice"
	}

	// Try bucket-prefixed or raw path
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

	if h.cfg.GoStoreAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.cfg.GoStoreAPIKey)
		req.Header.Set("X-API-Key", h.cfg.GoStoreAPIKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		response.NotFound(c, "Media file not found")
		return
	}
	defer resp.Body.Close()

	// Asynchronously cache on local disk for subsequent rapid requests
	subfolder := filepath.Dir(cleanPath)
	filename := filepath.Base(cleanPath)
	targetDir := media.ResolveMediaDir(subfolder)
	cachePath := filepath.Join(targetDir, filename)

	if cType := resp.Header.Get("Content-Type"); cType != "" {
		c.Header("Content-Type", cType)
	}
	c.Header("Content-Length", resp.Header.Get("Content-Length"))

	// Stream to client and write to disk cache concurrently
	var writer io.Writer = c.Writer
	if cacheFile, err := os.Create(cachePath); err == nil {
		defer cacheFile.Close()
		writer = io.MultiWriter(c.Writer, cacheFile)
	}

	_, _ = io.Copy(writer, resp.Body)
}

// ListMedia returns a unified media browser list for the admin portal.
func (h *GoStoreHandler) ListMedia(c *gin.Context) {
	bucket := c.DefaultQuery("bucket", h.cfg.GoStoreBucket)
	if bucket == "" {
		bucket = "neighborservice"
	}

	gostoreURL := h.cfg.GoStoreURL
	if gostoreURL == "" {
		gostoreURL = "https://file.proleadsolutions.co"
	}

	targetURL := fmt.Sprintf("%s/v0/b/%s/o", strings.TrimRight(gostoreURL, "/"), url.PathEscape(bucket))
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, targetURL, nil)
	if err != nil {
		response.InternalError(c, "Failed to create upstream request")
		return
	}

	if h.cfg.GoStoreAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.cfg.GoStoreAPIKey)
		req.Header.Set("X-API-Key", h.cfg.GoStoreAPIKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		// Fallback: list local media directory
		baseDir := media.GetBaseMediaDir()
		var localFiles []gin.H
		_ = filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				rel, _ := filepath.Rel(baseDir, path)
				localFiles = append(localFiles, gin.H{
					"name":         rel,
					"size":         info.Size(),
					"updated":      info.ModTime().Format(time.RFC3339),
					"content_type": http.DetectContentType([]byte{}),
					"bucket":       bucket,
					"url":          "/media/" + rel,
				})
			}
			return nil
		})
		response.JSON(c, http.StatusOK, gin.H{"items": localFiles, "source": "local"})
		return
	}
	defer resp.Body.Close()

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		response.InternalError(c, "Failed to parse upstream response")
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"data": data, "source": "gostore"})
}
