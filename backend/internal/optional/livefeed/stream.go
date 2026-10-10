package livefeed

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// heartbeatInterval bounds how long an idle SSE connection goes without any
// bytes on the wire. Comment-only frames are not dispatched as events; they
// keep intermediaries (browsers, reverse proxies) from timing out a quiet
// stream.
//
// writeTimeout bounds how long any single write may take before the
// connection is considered stalled and torn down, which frees its limiter
// slot. A healthy client only needs the write to reach the kernel send
// buffer. Both are vars so tests can shrink them.
var (
	heartbeatInterval = 20 * time.Second
	writeTimeout      = 15 * time.Second
)

// Identify extracts the authenticated user's id from a request. Return false
// when the route is not behind your auth middleware.
type Identify func(c *gin.Context) (string, bool)

// Register mounts GET <group>/events, an SSE stream of the hub's events.
// Pass your session middleware as auth and a function reading the verified
// user id it stored (operatorapi.OperatorIDFrom).
func Register(group *gin.RouterGroup, auth gin.HandlerFunc, identify Identify, hub *Hub) {
	group.GET("/events", auth, StreamHandler(hub, identify))
}

// StreamHandler serves one SSE stream. It never decides what counts as an
// event; it formats and relays whatever was published to the hub.
func StreamHandler(hub *Hub, identify Identify) gin.HandlerFunc {
	limiter := newConnLimiter(connsPerUser, connsGlobal)
	return func(c *gin.Context) {
		userID, ok := identify(c)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !limiter.acquire(userID) {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		defer limiter.release(userID)

		events, unsubscribe := hub.Subscribe()
		defer unsubscribe()

		rc := http.NewResponseController(c.Writer)

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no") // disable proxy buffering of the stream
		if !setWriteDeadline(rc) {
			return
		}
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Flush()

		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()

		ctx := c.Request.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-events:
				if !ok {
					return
				}
				if !writeEvent(c, rc, event) {
					return
				}
			case <-ticker.C:
				if !setWriteDeadline(rc) {
					return
				}
				if _, err := fmt.Fprint(c.Writer, ": ping\n\n"); err != nil {
					return
				}
				c.Writer.Flush()
			}
		}
	}
}

// setWriteDeadline arms writeTimeout on the connection ahead of the write that
// follows. It reports false only if the connection refuses deadlines (a
// non-conforming ResponseWriter such as httptest.NewRecorder), so such a
// writer fails closed instead of silently never timing out.
func setWriteDeadline(rc *http.ResponseController) bool {
	return rc.SetWriteDeadline(time.Now().Add(writeTimeout)) == nil
}

// writeEvent formats one Event as an SSE frame and flushes it immediately.
func writeEvent(c *gin.Context, rc *http.ResponseController, event Event) bool {
	payload, err := json.Marshal(event)
	if err != nil {
		// Our own data failed to marshal: skip the event, keep the stream.
		return true
	}
	if !setWriteDeadline(rc) {
		return false
	}
	if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event.Type, payload); err != nil {
		return false
	}
	c.Writer.Flush()
	return true
}
