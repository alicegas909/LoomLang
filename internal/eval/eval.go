// Package eval evaluates an AST into commands (milestone M2).
//
// The tree-walking evaluator comes first. A bytecode VM can later implement
// the same Evaluator interface without touching anything downstream.
package eval

import (
	"errors"

	"github.com/you/loom/internal/command"
)

// Evaluator turns source text into commands. A submission is evaluated
// completely before anything is returned; on error nothing is applied.
type Evaluator interface {
	Eval(src string) ([]command.Command, error)
}

// TreeWalker is the v0.1 evaluator.
type TreeWalker struct {
	env *Env
}

// NewTreeWalker returns an evaluator with an empty global environment.
func NewTreeWalker() *TreeWalker {
	return &TreeWalker{env: NewEnv(nil)}
}

// Eval implements Evaluator. TODO(M2).
func (t *TreeWalker) Eval(src string) ([]command.Command, error) {
	return nil, errors.New("eval: not implemented yet (milestone M2)")
}
