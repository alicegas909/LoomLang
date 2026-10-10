package parser

// Pattern-mode sub-parser (milestone M1): sequences, <alternation>, [stack, layers],
// modifiers (*, /, !, @, ?, euclid) and splices. Modifiers must touch their
// element; spacing only matters next to pattern literals (see docs/grammar.ebnf).
