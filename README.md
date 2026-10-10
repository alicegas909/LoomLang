# loom

A REPL-driven language for live coding music, written in Go. *`loom` is a working title.*

You define synths and mixer channels, write patterns in a built-in mini-notation,
and attach patterns to channels while the music keeps playing.

```text
pluck = synth {
  osc = saw * 0.5
  out = (osc >> lpf 1200 0.3) * adsr 0.01 0.3 0.4 0.5
}
lead = channel pluck
lead <- [a2 c3 e3 g3]*2 |> every 4 rev
```

**Status:** scaffold for the audio side, with the language front end filling in.
M0 (10 example programs), M1 (lexer, parser, AST printer, fuzzing, golden dumps,
per-statement error recovery) and M2 (curried tree-walking evaluator for plain
expressions, plus a readline REPL) are done; `synth`, `channel`, `attach` and
patterns still report a "not yet" error naming their milestone.

## Layout

| Path | Role | Milestone |
|---|---|---|
| `cmd/loom` | REPL entry point and composition root | M2 |
| `internal/lexer` | tokens with spacing flags | M1 |
| `internal/ast` | Stmt / Expr / Pat nodes | M1 |
| `internal/parser` | Pratt parser plus pattern-mode sub-parser | M1 |
| `internal/command` | messages from language to runtime | M2 |
| `internal/eval` | values, environments, built-ins, evaluator | M2 |
| `internal/pattern` | rational time, events, combinators, mini-notation | M3 |
| `internal/sched` | clock and lookahead scheduler | M4 |
| `internal/engine` | channels, buses, voices, mixer (real-time) | M5-M6 |
| `internal/engine/graph` | synth/FX graph compiler and templates | M5 |
| `internal/engine/units` | unit registry, one file per unit | M5-M6 |
| `internal/audio` | audio device, lock-free rings | M5 |
| `internal/repl` | line editing, history, error display | M2 |
| `internal/midi`, `internal/osc` | extra output backends | M8 |
| `spikes/audiogc` | separate module: does Go's GC glitch audio? | before M5 |
| `testdata/examples` | example programs, used as parser fixtures | M0-M1 |
| `docs/` | grammar and design doc | M0 |

## Import rules

Arrows point one way only. The real-time side never imports the language side.

```text
lexer <- ast <- parser
pattern, units, audio     leaf packages (no internal imports)
graph   -> units
command -> graph, pattern
eval    -> parser, ast, command, pattern, units
sched   -> command, pattern
engine  -> command, graph, units      (never eval or parser)
repl    -> eval
cmd/loom -> everything
```

## Develop

```text
make test     # go test ./...
make race     # with the race detector
make vet
make fmt      # run once: stubs were written without a compiler
make run
```

The only third-party dependency so far is readline (M2). Add the rest when you
reach their milestones:

```text
M2  github.com/ergochat/readline   # done
M5  go get github.com/gen2brain/malgo
M6  go get github.com/go-audio/wav
M8  go get gitlab.com/gomidi/midi/v2 github.com/hypebeast/go-osc
```

## Milestones

- [x] M0 Freeze the spec; grow `testdata/examples` to 10 programs
- [x] M1 Lexer (done), parser, AST printer (golden tests, fuzzing)
- [x] M2 Evaluator and REPL for plain expressions
- [ ] M3 Pattern engine, mini-notation, `show`
- [ ] M4 Scheduler and stub engine on a fake clock
- [ ] M5 First sound: audio device, rings, registry, graph compiler, voices
- [ ] M6 Starter units, FX chains, buses, sends, master
- [ ] M7 Live ergonomics: rollback, undo, hush/solo/mute, `@N`
- [ ] M8 Samples, MIDI/OSC, then decide on a bytecode VM

## Renaming

`loom` and `github.com/you/loom` are placeholders:

```text
grep -rl 'github.com/you/loom' . | xargs sed -i 's#github.com/you/loom#github.com/<you>/<name>#g'
git mv cmd/loom cmd/<name>
```
