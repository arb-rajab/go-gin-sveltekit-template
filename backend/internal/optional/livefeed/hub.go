// Package livefeed is an optional, in-process pub/sub hub that fans events out
// to connected browser clients over Server-Sent Events. It holds no database
// connection and makes no write of its own: publish an event only after the
// write it describes has committed.
//
// Delete this directory (and the two livefeed lines in main.go) if you do not
// need live updates. Note that the hub is per-process: with more than one
// backend replica, a client only sees events published by the replica it is
// connected to. Put a broker (Redis pub/sub, NATS) behind Publish if you scale
// out.
package livefeed

import (
	"sync"
	"time"
)

// Event is the one payload shape the hub broadcasts, serialized as SSE `data:`
// JSON. Type becomes the SSE event name; Subject identifies what changed;
// Data carries anything else the client needs.
type Event struct {
	Type       string         `json:"type"`
	Subject    string         `json:"subject"`
	OccurredAt time.Time      `json:"occurred_at"`
	Data       map[string]any `json:"data,omitempty"`
}

// subscriberBuffer is how many undelivered events a slow subscriber can
// accumulate before Publish starts dropping for it. This is a live-view
// convenience, not a durable queue: a client that misses an event still has
// the correct picture on its next page load or reconnect.
const subscriberBuffer = 16

// Hub is the fan-out point: any number of subscribers, each its own buffered
// channel. The zero value is not usable; construct with NewHub.
type Hub struct {
	mu          sync.Mutex
	subscribers map[chan Event]struct{}
}

// NewHub constructs an empty Hub, ready for Subscribe/Publish.
func NewHub() *Hub {
	return &Hub{subscribers: make(map[chan Event]struct{})}
}

// Subscribe registers a new subscriber and returns its event channel and an
// unsubscribe function the caller must call exactly once (typically via
// defer) when done reading. Only unsubscribe closes the channel, so a range
// over it blocks until then, never spuriously.
func (h *Hub) Subscribe() (events <-chan Event, unsubscribe func()) {
	ch := make(chan Event, subscriberBuffer)
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	unsub := func() {
		once.Do(func() {
			h.mu.Lock()
			delete(h.subscribers, ch)
			h.mu.Unlock()
			close(ch)
		})
	}
	return ch, unsub
}

// Publish fans an event out to every currently subscribed channel. Delivery is
// best-effort and non-blocking per subscriber: a subscriber whose buffer is
// full simply does not receive this event, rather than Publish blocking the
// caller on a slow or stalled connection.
func (h *Hub) Publish(event Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subscribers {
		select {
		case ch <- event:
		default:
			// Slow subscriber: drop for it, not for anyone else.
		}
	}
}

// SubscriberCount reports how many subscribers are currently registered.
func (h *Hub) SubscriberCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.subscribers)
}
