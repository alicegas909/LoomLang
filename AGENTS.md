# AGENTS.md

`loom` is a REPL-driven live-coding music language in Go. This repo is an
intentional scaffold: most packages are documented stubs. `README.md` is the
authoritative design/milestone doc; `docs/grammar.ebnf` is the parser spec.

## Current state (verify before assuming things work)

- Real code: `internal/lexer` (complete + tested), `internal/ast` (node types +
  `Dump` printer), `internal/parser` (Pratt expression parser +
  recursive-descent pattern sub-parser), `internal/pattern/rat.go`,
  `internal/eval` (value model + tree-walker for plain expressions), `internal/repl`
  (plain and readline front-ends), `cmd/loom`, `internal/sched/clock.go`
  (`FakeClock`), `internal/engine/units/registry.go`.
- Still stubs returning `ErrNotImplemented` / `TODO(M...)`: `internal/sched`,
  `internal/engine`, all `internal/engine/units/*` units, and
  graph/command/pattern combinators.
- M1 parser works: expressions, patterns, statements, synth defs and line
  continuations parse, with golden AST dumps. Syntax errors recover per
  statement: `Parse` returns the statements it built plus an `ErrorList` of
  `*SyntaxError`s, so one block reports every mistake. Do **not** apply the
  returned statements when the error is non-nil.
- M2 evaluator works for the plain-expression core only: numbers (notes become
  MIDI numbers), strings, variables, arithmetic, curried lambdas, `|>`/`>>` and
  application. `synth`, `channel`, `attach`, dotted paths and patterns return a
  `*RuntimeError` naming the milestone that implements them. `Eval` returns
  `eval.Result{Value, Commands}`; a failed submission is rolled back, and
  `parser.Incomplete` + `eval.ErrIncomplete` let the REPL keep reading an
  unfinished line. `repl.Run` is the testable io.Reader loop; `RunInteractive`
  adds readline history/editing. A submission of `quit` (see `repl.QuitWord`)
  ends the session; it is REPL-only, not a language keyword.
- `parser_test.go`, `FuzzParse`, `internal/eval/*_test.go` and
  `internal/repl/repl_test.go` are live.

## Commands

- Run `make fmt` (gofmt) after editing; stub-authored files occasionally drift.
- `make test` / `make race` / `make vet` / `make build` / `make run`
- Single test: `go test ./internal/lexer -run TestSpacingFlags -v`
- Fuzz: `make fuzz` (= `go test ./internal/parser -fuzz=FuzzParse -fuzztime=30s`);
  the lexer has `FuzzLex`. Fuzz targets must never panic or hang.
- Golden AST dumps: `go test ./internal/parser -run TestGolden`; regenerate after
  intended parser changes with `... -run TestGolden -update`.
- The only third-party dependency is `github.com/ergochat/readline` (M2). Add
  others only at the milestone listed in README (`malgo` M5, `wav` M6,
  MIDI/OSC M8).

## Architecture constraints

- Import arrows point one way (see README "Import rules"): the real-time side
  (`engine`) never imports the language side (`eval`, `parser`). `command` is the
  only eval↔runtime contract. Preserve this direction when adding imports.
- `internal/engine/units`: one file per unit; each self-registers via `Register`
  in `init()`, so a unit is dead unless its package is imported (blank import
  works). Adding a unit must not touch parser, eval, or engine.
- `eval.Env` is a chained scope; the global env persists across REPL submissions.

## Lexer conventions (easy to get wrong)

- The lexer emits single-character punctuation only. Multi-char operators
  (`|>`, `>>`, `<-`, `->`) are assembled by the parser from adjacent tokens.
- It tags every token with `SpaceBefore`/`SpaceAfter`; pattern modifiers are
  space-sensitive (`[a b]*2` vs `[a b] * 2`). The parser depends on these flags —
  do not "simplify" them away.
- `Kind` ordering matters: `kindCount` must stay last, and the lexer test asserts
  exactly 32 kinds. Update spec, names, and test together.
- Note literals (`c4`, `fs3`) are classified in the lexer, not the parser.

## Test fixtures

- `testdata/examples/*.loom` are parser fixtures, not runnable programs.
- `testdata/golden/*.golden` are `ast.DumpStmts` snapshots of those fixtures.

## Separate module

- `spikes/audiogc/` is an independent Go module (`make spike` runs it from its
  own dir). Never add its `malgo` dependency to the main module.
