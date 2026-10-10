// Package audio wraps the OS audio device and the lock-free rings that carry
// messages to the audio thread (milestone M5). It knows nothing about the
// language or the engine; cmd/loom wires engine.Render into a Device.
package audio

// RenderFunc fills out with interleaved stereo samples. It runs on the audio
// thread and must not allocate or block.
type RenderFunc func(out []float32)

// Device is an audio output backend (planned: malgo first, oto as a fallback).
type Device interface {
	Start(render RenderFunc) error
	Stop() error
	SampleRate() float64
}
