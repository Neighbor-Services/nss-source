package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ProleadConfig defines connection settings for the Prolead Centralized Logger.
type ProleadConfig struct {
	ServerURL     string        // Default: http://localhost:8080
	APIKey        string        // e.g. "pl_live_..."
	Service       string        // e.g. "neighbor-services-backend"
	Environment   string        // "production", "development", "staging"
	BatchSize     int           // Default: 50
	FlushInterval time.Duration // Default: 2s
	Enabled       bool          // Whether remote streaming is active
}

// HTTPMetadata contains HTTP request context for the Prolead dashboard.
type HTTPMetadata struct {
	Method     string `json:"method,omitempty"`
	Path       string `json:"path,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
	DurationMs int64  `json:"duration_ms,omitempty"`
	ClientIP   string `json:"client_ip,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
}

// LogPayload represents a single structured log entry for the Prolead API.
type LogPayload struct {
	Timestamp   *time.Time             `json:"timestamp,omitempty"`
	Level       string                 `json:"level"`
	Service     string                 `json:"service,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Message     string                 `json:"message"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
	HTTP        *HTTPMetadata          `json:"http,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
}

// ProleadClient handles non-blocking buffered batch log transmission.
type ProleadClient struct {
	cfg        ProleadConfig
	bufferCh   chan *LogPayload
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	httpClient *http.Client
}

// NewProleadClient initializes and starts the background batch worker.
func NewProleadClient(cfg ProleadConfig) *ProleadClient {
	if cfg.ServerURL == "" {
		cfg.ServerURL = os.Getenv("PROLEAD_LOGGER_URL")
		if cfg.ServerURL == "" {
			cfg.ServerURL = "https://logger.proleadsolutions.co"
		}
	}
	cfg.ServerURL = strings.TrimRight(cfg.ServerURL, "/")

	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("PROLEAD_API_KEY")
	}
	if cfg.Service == "" {
		cfg.Service = "neighbor-services-backend"
	}
	if cfg.Environment == "" {
		cfg.Environment = os.Getenv("ENV")
		if cfg.Environment == "" {
			cfg.Environment = "development"
		}
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 50
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 2 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &ProleadClient{
		cfg:        cfg,
		bufferCh:   make(chan *LogPayload, 5000),
		ctx:        ctx,
		cancel:     cancel,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	c.wg.Add(1)
	go c.batchFlusher()
	return c
}

// Log queues a log payload asynchronously. Safe for high-concurrency use.
func (c *ProleadClient) Log(payload *LogPayload) {
	if c == nil || payload == nil {
		return
	}
	if payload.Service == "" {
		payload.Service = c.cfg.Service
	}
	if payload.Environment == "" {
		payload.Environment = c.cfg.Environment
	}
	if payload.Timestamp == nil {
		now := time.Now().UTC()
		payload.Timestamp = &now
	}

	payload.Message = MaskPII(payload.Message)
	if payload.StackTrace != "" {
		payload.StackTrace = MaskPII(payload.StackTrace)
	}

	select {
	case c.bufferCh <- payload:
	default:
		// Saturated queue: drop gracefully to avoid blocking host operations
	}
}

// Close flushes buffered logs and terminates background workers.
func (c *ProleadClient) Close() {
	if c == nil {
		return
	}
	c.cancel()
	c.wg.Wait()
}

func (c *ProleadClient) batchFlusher() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.cfg.FlushInterval)
	defer ticker.Stop()

	batch := make([]*LogPayload, 0, c.cfg.BatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		toSend := make([]*LogPayload, len(batch))
		copy(toSend, batch)
		batch = batch[:0]
		c.wg.Add(1)
		go func(items []*LogPayload) {
			defer c.wg.Done()
			c.sendBatch(items)
		}(toSend)
	}

	for {
		select {
		case <-c.ctx.Done():
			for {
				select {
				case item := <-c.bufferCh:
					batch = append(batch, item)
					if len(batch) >= c.cfg.BatchSize {
						flush()
					}
				default:
					flush()
					return
				}
			}
		case item := <-c.bufferCh:
			batch = append(batch, item)
			if len(batch) >= c.cfg.BatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (c *ProleadClient) sendBatch(batch []*LogPayload) {
	data, err := json.Marshal(batch)
	if err != nil {
		return
	}

	endpoint := fmt.Sprintf("%s/api/v1/ingest", c.cfg.ServerURL)
	const maxRetries = 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(data))
		if err != nil {
			return
		}

		req.Header.Set("Content-Type", "application/json")
		if c.cfg.APIKey != "" {
			req.Header.Set("X-API-Key", c.cfg.APIKey)
		}

		resp, err := c.httpClient.Do(req)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode < 500 {
				return
			}
		}

		if attempt < maxRetries-1 {
			time.Sleep(time.Duration((attempt+1)*100) * time.Millisecond)
		}
	}
}

// ProleadSlogHandler implements slog.Handler to dispatch logs to Prolead.
type ProleadSlogHandler struct {
	client   *ProleadClient
	minLevel slog.Level
	attrs    []slog.Attr
}

func NewProleadSlogHandler(client *ProleadClient, minLevel slog.Level) *ProleadSlogHandler {
	return &ProleadSlogHandler{
		client:   client,
		minLevel: minLevel,
	}
}

func (h *ProleadSlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func (h *ProleadSlogHandler) Handle(_ context.Context, r slog.Record) error {
	var lvl string
	switch {
	case r.Level >= slog.LevelError:
		lvl = "ERROR"
	case r.Level >= slog.LevelWarn:
		lvl = "WARN"
	case r.Level >= slog.LevelInfo:
		lvl = "INFO"
	default:
		lvl = "DEBUG"
	}

	ctxData := make(map[string]interface{})
	for _, a := range h.attrs {
		ctxData[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		ctxData[a.Key] = a.Value.Any()
		return true
	})

	ts := r.Time.UTC()
	h.client.Log(&LogPayload{
		Timestamp: &ts,
		Level:     lvl,
		Message:   r.Message,
		Context:   ctxData,
	})
	return nil
}

func (h *ProleadSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ProleadSlogHandler{
		client:   h.client,
		minLevel: h.minLevel,
		attrs:    newAttrs,
	}
}

func (h *ProleadSlogHandler) WithGroup(name string) slog.Handler {
	return h
}

// MultiHandler dispatches records to multiple slog.Handlers.
type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) slog.Handler {
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			_ = h.Handle(ctx, r)
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: handlers}
}
