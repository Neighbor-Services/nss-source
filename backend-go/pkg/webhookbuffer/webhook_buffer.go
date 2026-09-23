package webhookbuffer

import (
	"sync"
	"sync/atomic"
	"time"
)

// WebhookEvent represents a single captured webhook payload for admin inspection.
type WebhookEvent struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`       // "stripe", "checkr", "apple", "google"
	EventType   string    `json:"event_type"`   // e.g. "payment_intent.succeeded"
	Status      string    `json:"status"`       // "OK", "ERROR"
	PayloadSnip string    `json:"payload_snip"` // First 500 chars of raw body
	ErrorMsg    string    `json:"error_msg,omitempty"`
	ReceivedAt  time.Time `json:"received_at"`
}

const maxEvents = 200

// GlobalBuffer is the singleton ring buffer accessible from any handler.
var GlobalBuffer = &RingBuffer{
	events: make([]WebhookEvent, 0, maxEvents),
}

// RingBuffer is a thread-safe fixed-size event ring buffer.
type RingBuffer struct {
	mu     sync.RWMutex
	events []WebhookEvent
	counter int64
}

// NextID returns a monotonically increasing integer ID for new events.
func (r *RingBuffer) NextID() int64 {
	return atomic.AddInt64(&r.counter, 1)
}

// Push adds a new event. If at capacity, oldest is dropped.
func (r *RingBuffer) Push(evt WebhookEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if evt.ReceivedAt.IsZero() {
		evt.ReceivedAt = time.Now().UTC()
	}
	if len(evt.PayloadSnip) > 500 {
		evt.PayloadSnip = evt.PayloadSnip[:500] + "..."
	}

	if len(r.events) >= maxEvents {
		// Drop oldest
		r.events = r.events[1:]
	}
	r.events = append(r.events, evt)
}

// GetAll returns a copy of all events, newest first.
func (r *RingBuffer) GetAll() []WebhookEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]WebhookEvent, len(r.events))
	for i, j := len(r.events)-1, 0; i >= 0; i, j = i-1, j+1 {
		result[j] = r.events[i]
	}
	return result
}
