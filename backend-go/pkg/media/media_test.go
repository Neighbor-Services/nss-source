package media_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-go/pkg/media"
	"github.com/stretchr/testify/assert"
)

func createMultipartFile(filename string, content []byte) (*multipart.FileHeader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(content); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if err := req.ParseMultipartForm(10 << 20); err != nil {
		return nil, err
	}

	files := req.MultipartForm.File["file"]
	if len(files) == 0 {
		return nil, http.ErrMissingFile
	}
	return files[0], nil
}

func TestValidateAndSaveUploadedFile_ValidPNG(t *testing.T) {
	// Minimal valid PNG header bytes
	pngBytes := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D}
	fileHeader, err := createMultipartFile("avatar.png", pngBytes)
	assert.NoError(t, err)

	relPath, err := media.ValidateAndSaveUploadedFile(fileHeader, "test_uploads", 5*1024*1024)
	assert.NoError(t, err)
	assert.NotEmpty(t, relPath)
	assert.Contains(t, relPath, "test_uploads")
	assert.Contains(t, relPath, ".png")
}

func TestValidateAndSaveUploadedFile_InvalidExtension(t *testing.T) {
	fileHeader, err := createMultipartFile("script.sh", []byte("#!/bin/bash\necho hello"))
	assert.NoError(t, err)

	_, err = media.ValidateAndSaveUploadedFile(fileHeader, "test_uploads", 5*1024*1024)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid file extension")
}

func TestValidateAndSaveUploadedFile_Oversized(t *testing.T) {
	largeBytes := make([]byte, 2048)
	fileHeader, err := createMultipartFile("image.jpg", largeBytes)
	assert.NoError(t, err)

	// Set maxBytes to 1024 bytes
	_, err = media.ValidateAndSaveUploadedFile(fileHeader, "test_uploads", 1024)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}
