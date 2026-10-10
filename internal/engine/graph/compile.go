package graph

// Compile (milestone M5): turn an evaluated synth body or FX function into a
// Template. It resolves unit names through the units registry, applies
// positional arguments, promotes numbers to constant signals, and turns
// + and * between signals into internal mix and gain units.
