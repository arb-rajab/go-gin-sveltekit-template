package livefeed

import (
	"testing"
	"time"
)

func TestHub_PublishDeliversToSubscriber(t *testing.T) {
	hub := NewHub()
	events, unsubscribe := hub.Subscribe()
	defer unsubscribe()

	want := Event{Type: "item_updated", Subject: "item-1", OccurredAt: time.Now()}
	hub.Publish(want)

	select {
	case got := <-events:
		if got.Type != want.Type || got.Subject != want.Subject {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("subscriber never received the published event")
	}
}

func TestHub_PublishFansOutToEverySubscriber(t *testing.T) {
	hub := NewHub()
	eventsA, unsubA := hub.Subscribe()
	defer unsubA()
	eventsB, unsubB := hub.Subscribe()
	defer unsubB()

	hub.Publish(Event{Type: "item_created", Subject: "item-2"})

	for _, ch := range []<-chan Event{eventsA, eventsB} {
		select {
		case got := <-ch:
			if got.Type != "item_created" {
				t.Fatalf("got %+v", got)
			}
		case <-time.After(time.Second):
			t.Fatal("a subscriber never received the fan-out event")
		}
	}
}

func TestHub_UnsubscribeStopsDeliveryAndClosesChannel(t *testing.T) {
	hub := NewHub()
	events, unsubscribe := hub.Subscribe()
	if got := hub.SubscriberCount(); got != 1 {
		t.Fatalf("expected 1 subscriber, got %d", got)
	}

	unsubscribe()
	if got := hub.SubscriberCount(); got != 0 {
		t.Fatalf("expected 0 subscribers after unsubscribe, got %d", got)
	}

	hub.Publish(Event{Type: "item_updated", Subject: "item-3"})

	if _, ok := <-events; ok {
		t.Fatal("expected channel to be closed with no further events after unsubscribe")
	}
}

func TestHub_SlowSubscriberDropsRatherThanBlocksPublish(t *testing.T) {
	hub := NewHub()
	events, unsubscribe := hub.Subscribe()
	defer unsubscribe()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < subscriberBuffer+10; i++ {
			hub.Publish(Event{Type: "item_updated", Subject: "item-4"})
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Publish blocked on a full subscriber buffer instead of dropping")
	}

	// The subscriber still works; it just did not receive every event.
	drained := 0
	for {
		select {
		case <-events:
			drained++
		default:
			if drained == 0 {
				t.Fatal("expected at least some events to have been delivered before the buffer filled")
			}
			return
		}
	}
}
