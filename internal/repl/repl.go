// Package repl is the interactive loop (milestone M2): line editing, history,
// multi-line input and error display. Enter submits when the statement is
// complete (balanced brackets, no trailing operator).
package repl

import (
	"context"
	"errors"
	"io"

	"github.com/you/loom/internal/eval"
)

// REPL reads submissions and sends them to an evaluator.
type REPL struct {
	ev eval.Evaluator
}

// New returns a REPL backed by ev.
func New(ev eval.Evaluator) *REPL {
	return &REPL{ev: ev}
}

// Run reads from in until EOF or ctx is cancelled. TODO(M2).
func (r *REPL) Run(ctx context.Context, in io.Reader, out io.Writer) error {
	return errors.New("repl: not implemented yet (milestone M2)")
}
