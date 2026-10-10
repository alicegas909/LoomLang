// Package graph compiles synth bodies and FX functions into immutable
// templates (milestone M5). A Template is built once at definition time and
// instantiated per voice from a preallocated pool.
package graph

// Template is an immutable compiled signal graph: units in topological order,
// constants folded, buffers pre-assigned. Feedback loops are rejected in v0.1.
type Template struct {
	Nodes []Node
}

// Node is one unit in a template. TODO(M5): inputs, edges, buffer indices.
type Node struct {
	Unit string
}
