package eval

import (
	"errors"
	"fmt"

	"github.com/you/loom/internal/lexer"
)

// ErrIncomplete is returned by Eval when src is not yet a complete submission
// (the input stops at a point where more text could still make it valid, e.g.
// an unclosed bracket or a trailing operator). The REPL keeps reading instead
// of reporting an error.
var ErrIncomplete = errors.New("incomplete input")

// RuntimeError is an evaluation error at a source position.
type RuntimeError struct {
	Pos lexer.Pos
	Msg string
}

func (e *RuntimeError) Error() string {
	return fmt.Sprintf("%d:%d: %s", e.Pos.Line, e.Pos.Col, e.Msg)
}
