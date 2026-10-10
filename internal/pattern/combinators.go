package pattern

// Combinators (milestone M3). Each is a Pattern -> Pattern wrapper:
//
//   fastcat, slowcat, stack, fast, slow, rev, every, shift,
//   degrade (seeded, deterministic), euclid, weighted sequences, pure.
//
// Mini-notation syntax desugars to these; see docs/grammar.ebnf.
