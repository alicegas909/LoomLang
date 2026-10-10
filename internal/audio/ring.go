package audio

// Ring (milestone M5): a single-producer single-consumer lock-free ring of
// fixed-size, pointer-free structs, built on sync/atomic head/tail indices.
// The scheduler pushes engine.NoteEvent values; the audio callback pops them.
// Capacity is a power of two so indexing is a mask, not a modulo.
