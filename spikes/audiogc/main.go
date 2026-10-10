// Spike: does Go's garbage collector glitch a real-time audio callback?
//
// This is its own module so its audio dependency never touches the main one.
// To run the real thing:
//
//	cd spikes/audiogc
//	go get github.com/gen2brain/malgo
//
// Plan (about 50 lines):
//  1. Open a playback device with malgo: 48 kHz, stereo, float32, 256-frame period.
//  2. In the data callback, render 8 detuned saw voices through a one-pole low-pass.
//     No allocations in the callback.
//  3. In another goroutine, allocate aggressively (e.g. 1 MB slices in a loop).
//  4. Count callbacks that arrive late or underrun, for 60 seconds.
//  5. Repeat with period sizes 128, 256 and 512 and with GOGC tuned.
//
// Pass: zero audible glitches at a 256-frame period. Fail: consider the
// two-process design (Go language + separate audio engine).
package main

import "fmt"

func main() {
	fmt.Println("audio/GC spike: not implemented yet, see the plan in main.go")
}
