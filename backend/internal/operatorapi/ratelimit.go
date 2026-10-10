package operatorapi

import (
	"sync"
	"time"
)

// loginRateLimiter is a fixed-window failed-login limiter kept in-process: no
// new dependency and no Redis. It resets on process restart, and it is
// per-instance, so with several replicas the effective limit is multiplied.
// Move it to a shared store if you scale out.
//
// Two independent keys are tracked and both must pass: the attempted email
// (bounds guesses against one account) and the remote IP (bounds a single
// source's total volume regardless of which email it tries).
type loginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*window
	limit    int
	window   time.Duration
}

type window struct {
	count      int
	windowEnds time.Time
}

// newLoginRateLimiter constructs a limiter allowing limit failed attempts per
// key within windowDuration before further attempts are refused.
func newLoginRateLimiter(limit int, windowDuration time.Duration) *loginRateLimiter {
	return &loginRateLimiter{
		attempts: make(map[string]*window),
		limit:    limit,
		window:   windowDuration,
	}
}

// allow reports whether a new attempt for key is permitted right now, without
// recording anything. Check it before doing any password verification work.
func (l *loginRateLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	w, ok := l.attempts[key]
	if !ok || now.After(w.windowEnds) {
		return true
	}
	return w.count < l.limit
}

// recordFailure counts one failed attempt against key, starting a fresh window
// if none is active or the previous one expired.
func (l *loginRateLimiter) recordFailure(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	w, ok := l.attempts[key]
	if !ok || now.After(w.windowEnds) {
		l.attempts[key] = &window{count: 1, windowEnds: now.Add(l.window)}
		return
	}
	w.count++
}

// reset clears key's window; called on a successful login so a legitimate user
// who mistyped a few times is not left partially throttled.
func (l *loginRateLimiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}

// Defaults: generous enough that a real user mistyping a password a few times
// is never locked out, tight enough to make online brute-forcing impractical.
const (
	loginFailureLimit = 5
	loginRateWindow   = 15 * time.Minute
)
