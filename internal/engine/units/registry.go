// Package units holds the DSP unit registry. Each unit lives in its own file
// and registers itself from init(). Every registered unit automatically becomes
// a language function; adding one never touches the parser, evaluator or engine.
//
// Packages that need the units must import this package (a blank import is
// enough) or the init() functions will not run.
package units

import "sort"

// Kind says how a unit is used in the language.
type Kind int

const (
	// Source units have defaults for every input, so a bare name is a Signal (saw).
	Source Kind = iota
	// Processor units take a signal as Inputs[0]; a bare name is a Signal -> Signal function (lpf).
	Processor
)

// Unit is one DSP block instance, with its own state.
type Unit interface {
	// Process renders one block. It must not allocate.
	Process(in [][]float32, out []float32)
}

// Input describes one named input of a unit.
type Input struct {
	Name    string
	Default float64
	Builtin string // e.g. "gate": a voice built-in used when no argument is given
}

// UnitDef describes a unit type.
type UnitDef struct {
	Name   string
	Kind   Kind
	Inputs []Input
	New    func(sampleRate float64) Unit
}

var registry = map[string]*UnitDef{}

// Register adds a unit definition. Call it from init().
func Register(d *UnitDef) {
	if _, dup := registry[d.Name]; dup {
		panic("units: duplicate registration of " + d.Name)
	}
	registry[d.Name] = d
}

// Lookup returns the definition registered under name.
func Lookup(name string) (*UnitDef, bool) {
	d, ok := registry[name]
	return d, ok
}

// Names returns all registered unit names, sorted.
func Names() []string {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
