package engine

// NoteEvent is what crosses from the scheduler to the audio thread: a fixed-size,
// pointer-free struct, so it can travel through a lock-free ring.
type NoteEvent struct {
	Frame   uint64  // absolute sample-clock time
	Channel uint16  // channel index
	Value   float32 // MIDI note (unused for words)
	Word    uint32  // interned sample name id
	Dur     uint32  // frames until the gate goes low
	Vel     float32
}
