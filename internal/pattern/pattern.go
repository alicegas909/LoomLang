// Package pattern is the pattern engine (milestone M3): exact rational time,
// events, and the Pattern interface. Patterns are pure values; querying one has
// no side effects, and randomness is a pure function of (cycle, position, seed).
package pattern

// Span is a time range in cycles. One cycle = 1.
type Span struct {
	Start, End Rat
}

// Value is the payload of an event: a number (MIDI note or param value) or a word (sample name).
type Value struct {
	Num    float64
	Word   string
	IsWord bool
}

// Event is one occurrence of a value.
type Event struct {
	Whole Span // the event's full extent
	Part  Span // the portion inside the query span
	Value Value
}

// Pattern is a pure function from a time span to events.
type Pattern interface {
	Query(q Span) []Event
}
