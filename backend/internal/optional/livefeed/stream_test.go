package livefeed

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newStreamServer(t *testing.T, hub *Hub) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	identify := func(c *gin.Context) (string, bool) {
		id := c.GetHeader("X-Test-User")
		return id, id != ""
	}
	Register(r.Group("/api/v1"), func(c *gin.Context) { c.Next() }, identify, hub)
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

func TestStream_RelaysPublishedEvents(t *testing.T) {
	hub := NewHub()
	srv := newStreamServer(t, hub)

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/api/v1/events", nil)
	req.Header.Set("X-Test-User", "u1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content type = %q", ct)
	}

	deadline := time.Now().Add(2 * time.Second)
	for hub.SubscriberCount() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	hub.Publish(Event{Type: "item_created", Subject: "item-1", OccurredAt: time.Now()})

	scanner := bufio.NewScanner(resp.Body)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) == 2 {
			break
		}
	}
	if len(lines) < 2 || lines[0] != "event: item_created" || !strings.Contains(lines[1], `"subject":"item-1"`) {
		t.Fatalf("unexpected SSE frame: %v", lines)
	}
}

func TestStream_RejectsUnauthenticatedAndOverLimit(t *testing.T) {
	hub := NewHub()
	srv := newStreamServer(t, hub)

	anonReq, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/api/v1/events", nil)
	resp, err := http.DefaultClient.Do(anonReq)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}

	var open []*http.Response
	defer func() {
		for _, r := range open {
			_ = r.Body.Close()
		}
	}()
	for i := 0; i < connsPerUser; i++ {
		req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/api/v1/events", nil)
		req.Header.Set("X-Test-User", "u1")
		r, err := http.DefaultClient.Do(req) //nolint:bodyclose // closed by the deferred loop above
		if err != nil {
			t.Fatal(err)
		}
		open = append(open, r) //nolint:bodyclose // closed by the deferred loop above
	}
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/api/v1/events", nil)
	req.Header.Set("X-Test-User", "u1")
	over, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = over.Body.Close()
	if over.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", over.StatusCode)
	}
}
