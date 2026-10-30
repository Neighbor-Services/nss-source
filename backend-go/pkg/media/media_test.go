package media_test

import (
	"bytes"
	"encoding/binary"
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

// buildSyntheticMP4 constructs a minimal valid MP4 byte stream with an mvhd duration.
func buildSyntheticMP4(timescale uint32, duration uint32) []byte {
	buf := &bytes.Buffer{}

	// 1. ftyp box
	ftypPayload := []byte("isom\x00\x00\x02\x00isommp41")
	ftypSize := uint32(8 + len(ftypPayload))
	binary.Write(buf, binary.BigEndian, ftypSize)
	buf.WriteString("ftyp")
	buf.Write(ftypPayload)

	// 2. mvhd box
	mvhdBuf := &bytes.Buffer{}
	mvhdBuf.WriteByte(0)                    // version 0
	mvhdBuf.Write([]byte{0, 0, 0})          // flags
	mvhdBuf.Write(make([]byte, 8))          // creation & modification time (8 bytes)
	binary.Write(mvhdBuf, binary.BigEndian, timescale)
	binary.Write(mvhdBuf, binary.BigEndian, duration)
	mvhdBuf.Write(make([]byte, 80))         // rest of mvhd fields (rate, volume, matrix, etc.)

	mvhdSize := uint32(8 + mvhdBuf.Len())
	moovPayload := &bytes.Buffer{}
	binary.Write(moovPayload, binary.BigEndian, mvhdSize)
	moovPayload.WriteString("mvhd")
	moovPayload.Write(mvhdBuf.Bytes())

	// 3. moov box
	moovSize := uint32(8 + moovPayload.Len())
	binary.Write(buf, binary.BigEndian, moovSize)
	buf.WriteString("moov")
	buf.Write(moovPayload.Bytes())

	return buf.Bytes()
}

func TestValidateAndSaveUploadedFile_ValidPNG(t *testing.T) {
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

	_, err = media.ValidateAndSaveUploadedFile(fileHeader, "test_uploads", 1024)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

func TestDetectMP4Duration_Valid20Seconds(t *testing.T) {
	mp4Bytes := buildSyntheticMP4(1000, 20000) // 20.0s
	reader := bytes.NewReader(mp4Bytes)

	duration, err := media.DetectMP4Duration(reader)
	assert.NoError(t, err)
	assert.InDelta(t, 20.0, duration, 0.01)

	fileHeader, err := createMultipartFile("job_sample.mp4", mp4Bytes)
	assert.NoError(t, err)

	relPath, err := media.ValidateAndSaveUploadedFile(fileHeader, "request_videos", 50*1024*1024)
	assert.NoError(t, err)
	assert.NotEmpty(t, relPath)
	assert.Contains(t, relPath, "request_videos")
}

func TestDetectMP4Duration_Oversized45SecondsRejected(t *testing.T) {
	mp4Bytes := buildSyntheticMP4(1000, 45000) // 45.0s > 30s limit
	reader := bytes.NewReader(mp4Bytes)

	duration, err := media.DetectMP4Duration(reader)
	assert.NoError(t, err)
	assert.InDelta(t, 45.0, duration, 0.01)

	fileHeader, err := createMultipartFile("long_job.mp4", mp4Bytes)
	assert.NoError(t, err)

	_, err = media.ValidateAndSaveUploadedFile(fileHeader, "request_videos", 50*1024*1024)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "video duration exceeds 30-second limit")
}
