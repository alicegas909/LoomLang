package eval

// Value is any runtime value. Planned types (see the design doc, section 3.4):
// Number, String, Pattern, Signal, Function, Synth, Channel, Bus, Command.
type Value interface {
	TypeName() string
}

// Number is the single numeric type. Notes are MIDI numbers.
type Number float64

// TypeName implements Value.
func (Number) TypeName() string { return "number" }
