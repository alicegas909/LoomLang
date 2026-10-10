// Package sched is the clock and lookahead scheduler (milestone M4).
package sched

import "sync/atomic"

// Clock is the master time source. In production it is the audio engine's
// sample counter (engine.Engine satisfies this interface); in tests it is a FakeClock.
type Clock interface {
	// Frames returns the number of audio frames rendered so far.
	Frames() uint64
	// SampleRate returns frames per second.
	SampleRate() float64
}

// FakeClock is a manually advanced Clock for deterministic tests.
type FakeClock struct {
	rate   float64
	frames atomic.Uint64
}

// NewFakeClock returns a FakeClock at frame 0.
func NewFakeClock(sampleRate float64) *FakeClock {
	return &FakeClock{rate: sampleRate}
}

// Frames implements Clock.
func (c *FakeClock) Frames() uint64 { return c.frames.Load() }

// SampleRate implements Clock.
func (c *FakeClock) SampleRate() float64 { return c.rate }

// Advance moves the clock forward by n frames.
func (c *FakeClock) Advance(n uint64) { c.frames.Add(n) }
