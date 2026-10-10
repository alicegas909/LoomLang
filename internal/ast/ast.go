// Package ast defines the syntax tree (milestone M1).
//
// Three node families: Stmt, Expr and Pat (pattern mode). A Splice node
// (a Pat holding an Expr) bridges pattern mode back to expression mode.
// Every node keeps its source Span so errors can point at the right place.
//
// Nodes to add in M1:
//
//	Stmt: Assign, Attach, ExprStmt (with optional @N quantize prefix)
//	Expr: Number, Note, String, Ident, Unary, Binary, Apply, Field, Lambda,
//	      SynthDef, PatternExpr (a Pat plus tight modifiers)
//	Pat:  Seq, Alt, Stack, Atom, Rest, Splice, Fast, Slow, Repeat, Weight,
//	      Degrade, Euclid
package ast

import "github.com/you/loom/internal/lexer"

// Span is a source range.
type Span struct {
	Start, End lexer.Pos
}

// Node is implemented by every AST node.
type Node interface {
	Span() Span
}

// Stmt is a top-level statement.
type Stmt interface {
	Node
	stmtNode()
}

// Expr is an expression-mode node.
type Expr interface {
	Node
	exprNode()
}

// Pat is a pattern-mode node.
type Pat interface {
	Node
	patNode()
}
