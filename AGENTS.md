# AGENTS.md

`loom` is a REPL-driven live-coding music language in Go. This repo is an
intentional scaffold: most packages are documented stubs. `README.md` is the
authoritative design/milestone doc; `docs/grammar.ebnf` is the parser spec.

## Current state (verify before assuming things work)

- Real code lives only in `internal/lexer` (complete + tested), `internal/pattern/rat.go`,
  `internal/eval/env.go`, `internal/sched/clock.go` (`FakeClock`), and
  `internal/engine/units/registry.go`.
- `internal/parser`, `internal/eval`, `internal/sched`, `internal/repl`,
  `internal/engine`, all `internal/engine/units/*` units, and graph/command/pattern
  combinators are stubs returning `ErrNotImplemented` / marked `TODO(M...)`.
- `parser_test.go` and `FuzzParse` are `t.Skip`ped until M1. `go test ./...`
  "passing" does **not** mean the parser works.
- The parser (M1) is the next milestone. It must build on the existing lexer.

## Commands

- First run after editing stubs: `make fmt` (gofmt) — stub files predate the
  compiler; `internal/pattern/pattern.go` is not currently gofmt-clean.
- `make test` / `make race` / `make vet` / `make build` / `make run`
- Single test: `go test ./internal/lexer -run TestSpacingFlags -v`
- Fuzz: `make fuzz` (= `go test ./internal/parser -fuzz=FuzzParse -fuzztime=30s`);
  the lexer has `FuzzLex`. Fuzz targets must never panic or hang.
- Go 1.22, **no third-party dependencies yet**. Add them only at the milestone
  listed in README (`readline` M2, `malgo` M5, `wav` M6, MIDI/OSC M8).

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
- `testdata/golden/` is for expected AST dumps (only `.gitkeep` so far).

## Separate module

- `spikes/audiogc/` is an independent Go module (`make spike` runs it from its
  own dir). Never add its `malgo` dependency to the main module.
