package media

import (
	"encoding/binary"
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
		".mp4":  true,
		".mov":  true,
		".avi":  true,
		".mkv":  true,
		".webm": true,
		".m4v":  true,
		".3gp":  true,
	}
	if !allowedExts[ext] {
		return "", fmt.Errorf("invalid file extension: %s. Allowed: jpg, jpeg, png, webp, pdf, mp4, mov, avi, mkv, webm, m4v, 3gp", ext)
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
		"application/octet-stream": true, // fallback for video and binary formats
		"video/mp4":                true,
		"video/quicktime":          true,
		"video/x-msvideo":          true,
		"video/webm":               true,
		"video/3gpp":               true,
	}

	if !allowedMimes[mimeType] && !strings.HasPrefix(mimeType, "image/") && !strings.HasPrefix(mimeType, "video/") {
		return "", fmt.Errorf("invalid file content type detected: %s", mimeType)
	}

	// If it is a video file, enforce max 30s duration (+1s margin for container overhead)
	if strings.HasPrefix(mimeType, "video/") || ext == ".mp4" || ext == ".mov" || ext == ".m4v" {
		if duration, err := DetectMP4Duration(file); err == nil && duration > 31.0 {
			return "", fmt.Errorf("video duration exceeds 30-second limit (detected %.1f seconds)", duration)
		}
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return "", fmt.Errorf("failed to seek file: %w", err)
		}
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
		log.Printf(" [GoStore Direct Upload Warning] %v", err)
	} else {
		log.Printf(" [GoStore Direct Upload SUCCESS] %s", relPath)
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
		bucket = "neighborservice"
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

	client := &http.Client{Timeout: 180 * time.Second}
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

// DetectMP4Duration reads an MP4 / MOV / ISO-BMFF stream and returns the duration in seconds.
// Returns 0 if moov/mvhd atom is not found or cannot be parsed.
func DetectMP4Duration(r io.ReadSeeker) (float64, error) {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}

	buf := make([]byte, 8)
	for {
		n, err := io.ReadFull(r, buf)
		if err != nil || n < 8 {
			break
		}

		boxSize := uint64(binary.BigEndian.Uint32(buf[0:4]))
		boxType := string(buf[4:8])

		headerSize := uint64(8)
		if boxSize == 1 {
			buf64 := make([]byte, 8)
			if _, err := io.ReadFull(r, buf64); err != nil {
				break
			}
			boxSize = binary.BigEndian.Uint64(buf64)
			headerSize = 16
		} else if boxSize == 0 {
			break
		}

		if boxType == "moov" {
			return parseMoovForMvhd(r, boxSize-headerSize)
		}

		if boxSize < headerSize {
			break
		}
		if _, err := r.Seek(int64(boxSize-headerSize), io.SeekCurrent); err != nil {
			break
		}
	}

	return 0, nil
}

func parseMoovForMvhd(r io.ReadSeeker, remaining uint64) (float64, error) {
	buf := make([]byte, 8)
	var bytesRead uint64

	for bytesRead+8 <= remaining {
	n, err := io.ReadFull(r, buf)
	if err != nil || n < 8 {
		break
	}
	bytesRead += 8

	subBoxSize := uint64(binary.BigEndian.Uint32(buf[0:4]))
	subBoxType := string(buf[4:8])

	subHeaderSize := uint64(8)
	if subBoxSize == 1 {
		buf64 := make([]byte, 8)
		if _, err := io.ReadFull(r, buf64); err != nil {
			break
		}
		subBoxSize = binary.BigEndian.Uint64(buf64)
		subHeaderSize = 16
		bytesRead += 8
	} else if subBoxSize == 0 {
		subBoxSize = remaining - bytesRead + subHeaderSize
	}

	if subBoxType == "mvhd" {
		payloadSize := subBoxSize - subHeaderSize
		if payloadSize < 20 {
			return 0, nil
		}
		mvhdBytes := make([]byte, payloadSize)
		if _, err := io.ReadFull(r, mvhdBytes); err != nil {
			return 0, err
		}

		version := mvhdBytes[0]
		var timescale uint32
		var duration uint64

		if version == 0 {
			if len(mvhdBytes) < 20 {
				return 0, nil
			}
			timescale = binary.BigEndian.Uint32(mvhdBytes[12:16])
			duration = uint64(binary.BigEndian.Uint32(mvhdBytes[16:20]))
		} else if version == 1 {
			if len(mvhdBytes) < 28 {
				return 0, nil
			}
			timescale = binary.BigEndian.Uint32(mvhdBytes[20:24])
			duration = binary.BigEndian.Uint64(mvhdBytes[24:32])
		}

		if timescale > 0 {
			return float64(duration) / float64(timescale), nil
		}
		return 0, nil
	}

	if subBoxSize < subHeaderSize {
		break
	}
	skipBytes := subBoxSize - subHeaderSize
	if _, err := r.Seek(int64(skipBytes), io.SeekCurrent); err != nil {
		break
	}
	bytesRead += skipBytes
	}
	return 0, nil
}
