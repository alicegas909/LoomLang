package eval

import (
	"strconv"
	"strings"

	"github.com/you/loom/internal/ast"
)

// Value is any runtime value. M2 implements the plain-expression core:
// Number, String and Function. Pattern, Signal, Synth, Channel, Bus and
// Command values arrive with the pattern engine (M3) and the engine (M5-M6).
type Value interface {
	TypeName() string
}

// Number is the single numeric type. Notes are MIDI numbers.
type Number float64

// TypeName implements Value.
func (Number) TypeName() string { return "number" }

// String is a text value.
type String string

// TypeName implements Value.
func (String) TypeName() string { return "string" }

// Function is a user closure produced by a lambda. Params are the parameters
// still to be bound; Env is the scope the lambda captured.
type Function struct {
	Params []string
	Body   ast.Expr
	Env    *Env
}

// TypeName implements Value.
func (*Function) TypeName() string { return "function" }

// Builtin is a primitive function implemented in Go. Arity is the number of
// arguments it still expects; Fn is called with exactly Arity of them.
type Builtin struct {
	Name  string
	Arity int
	Fn    func(args []Value) (Value, error)
}

// TypeName implements Value.
func (*Builtin) TypeName() string { return "builtin" }

// Display renders a value the way the REPL prints it.
func Display(v Value) string {
	switch x := v.(type) {
	case nil:
		return ""
	case Number:
		return strconv.FormatFloat(float64(x), 'g', -1, 64)
	case String:
		return strconv.Quote(string(x))
	case *Function:
		return "<function>"
	case *Builtin:
		return "<builtin " + x.Name + ">"
	default:
		return "<" + v.TypeName() + ">"
	}
}

// NoteToMIDI converts a note literal such as c4, fs3 or bb2 to a MIDI number
// (c4 = 60, a4 = 69). It reports false if name is not a note.
func NoteToMIDI(name string) (float64, bool) {
	if name == "" {
		return 0, false
	}
	base := map[byte]int{'c': 0, 'd': 2, 'e': 4, 'f': 5, 'g': 7, 'a': 9, 'b': 11}
	letter := name[0]
	if letter >= 'A' && letter <= 'G' {
		letter += 'a' - 'A'
	}
	semi, ok := base[letter]
	if !ok {
		return 0, false
	}
	rest := name[1:]
	switch {
	case strings.HasPrefix(rest, "s"):
		semi++
		rest = rest[1:]
	case strings.HasPrefix(rest, "f"):
		semi--
		rest = rest[1:]
	}
	if rest == "" {
		return 0, false
	}
	oct, err := strconv.Atoi(rest)
	if err != nil {
		return 0, false
	}
	return float64(12*(oct+1) + semi), true
}
