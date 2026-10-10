package ast

import (
	"fmt"
	"strconv"
	"strings"
)

// Dump renders a node as an indented tree. Output is deterministic, so the
// golden tests can compare it byte for byte. Scalar fields sit on the node's
// line; child nodes are indented two spaces.
func Dump(n Node) string {
	var pr printer
	pr.node(n)
	return pr.b.String()
}

// DumpStmts renders every statement, separated by a blank line.
func DumpStmts(stmts []Stmt) string {
	var b strings.Builder
	for i, s := range stmts {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(Dump(s))
	}
	return b.String()
}

type printer struct {
	b     strings.Builder
	depth int
}

func (pr *printer) line(format string, args ...any) {
	for i := 0; i < pr.depth; i++ {
		pr.b.WriteString("  ")
	}
	fmt.Fprintf(&pr.b, format, args...)
	pr.b.WriteByte('\n')
}

func (pr *printer) child(n Node) {
	pr.depth++
	pr.node(n)
	pr.depth--
}

func (pr *printer) node(n Node) {
	switch n := n.(type) {
	// statements
	case *Assign:
		pr.line("Assign target=%s", strings.Join(n.Target, "."))
		pr.child(n.Value)
	case *Attach:
		if n.HasQuant {
			pr.line("Attach target=%s quant=%s", strings.Join(n.Target, "."), num(n.Quant))
		} else {
			pr.line("Attach target=%s", strings.Join(n.Target, "."))
		}
		pr.child(n.Value)
	case *ExprStmt:
		if n.HasQuant {
			pr.line("ExprStmt quant=%s", num(n.Quant))
		} else {
			pr.line("ExprStmt")
		}
		pr.child(n.Value)

	// expressions
	case *Number:
		pr.line("Number %s", n.Text)
	case *Note:
		pr.line("Note %s", n.Text)
	case *String:
		pr.line("String %q", n.Value)
	case *Ident:
		pr.line("Ident %s", n.Name)
	case *Unary:
		pr.line("Unary %s", n.Op)
		pr.child(n.X)
	case *Binary:
		pr.line("Binary %s", n.Op)
		pr.child(n.X)
		pr.child(n.Y)
	case *Apply:
		pr.line("Apply")
		pr.child(n.Func)
		for _, a := range n.Args {
			pr.child(a)
		}
	case *Field:
		pr.line("Field %s", n.Name)
		pr.child(n.X)
	case *Lambda:
		pr.line("Lambda (%s)", strings.Join(n.Params, " "))
		pr.child(n.Body)
	case *SynthDef:
		pr.line("SynthDef")
		pr.depth++
		for _, s := range n.Stmts {
			pr.synthStmt(s)
		}
		pr.depth--
	case *PatternExpr:
		pr.line("PatternExpr")
		pr.child(n.Pat)

	// patterns
	case *Seq:
		pr.line("Seq")
		pr.pats(n.Steps)
	case *Alt:
		pr.line("Alt")
		pr.pats(n.Steps)
	case *Stack:
		pr.line("Stack")
		pr.pats(n.Layers)
	case *Atom:
		pr.line("Atom %s %s", n.Kind, n.Text)
	case *Rest:
		pr.line("Rest")
	case *Splice:
		pr.line("Splice")
		pr.child(n.Expr)
	case *Fast:
		pr.line("Fast")
		pr.child(n.Pat)
		pr.child(&n.Factor)
	case *Slow:
		pr.line("Slow")
		pr.child(n.Pat)
		pr.child(&n.Factor)
	case *Repeat:
		pr.line("Repeat %d", n.Count)
		pr.child(n.Pat)
	case *Weight:
		pr.line("Weight %s", num(n.Amount))
		pr.child(n.Pat)
	case *Degrade:
		pr.line("Degrade %s", num(n.Amount))
		pr.child(n.Pat)
	case *Euclid:
		pr.line("Euclid %d %d %d", n.Steps, n.Pulses, n.Rotate)
		pr.child(n.Pat)

	// modifier factor
	case *Factor:
		pr.line("Factor")
		if n.Expr != nil {
			pr.child(n.Expr)
		} else if n.Pat != nil {
			pr.child(n.Pat)
		}
	default:
		pr.line("%T", n)
	}
}

func (pr *printer) synthStmt(s SynthStmt) {
	if s.Param {
		pr.line("SynthStmt param %s", s.Name)
	} else {
		pr.line("SynthStmt %s", s.Name)
	}
	pr.child(s.Value)
}

func (pr *printer) pats(ps []Pat) {
	for _, p := range ps {
		pr.child(p)
	}
}

func num(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}
