package livefeed

import "sync"

// connLimiter caps how many concurrent event streams a single authenticated
// user may hold open, plus a global ceiling across all users. It is
// in-process: a single instance never needs a shared limiter, and a restart
// severs every held connection anyway. Without it, session auth alone lets
// one user open unbounded streams, each its own goroutine, heartbeat ticker
// and hub subscriber slot.
type connLimiter struct {
	mu          sync.Mutex
	perUser     map[string]int
	total       int
	perUserMax  int
	globalLimit int
}

func newConnLimiter(perUserMax, globalLimit int) *connLimiter {
	return &connLimiter{
		perUser:     make(map[string]int),
		perUserMax:  perUserMax,
		globalLimit: globalLimit,
	}
}

// acquire reserves one slot for userID and reports whether it was granted.
// A caller that receives true must call release(userID) exactly once.
func (l *connLimiter) acquire(userID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.total >= l.globalLimit || l.perUser[userID] >= l.perUserMax {
		return false
	}
	l.perUser[userID]++
	l.total++
	return true
}

func (l *connLimiter) release(userID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.perUser[userID]--
	if l.perUser[userID] <= 0 {
		delete(l.perUser, userID)
	}
	l.total--
}

// Defaults: generous enough for one user with several open tabs, tight enough
// that one misbehaving session cannot exhaust the process's connection budget.
const (
	connsPerUser = 4
	connsGlobal  = 50
)
