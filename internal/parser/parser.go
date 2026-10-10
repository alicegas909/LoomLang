// Package parser builds an AST from source (milestone M1).
//
// Expression mode uses a Pratt parser. Pattern mode (pattern.go) is a
// recursive-descent sub-parser. The two call each other at '[', '<', '(' and
// '$'. On a syntax error the parser skips to the end of the statement and
// reports line:col; it never panics and never half-applies a submission.
package parser

import (
	"errors"

	"github.com/you/loom/internal/ast"
)

// ErrNotImplemented is returned until M1 lands.
var ErrNotImplemented = errors.New("parser: not implemented yet (milestone M1)")

// Parse parses one submission (a REPL block or a whole file).
func Parse(src string) ([]ast.Stmt, error) {
	return nil, ErrNotImplemented
}
