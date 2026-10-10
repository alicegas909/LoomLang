package engine

// Voice (milestone M5): one instance of a synth graph, preallocated in a pool.
// A note event takes a free voice or steals the oldest. The gate goes low after
// Dur frames; the voice is freed when its envelope finishes.
