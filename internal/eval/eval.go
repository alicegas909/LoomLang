// Package eval evaluates an AST into values and commands (milestone M2).
//
// The tree-walking evaluator comes first. A bytecode VM can later implement
// the same Evaluator interface without touching anything downstream.
//
// M2 covers the plain-expression core: numbers, notes (as MIDI numbers),
// strings, variables, arithmetic, lambdas and application (curried). Synths,
// channels, patterns and attaches report a "not yet" error pointing at the
// milestone that implements them.
package eval

import (
	"github.com/you/loom/internal/ast"
	"github.com/you/loom/internal/command"
	"github.com/you/loom/internal/lexer"
	"github.com/you/loom/internal/parser"
)

// Evaluator turns source text into a result. A submission is either applied
// completely or not at all: on error every binding it made is rolled back.
type Evaluator interface {
	Eval(src string) (Result, error)
}

// Result is the outcome of one submission: the value of its last expression
// (nil if none) plus the commands to apply to the running system.
type Result struct {
	Value    Value
	Commands []command.Command
}

// TreeWalker is the v0.1 evaluator.
type TreeWalker struct {
	env *Env
}

// NewTreeWalker returns an evaluator with an empty global environment.
func NewTreeWalker() *TreeWalker {
	return &TreeWalker{env: NewEnv(nil)}
}

// Eval implements Evaluator.
func (t *TreeWalker) Eval(src string) (Result, error) {
	stmts, err := parser.Parse(src)
	if err != nil {
		if parser.Incomplete(src) {
			return Result{}, ErrIncomplete
		}
		return Result{}, err
	}

	saved := make(map[string]Value, len(t.env.vars))
	for k, v := range t.env.vars {
		saved[k] = v
	}

	var res Result
	for _, stmt := range stmts {
		v, err := t.evalStmt(t.env, stmt)
		if err != nil {
			t.env.vars = saved
			return Result{}, err
		}
		if v != nil {
			res.Value = v
		}
	}
	return res, nil
}

func (t *TreeWalker) evalStmt(env *Env, s ast.Stmt) (Value, error) {
	switch n := s.(type) {
	case *ast.Assign:
		if len(n.Target) != 1 {
			return nil, &RuntimeError{n.Sp.Start, "dotted assignment is not supported until milestone M5"}
		}
		v, err := t.eval(env, n.Value)
		if err != nil {
			return nil, err
		}
		env.Set(n.Target[0], v)
		return nil, nil
	case *ast.ExprStmt:
		return t.eval(env, n.Value)
	case *ast.Attach:
		return nil, &RuntimeError{n.Sp.Start, "attach needs a running engine (milestone M5)"}
	default:
		return nil, &RuntimeError{s.Span().Start, "unsupported statement"}
	}
}

func (t *TreeWalker) eval(env *Env, e ast.Expr) (Value, error) {
	switch n := e.(type) {
	case *ast.Number:
		return Number(n.Value), nil
	case *ast.Note:
		f, ok := NoteToMIDI(n.Text)
		if !ok {
			return nil, &RuntimeError{n.Sp.Start, "not a note: " + n.Text}
		}
		return Number(f), nil
	case *ast.String:
		return String(n.Value), nil
	case *ast.Ident:
		v, ok := env.Get(n.Name)
		if !ok {
			return nil, &RuntimeError{n.Sp.Start, "undefined: " + n.Name}
		}
		return v, nil
	case *ast.Unary:
		x, err := t.eval(env, n.X)
		if err != nil {
			return nil, err
		}
		num, ok := x.(Number)
		if !ok {
			return nil, &RuntimeError{n.Sp.Start, "unary " + n.Op + " needs a number, got " + x.TypeName()}
		}
		return -num, nil
	case *ast.Binary:
		return t.evalBinary(env, n)
	case *ast.Lambda:
		return &Function{Params: n.Params, Body: n.Body, Env: env}, nil
	case *ast.Apply:
		fn, err := t.eval(env, n.Func)
		if err != nil {
			return nil, err
		}
		args := make([]Value, len(n.Args))
		for i, a := range n.Args {
			if args[i], err = t.eval(env, a); err != nil {
				return nil, err
			}
		}
		return t.apply(n.Sp.Start, fn, args)
	case *ast.Field:
		return nil, &RuntimeError{n.Sp.Start, "record fields are not supported yet"}
	case *ast.SynthDef:
		return nil, &RuntimeError{n.Sp.Start, "synth needs the graph compiler (milestone M5)"}
	case *ast.PatternExpr:
		return nil, &RuntimeError{n.Sp.Start, "patterns need the pattern engine (milestone M3)"}
	default:
		return nil, &RuntimeError{e.Span().Start, "unsupported expression"}
	}
}

func (t *TreeWalker) evalBinary(env *Env, n *ast.Binary) (Value, error) {
	if n.Op == "|>" || n.Op == ">>" {
		x, err := t.eval(env, n.X)
		if err != nil {
			return nil, err
		}
		y, err := t.eval(env, n.Y)
		if err != nil {
			return nil, err
		}
		if n.Op == "|>" {
			return t.apply(n.Sp.Start, y, []Value{x})
		}
		return t.compose(n.Sp.Start, x, y)
	}

	x, err := t.eval(env, n.X)
	if err != nil {
		return nil, err
	}
	y, err := t.eval(env, n.Y)
	if err != nil {
		return nil, err
	}
	xn, ok := x.(Number)
	if !ok {
		return nil, &RuntimeError{n.X.Span().Start, "operand of " + n.Op + " must be a number, got " + x.TypeName()}
	}
	yn, ok := y.(Number)
	if !ok {
		return nil, &RuntimeError{n.Y.Span().Start, "operand of " + n.Op + " must be a number, got " + y.TypeName()}
	}
	switch n.Op {
	case "+":
		return xn + yn, nil
	case "-":
		return xn - yn, nil
	case "*":
		return xn * yn, nil
	case "/":
		if yn == 0 {
			return nil, &RuntimeError{n.Sp.Start, "division by zero"}
		}
		return xn / yn, nil
	}
	return nil, &RuntimeError{n.Sp.Start, "unknown operator " + n.Op}
}

// apply calls fn with args, binding as many as it can. A function given fewer
// arguments than it expects returns a partially applied function (currying).
func (t *TreeWalker) apply(pos lexer.Pos, fn Value, args []Value) (Value, error) {
	for len(args) > 0 {
		switch f := fn.(type) {
		case *Function:
			n := len(f.Params)
			if n == 0 {
				return nil, &RuntimeError{pos, "cannot apply a function with no parameters"}
			}
			take := len(args)
			if take > n {
				take = n
			}
			callEnv := NewEnv(f.Env)
			for i := 0; i < take; i++ {
				callEnv.Set(f.Params[i], args[i])
			}
			if take < n {
				return &Function{Params: f.Params[take:], Body: f.Body, Env: callEnv}, nil
			}
			res, err := t.eval(callEnv, f.Body)
			if err != nil {
				return nil, err
			}
			if args = args[take:]; len(args) == 0 {
				return res, nil
			}
			fn = res
		case *Builtin:
			if f.Arity <= 0 {
				return nil, &RuntimeError{pos, "cannot apply " + f.Name}
			}
			if len(args) < f.Arity {
				captured := append([]Value(nil), args...)
				return &Builtin{
					Name:  f.Name,
					Arity: f.Arity - len(args),
					Fn: func(more []Value) (Value, error) {
						return f.Fn(append(captured, more...))
					},
				}, nil
			}
			res, err := f.Fn(args[:f.Arity])
			if err != nil {
				return nil, err
			}
			if args = args[f.Arity:]; len(args) == 0 {
				return res, nil
			}
			fn = res
		default:
			return nil, &RuntimeError{pos, "cannot apply a " + fn.TypeName()}
		}
	}
	return fn, nil
}

// compose builds g . f: a one-argument function that feeds its argument
// through f and then through g.
func (t *TreeWalker) compose(pos lexer.Pos, f, g Value) (Value, error) {
	if !isCallable(f) {
		return nil, &RuntimeError{pos, "left of >> must be a function, got " + f.TypeName()}
	}
	if !isCallable(g) {
		return nil, &RuntimeError{pos, "right of >> must be a function, got " + g.TypeName()}
	}
	return &Builtin{Name: ">>", Arity: 1, Fn: func(args []Value) (Value, error) {
		mid, err := t.apply(pos, f, args)
		if err != nil {
			return nil, err
		}
		return t.apply(pos, g, []Value{mid})
	}}, nil
}

func isCallable(v Value) bool {
	switch v.(type) {
	case *Function, *Builtin:
		return true
	}
	return false
}
