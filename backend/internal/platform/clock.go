package platform

import "time"

// Clock abstracts time so tests can control it.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

// RealClock returns the actual system time.
type RealClock struct{}

func (RealClock) Now() time.Time                         { return time.Now() }
func (RealClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// FakeClock returns a fixed time for tests.
type FakeClock struct {
	now time.Time
}

func NewFakeClock(t time.Time) *FakeClock { return &FakeClock{now: t} }

func (c *FakeClock) Now() time.Time { return c.now }
func (c *FakeClock) After(d time.Duration) <-chan time.Time {
	c.now = c.now.Add(d)
	ch := make(chan time.Time, 1)
	ch <- c.now
	return ch
}

// Advance moves the fake clock forward.
func (c *FakeClock) Advance(d time.Duration) { c.now = c.now.Add(d) }
