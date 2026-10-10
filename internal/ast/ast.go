// Package ast defines the syntax tree (milestone M1).
//
// Three node families: Stmt, Expr and Pat (pattern mode). A Splice node
// (a Pat holding an Expr) bridges pattern mode back to expression mode.
// Every node keeps its source Span so errors can point at the right place.
package ast

import "github.com/you/loom/internal/lexer"

// Span is a source range. End is the position of the last token in the node.
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

// --- Statements ---

// Assign is `path = expr`.
type Assign struct {
	Sp     Span
	Target []string // a dotted path, e.g. ["lead", "gain"]
	Value  Expr
}

// Attach is `path <- expr`, optionally quantized by a leading `@NUMBER`.
type Attach struct {
	Sp       Span
	Target   []string
	Value    Expr
	Quant    float64
	HasQuant bool
}

// ExprStmt is a bare expression, optionally quantized by a leading `@NUMBER`.
type ExprStmt struct {
	Sp       Span
	Value    Expr
	Quant    float64
	HasQuant bool
}

func (*Assign) stmtNode()   {}
func (*Attach) stmtNode()   {}
func (*ExprStmt) stmtNode() {}

// --- Expressions ---

// Number is an unsigned numeric literal.
type Number struct {
	Sp    Span
	Text  string
	Value float64
}

// Note is a note literal such as c4 or fs3.
type Note struct {
	Sp   Span
	Text string
}

// String is a string literal; Value is the raw text without the quotes.
type String struct {
	Sp    Span
	Value string
}

// Ident is a name.
type Ident struct {
	Sp   Span
	Name string
}

// Unary is a prefix operator application (currently only `-`).
type Unary struct {
	Sp Span
	Op string
	X  Expr
}

// Binary is an infix operator: + - * / >> |>.
type Binary struct {
	Sp Span
	Op string
	X  Expr
	Y  Expr
}

// Apply is function application by juxtaposition: `f a b`.
type Apply struct {
	Sp   Span
	Func Expr
	Args []Expr
}

// Field is dotted access: `x.y`.
type Field struct {
	Sp   Span
	X    Expr
	Name string
}

// Lambda is `\x y -> body`.
type Lambda struct {
	Sp     Span
	Params []string
	Body   Expr
}

// SynthDef is `synth { ... }`.
type SynthDef struct {
	Sp    Span
	Stmts []SynthStmt
}

// SynthStmt is one binding in a synth body, optionally declared with `param`.
type SynthStmt struct {
	Sp    Span
	Param bool
	Name  string
	Value Expr
}

// PatternExpr is a pattern literal used in expression mode, e.g. `[a b]*2`.
type PatternExpr struct {
	Sp  Span
	Pat Pat
}

func (*Number) exprNode()      {}
func (*Note) exprNode()        {}
func (*String) exprNode()      {}
func (*Ident) exprNode()       {}
func (*Unary) exprNode()       {}
func (*Binary) exprNode()      {}
func (*Apply) exprNode()       {}
func (*Field) exprNode()       {}
func (*Lambda) exprNode()      {}
func (*SynthDef) exprNode()    {}
func (*PatternExpr) exprNode() {}

// --- Patterns (pattern mode) ---

// Seq is a space-separated sequence of steps.
type Seq struct {
	Sp    Span
	Steps []Pat
}

// Alt is an alternation `<a b c>`.
type Alt struct {
	Sp    Span
	Steps []Pat
}

// Stack is a comma-separated stack of layers `[a, b]`.
type Stack struct {
	Sp     Span
	Layers []Pat
}

// Atom is a single note, number or word in pattern mode.
type Atom struct {
	Sp   Span
	Kind lexer.Kind
	Text string
}

// Rest is `~`.
type Rest struct {
	Sp Span
}

// Splice is `$x` or `(expr)`: an expression spliced into a pattern.
type Splice struct {
	Sp   Span
	Expr Expr
}

// Fast is `pat * factor`.
type Fast struct {
	Sp     Span
	Pat    Pat
	Factor Factor
}

// Slow is `pat / factor`.
type Slow struct {
	Sp     Span
	Pat    Pat
	Factor Factor
}

// Repeat is `pat ! [count]`.
type Repeat struct {
	Sp    Span
	Pat   Pat
	Count int
}

// Weight is `pat @ amount`.
type Weight struct {
	Sp     Span
	Pat    Pat
	Amount float64
}

// Degrade is `pat ? [amount]`.
type Degrade struct {
	Sp     Span
	Pat    Pat
	Amount float64
}

// Euclid is `pat (steps, pulses [, rotate])`.
type Euclid struct {
	Sp            Span
	Pat           Pat
	Steps, Pulses int
	Rotate        int
}

func (*Seq) patNode()     {}
func (*Alt) patNode()     {}
func (*Stack) patNode()   {}
func (*Atom) patNode()    {}
func (*Rest) patNode()    {}
func (*Splice) patNode()  {}
func (*Fast) patNode()    {}
func (*Slow) patNode()    {}
func (*Repeat) patNode()  {}
func (*Weight) patNode()  {}
func (*Degrade) patNode() {}
func (*Euclid) patNode()  {}

// Factor is the argument of a Fast or Slow modifier. Exactly one of Expr and
// Pat is set: Expr for a number, `$x` or `(expr)`, Pat for `<sequence>`.
type Factor struct {
	Sp   Span
	Expr Expr
	Pat  Pat
}

// --- Span methods ---

func (n *Assign) Span() Span      { return n.Sp }
func (n *Attach) Span() Span      { return n.Sp }
func (n *ExprStmt) Span() Span    { return n.Sp }
func (n *Number) Span() Span      { return n.Sp }
func (n *Note) Span() Span        { return n.Sp }
func (n *String) Span() Span      { return n.Sp }
func (n *Ident) Span() Span       { return n.Sp }
func (n *Unary) Span() Span       { return n.Sp }
func (n *Binary) Span() Span      { return n.Sp }
func (n *Apply) Span() Span       { return n.Sp }
func (n *Field) Span() Span       { return n.Sp }
func (n *Lambda) Span() Span      { return n.Sp }
func (n *SynthDef) Span() Span    { return n.Sp }
func (n *PatternExpr) Span() Span { return n.Sp }
func (n *Seq) Span() Span         { return n.Sp }
func (n *Alt) Span() Span         { return n.Sp }
func (n *Stack) Span() Span       { return n.Sp }
func (n *Atom) Span() Span        { return n.Sp }
func (n *Rest) Span() Span        { return n.Sp }
func (n *Splice) Span() Span      { return n.Sp }
func (n *Fast) Span() Span        { return n.Sp }
func (n *Slow) Span() Span        { return n.Sp }
func (n *Repeat) Span() Span      { return n.Sp }
func (n *Weight) Span() Span      { return n.Sp }
func (n *Degrade) Span() Span     { return n.Sp }
func (n *Euclid) Span() Span      { return n.Sp }
func (n *Factor) Span() Span      { return n.Sp }
