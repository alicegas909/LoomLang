// Package engine is the real-time audio engine (milestones M5-M6).
//
// Real-time rules, no exceptions: the audio callback never allocates, never
// takes a mutex, never blocks on a channel, and never calls the interpreter.
// All buffers and voices are preallocated. The engine receives commands and
// events through lock-free rings; it must not import eval or parser.
package engine

import "sync/atomic"

const (
	// BlockSize is the internal processing block in frames.
	BlockSize = 64
	// OutChannels is the number of interleaved output channels (stereo).
	OutChannels = 2
)

// Engine renders audio. Its frame counter is the master clock for the scheduler.
type Engine struct {
	sampleRate float64
	frames     atomic.Uint64
}

// New returns an engine for the given sample rate.
func New(sampleRate float64) *Engine {
	return &Engine{sampleRate: sampleRate}
}

// SampleRate returns frames per second.
func (e *Engine) SampleRate() float64 { return e.sampleRate }

// Frames returns frames rendered so far.
func (e *Engine) Frames() uint64 { return e.frames.Load() }

// Render fills out (interleaved stereo) and advances the clock. It is the
// function handed to the audio device. Currently renders silence. TODO(M5).
func (e *Engine) Render(out []float32) {
	for i := range out {
		out[i] = 0
	}
	e.frames.Add(uint64(len(out) / OutChannels))
}
