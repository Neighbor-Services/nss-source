package media

import (
	"time"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// ResolveMediaDir dynamically resolves the media directory and ensures the subfolder exists.
// It checks environment variables MEDIA_ROOT, MEDIA_UPLOAD_DIR, then falls back to relative local paths.
func ResolveMediaDir(subfolder string) string {
	base := os.Getenv("MEDIA_ROOT")
	if base == "" {
		base = os.Getenv("MEDIA_UPLOAD_DIR")
	}
	if base == "" {
		// If monorepo structure ../backend/media exists, use it
		if fi, err := os.Stat("../backend/media"); err == nil && fi.IsDir() {
			base = "../backend/media"
		} else if fi, err := os.Stat("./media"); err == nil && fi.IsDir() {
			base = "./media"
		} else {
			base = "./media"
		}
	}

	target := filepath.Join(base, subfolder)
	_ = os.MkdirAll(target, 0755)
	return target
}

// GetBaseMediaDir returns the root media folder for static file serving.
func GetBaseMediaDir() string {
	base := os.Getenv("MEDIA_ROOT")
	if base == "" {
		base = os.Getenv("MEDIA_UPLOAD_DIR")
	}
	if base == "" {
		if fi, err := os.Stat("../backend/media"); err == nil && fi.IsDir() {
			base = "../backend/media"
		} else {
			base = "./media"
		}
	}
	_ = os.MkdirAll(base, 0755)
	return base
}

// FindMediaFile locates a file in the media directory, checking exact match first,
// then searching for filenames containing or suffixed with the requested filename.
func FindMediaFile(baseDir, relPath string) (string, bool) {
	cleanRel := filepath.Clean(strings.TrimPrefix(relPath, "/"))
	fullPath := filepath.Join(baseDir, cleanRel)

	// 1. Exact match
	if fi, err := os.Stat(fullPath); err == nil && !fi.IsDir() {
		return fullPath, true
	}

	// 2. If cleanRel starts with "media/", try without "media/"
	trimmedRel := strings.TrimPrefix(cleanRel, "media/")
	if trimmedRel != cleanRel {
		candidate := filepath.Join(baseDir, trimmedRel)
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
			return candidate, true
		}
	}

	// 3. Fuzzy match in subfolder (e.g. searching for original filename when prefixed by chat_timestamp_)
	dir := filepath.Dir(fullPath)
	targetBase := filepath.Base(fullPath)
	if entries, err := os.ReadDir(dir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				name := entry.Name()
				if strings.HasSuffix(name, targetBase) || strings.Contains(name, targetBase) {
					return filepath.Join(dir, name), true
				}
			}
		}
	}

	return "", false
}

// ValidateAndSaveUploadedFile enforces strict file size, magic-byte MIME detection,
// extension whitelisting, and writes to a sanitized UUID-based filename.
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
		"image/jpeg":      true,
		"image/png":       true,
		"image/webp":      true,
		"application/pdf": true,
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
	targetDir := ResolveMediaDir(subfolder)
	destPath := filepath.Join(targetDir, sanitizedName)

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

		// Sync with GoStore CAS engine asynchronously
	go func() {
		_ = SyncToGoStore(subfolder, sanitizedName, destPath, mimeType)
	}()

	relPath := filepath.Join(subfolder, sanitizedName)
	return relPath, nil
}

// SyncToGoStore uploads a local media file to the GoStore CAS object store.
func SyncToGoStore(subfolder, filename, filePath, contentType string) error {
	gostoreURL := os.Getenv("GOSTORE_URL")
	if gostoreURL == "" {
		if os.Getenv("GIN_MODE") == "release" || os.Getenv("ENV") == "production" || os.Getenv("APP_ENV") == "production" {
			gostoreURL = "https://file.proleadsolutions.co"
		} else {
			gostoreURL = "http://localhost:8080"
		}
	}

	apiKey := os.Getenv("GOSTORE_API_KEY")
	objectPath := fmt.Sprintf("%s/%s", subfolder, filename)

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	bucket := os.Getenv("GOSTORE_BUCKET")
	if bucket == "" {
		bucket = "neighborservice"
	}

	endpoint := fmt.Sprintf("%s/v0/b/%s/o?name=%s", strings.TrimRight(gostoreURL, "/"), url.PathEscape(bucket), strings.ReplaceAll(objectPath, "/", "%2F"))
	req, err := http.NewRequest(http.MethodPost, endpoint, f)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("gostore sync failed with status %d", resp.StatusCode)
	}
	return nil
}
