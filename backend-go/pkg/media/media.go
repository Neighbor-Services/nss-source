package media

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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

	// 2. Fuzzy match in subfolder (e.g. searching for original filename when prefixed by chat_timestamp_)
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

	// Return relative media path (e.g., "portfolio/uuid.jpg" or "/media/portfolio/uuid.jpg")
	relPath := filepath.Join(subfolder, sanitizedName)
	return relPath, nil
}
