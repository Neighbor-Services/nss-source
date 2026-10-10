package webhookbuffer

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRingBuffer_PushAndGetAll(t *testing.T) {
	buf := &RingBuffer{
		events: make([]WebhookEvent, 0, maxEvents),
	}

	for i := 1; i <= 5; i++ {
		buf.Push(WebhookEvent{
			ID:          fmt.Sprintf("evt_%d", i),
			Source:      "stripe",
			EventType:   "payment_intent.succeeded",
			Status:      "OK",
			PayloadSnip: fmt.Sprintf("payload_%d", i),
		})
	}

	events := buf.GetAll()
	assert.Len(t, events, 5)
	// Newest should be first
	assert.Equal(t, "evt_5", events[0].ID)
	assert.Equal(t, "evt_1", events[4].ID)
	assert.False(t, events[0].ReceivedAt.IsZero())
}

func TestRingBuffer_CapacityOverflow(t *testing.T) {
	buf := &RingBuffer{
		events: make([]WebhookEvent, 0, maxEvents),
	}

	for i := 1; i <= 250; i++ {
		buf.Push(WebhookEvent{
			ID:          fmt.Sprintf("evt_%d", i),
			Source:      "stripe",
			EventType:   "charge.succeeded",
			Status:      "OK",
			PayloadSnip: "payload",
		})
	}

	events := buf.GetAll()
	assert.Len(t, events, maxEvents)
	// Newest is 250, oldest kept is 51
	assert.Equal(t, "evt_250", events[0].ID)
	assert.Equal(t, fmt.Sprintf("evt_%d", 250-maxEvents+1), events[maxEvents-1].ID)
}

func TestRingBuffer_NextID(t *testing.T) {
	buf := &RingBuffer{}
	id1 := buf.NextID()
	id2 := buf.NextID()
	assert.Equal(t, int64(1), id1)
	assert.Equal(t, int64(2), id2)
}

func TestRingBuffer_PayloadTruncation(t *testing.T) {
	buf := &RingBuffer{}
	longPayload := make([]byte, 800)
	for i := range longPayload {
		longPayload[i] = 'A'
	}

	buf.Push(WebhookEvent{
		ID:          "long_evt",
		Source:      "checkr",
		EventType:   "report.completed",
		Status:      "OK",
		PayloadSnip: string(longPayload),
		ReceivedAt:  time.Now().UTC(),
	})

	events := buf.GetAll()
	assert.Len(t, events, 1)
	assert.True(t, len(events[0].PayloadSnip) <= 503)
	assert.Contains(t, events[0].PayloadSnip, "...")
}
