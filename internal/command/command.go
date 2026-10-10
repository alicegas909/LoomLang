// Package command defines the messages the evaluator emits to change the live
// system. It is the only contract between the language (eval) and the runtime
// (sched, engine), which is what lets a bytecode VM or a different audio
// engine replace either side later.
package command

import (
	"github.com/you/loom/internal/engine/graph"
	"github.com/you/loom/internal/pattern"
)

// Command is one change to the running system. One REPL submission yields a
// []Command that is applied atomically or not at all.
type Command interface {
	isCommand()
}

// DefineSynth registers a synth recipe under a name.
type DefineSynth struct {
	Name  string
	Graph *graph.Template
}

// DefineChannel creates or updates the channel Name, playing synth Synth.
type DefineChannel struct {
	Name  string
	Synth string
}

// DefineBus creates or updates a bus with an FX chain.
type DefineBus struct {
	Name string
	FX   *graph.Template
}

// SetProp sets a static property, e.g. lead.gain = 0.8 or lead.send.echo = 0.3.
type SetProp struct {
	Path  []string
	Value float64
}

// SetFX replaces a channel's or bus's insert FX chain.
type SetFX struct {
	Target string
	FX     *graph.Template
}

// AttachPattern attaches a pattern to a channel's note stream (Param empty) or
// to one of its params. Quant is the swap quantization in cycles (0 = next cycle).
type AttachPattern struct {
	Target string
	Param  string
	Pat    pattern.Pattern
	Quant  pattern.Rat
}

// AttachSignal attaches a continuous signal graph to a param.
type AttachSignal struct {
	Target string
	Param  string
	Graph  *graph.Template
}

// SetTempo changes the tempo.
type SetTempo struct {
	BPM float64
}

// ControlOp is a mixer/transport control.
type ControlOp int

const (
	Hush ControlOp = iota
	Solo
	Mute
	Unmute
	Undo
)

// Control is hush / solo / mute / unmute / undo.
type Control struct {
	Op     ControlOp
	Target string // empty means all
}

func (DefineSynth) isCommand()   {}
func (DefineChannel) isCommand() {}
func (DefineBus) isCommand()     {}
func (SetProp) isCommand()       {}
func (SetFX) isCommand()         {}
func (AttachPattern) isCommand() {}
func (AttachSignal) isCommand()  {}
func (SetTempo) isCommand()      {}
func (Control) isCommand()       {}
