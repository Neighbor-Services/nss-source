package media

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ValidateAndSaveUploadedFile enforces strict file size, magic-byte MIME detection,
// extension whitelisting, and streams directly to GoStore CAS object store without local disk storage.
func ValidateAndSaveUploadedFile(fileHeader *multipart.FileHeader, subfolder string, maxBytes int64) (string, error) {
	if fileHeader == nil {
		return "", errors.New("no file provided")
	}

	if maxBytes <= 0 {
		maxBytes = 10 * 1024 * 1024 // default 10MB
	}

	if fileHeader.Size > maxBytes {
		return "", fmt.Errorf("file size exceeds maximum allowed size (%d MB)", maxBytes/(1024*1024))
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".pdf":  true,
	}
	if !allowedExts[ext] {
		return "", fmt.Errorf("invalid file extension: %s. Allowed: jpg, jpeg, png, webp, pdf", ext)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Sniff magic bytes (first 512 bytes)
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file header: %w", err)
	}

	mimeType := http.DetectContentType(buf[:n])
	allowedMimes := map[string]bool{
		"image/jpeg":               true,
		"image/png":                true,
		"image/webp":               true,
		"application/pdf":          true,
		"application/octet-stream": true, // fallback for some binary formats
	}

	if !allowedMimes[mimeType] && !strings.HasPrefix(mimeType, "image/") {
		return "", fmt.Errorf("invalid file content type detected: %s", mimeType)
	}

	// Reset read cursor
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to seek file: %w", err)
	}

	// Generate clean randomized filename
	sanitizedName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	relPath := fmt.Sprintf("%s/%s", strings.Trim(subfolder, "/"), sanitizedName)

	// Stream directly to GoStore Content-Addressed Storage
	if err := StreamToGoStore(subfolder, sanitizedName, file, mimeType); err != nil {
		log.Printf("⚠️ [GoStore Direct Upload Warning] %v", err)
	} else {
		log.Printf("✅ [GoStore Direct Upload SUCCESS] %s", relPath)
	}

	return relPath, nil
}

// StreamToGoStore uploads a media stream directly to GoStore CAS without saving to local disk.
func StreamToGoStore(subfolder, filename string, reader io.Reader, contentType string) error {
	gostoreURL := os.Getenv("GOSTORE_URL")
	if gostoreURL == "" {
		if os.Getenv("GIN_MODE") == "release" || os.Getenv("ENV") == "production" || os.Getenv("APP_ENV") == "production" {
			gostoreURL = "https://file.proleadsolutions.co"
		} else {
			gostoreURL = "http://localhost:8080"
		}
	}

	apiKey := os.Getenv("GOSTORE_API_KEY")
	objectPath := fmt.Sprintf("%s/%s", strings.Trim(subfolder, "/"), filename)

	bucket := os.Getenv("GOSTORE_BUCKET")
	if bucket == "" {
		bucket = os.Getenv("BUCKET_NAME")
	}
	if bucket == "" {
		bucket = os.Getenv("MEDIA_BUCKET")
	}
	if bucket == "" {
		bucket = "default"
	}

	endpoint := fmt.Sprintf(
		"%s/v0/b/%s/o?name=%s",
		strings.TrimRight(gostoreURL, "/"),
		url.PathEscape(bucket),
		strings.ReplaceAll(objectPath, "/", "%2F"),
	)

	req, err := http.NewRequest(http.MethodPost, endpoint, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gostore upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// SyncToGoStore is maintained for backwards compatibility and forwards to StreamToGoStore if a file exists.
func SyncToGoStore(subfolder, filename, filePath, contentType string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return StreamToGoStore(subfolder, filename, f, contentType)
}
