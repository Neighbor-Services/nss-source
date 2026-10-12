package logger

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProleadClientStreaming(t *testing.T) {
	var receivedBatch []*LogPayload
	var mu sync.Mutex

	// Mock Prolead Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/ingest", r.URL.Path)
		assert.Equal(t, "test_api_key_123", r.Header.Get("X-API-Key"))

		var batch []*LogPayload
		_ = json.NewDecoder(r.Body).Decode(&batch)

		mu.Lock()
		receivedBatch = append(receivedBatch, batch...)
		mu.Unlock()

		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"accepted"}`))
	}))
	defer server.Close()

	cfg := ProleadConfig{
		ServerURL:     server.URL,
		APIKey:        "test_api_key_123",
		Service:       "ns-test",
		Environment:   "testing",
		BatchSize:     2,
		FlushInterval: 50 * time.Millisecond,
		Enabled:       true,
	}

	client := NewProleadClient(cfg)
	assert.NotNil(t, client)

	// Send logs with PII
	client.Log(&LogPayload{
		Level:   "INFO",
		Message: "User contact@neighbor.com registered",
	})
	client.Log(&LogPayload{
		Level:   "ERROR",
		Message: "Failed payment for 123-456-7890",
	})

	// Wait for batch flush
	time.Sleep(150 * time.Millisecond)
	client.Close()

	mu.Lock()
	defer mu.Unlock()

	assert.Len(t, receivedBatch, 2)
	assert.Equal(t, "User c***@neighbor.com registered", receivedBatch[0].Message)
	assert.Equal(t, "Failed payment for ***-***-****", receivedBatch[1].Message)
}

func TestMultiHandlerWithSlog(t *testing.T) {
	var receivedCount int
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var batch []*LogPayload
		_ = json.NewDecoder(r.Body).Decode(&batch)
		mu.Lock()
		receivedCount += len(batch)
		mu.Unlock()
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	l := InitGlobalLogger("debug", ProleadConfig{
		ServerURL:     server.URL,
		APIKey:        "key_multi_test",
		Service:       "test-service",
		Environment:   "test",
		BatchSize:     1,
		FlushInterval: 20 * time.Millisecond,
		Enabled:       true,
	})
	assert.NotNil(t, l)

	Info("multi handler test event", "user", "admin@neighbor.com")

	time.Sleep(100 * time.Millisecond)
	Close()

	mu.Lock()
	defer mu.Unlock()
	assert.GreaterOrEqual(t, receivedCount, 1)
}
