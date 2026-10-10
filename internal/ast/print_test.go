package ast

import (
	"testing"

	"github.com/you/loom/internal/lexer"
)

func sp() Span { return Span{Start: lexer.Pos{Line: 1, Col: 1}, End: lexer.Pos{Line: 1, Col: 1}} }

func TestDump(t *testing.T) {
	cases := []struct {
		name string
		node Node
		want string
	}{
		{"number", &Number{Sp: sp(), Text: "42"}, "Number 42\n"},
		{"string", &String{Sp: sp(), Value: "kick.wav"}, "String \"kick.wav\"\n"},
		{"unary", &Unary{Sp: sp(), Op: "-", X: &Ident{Sp: sp(), Name: "x"}},
			"Unary -\n  Ident x\n"},
		{"field", &Field{Sp: sp(), Name: "gain", X: &Ident{Sp: sp(), Name: "lead"}},
			"Field gain\n  Ident lead\n"},
		{"apply", &Apply{Sp: sp(), Func: &Ident{Sp: sp(), Name: "f"}, Args: []Expr{
			&Number{Sp: sp(), Text: "1"},
		}}, "Apply\n  Ident f\n  Number 1\n"},
		{"fast", &Fast{Sp: sp(), Pat: &Seq{Sp: sp(), Steps: []Pat{
			&Atom{Sp: sp(), Kind: lexer.Ident, Text: "bd"},
		}}, Factor: Factor{Sp: sp(), Expr: &Number{Sp: sp(), Text: "4"}}},
			"Fast\n  Seq\n    Atom Ident bd\n  Factor\n    Number 4\n"},
		{"slow", &Slow{Sp: sp(), Pat: &Atom{Sp: sp(), Kind: lexer.Ident, Text: "bd"},
			Factor: Factor{Sp: sp(), Expr: &Number{Sp: sp(), Text: "2"}}},
			"Slow\n  Atom Ident bd\n  Factor\n    Number 2\n"},
		{"splice", &Splice{Sp: sp(), Expr: &Ident{Sp: sp(), Name: "x"}},
			"Splice\n  Ident x\n"},
		{"weight", &Weight{Sp: sp(), Amount: 2, Pat: &Atom{Sp: sp(), Kind: lexer.Note, Text: "g4"}},
			"Weight 2\n  Atom Note g4\n"},
		{"euclid", &Euclid{Sp: sp(), Steps: 3, Pulses: 8, Rotate: 1, Pat: &Atom{Sp: sp(), Kind: lexer.Ident, Text: "bd"}},
			"Euclid 3 8 1\n  Atom Ident bd\n"},
		{"synth", &SynthDef{Sp: sp(), Stmts: []SynthStmt{
			{Sp: sp(), Param: true, Name: "cutoff", Value: &Number{Sp: sp(), Text: "1200"}},
		}}, "SynthDef\n  SynthStmt param cutoff\n    Number 1200\n"},
		{"quant", &ExprStmt{Sp: sp(), HasQuant: true, Quant: 4, Value: &Ident{Sp: sp(), Name: "lead"}},
			"ExprStmt quant=4\n  Ident lead\n"},
	}
	for _, c := range cases {
		if got := Dump(c.node); got != c.want {
			t.Errorf("%s:\n got  %q\n want %q", c.name, got, c.want)
		}
	}
}

func TestDumpStmts(t *testing.T) {
	stmts := []Stmt{
		&Assign{Sp: sp(), Target: []string{"x"}, Value: &Number{Sp: sp(), Text: "1"}},
		&ExprStmt{Sp: sp(), Value: &Ident{Sp: sp(), Name: "hush"}},
	}
	want := "Assign target=x\n  Number 1\n\nExprStmt\n  Ident hush\n"
	if got := DumpStmts(stmts); got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
