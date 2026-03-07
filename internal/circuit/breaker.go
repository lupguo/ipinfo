package circuit

import (
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // failing fast
	StateHalfOpen              // allowing one probe
)

// Breaker is a simple per-provider circuit breaker.
type Breaker struct {
	mu           sync.Mutex
	state        State
	failures     int
	maxFailures  int
	resetTimeout time.Duration
	openedAt     time.Time
	probing      bool // guard: only one probe at a time in HALF_OPEN
}

// New creates a Breaker with the given thresholds.
func New(maxFailures int, resetTimeout time.Duration) *Breaker {
	return &Breaker{
		state:        StateClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

// Allow reports whether a request may proceed.
// Returns false when the circuit is OPEN (and the reset window has not elapsed).
func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(b.openedAt) >= b.resetTimeout {
			// Transition to HALF_OPEN and allow one probe.
			b.state = StateHalfOpen
			b.probing = true
			return true
		}
		return false
	case StateHalfOpen:
		if b.probing {
			// Only one probe at a time.
			return false
		}
		return false
	}
	return false
}

// RecordSuccess records a successful call.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.failures = 0
	b.probing = false
	b.state = StateClosed
}

// RecordFailure records a failed call and potentially opens the circuit.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.probing = false
	b.failures++
	if b.failures >= b.maxFailures || b.state == StateHalfOpen {
		b.state = StateOpen
		b.openedAt = time.Now()
	}
}

// IsOpen returns true if the breaker is currently open (blocking traffic).
func (b *Breaker) IsOpen() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == StateOpen && time.Since(b.openedAt) >= b.resetTimeout {
		return false // about to become HALF_OPEN
	}
	return b.state == StateOpen
}
