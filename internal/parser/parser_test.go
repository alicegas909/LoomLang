package parser

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/you/loom/internal/ast"
)

var update = flag.Bool("update", false, "rewrite the golden AST dumps")

// TestExamplesParse checks that every program in testdata/examples parses.
func TestExamplesParse(t *testing.T) {
	files, err := filepath.Glob("../../testdata/examples/*.loom")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no example programs found in testdata/examples")
	}
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		stmts, err := Parse(string(src))
		if err != nil {
			t.Errorf("%s: %v", f, err)
			continue
		}
		if len(stmts) == 0 {
			t.Errorf("%s: parsed no statements", f)
		}
	}
}

// TestGolden compares the AST dump of every example to testdata/golden. Run
// `go test ./internal/parser -run TestGolden -update` to regenerate them.
func TestGolden(t *testing.T) {
	files, err := filepath.Glob("../../testdata/examples/*.loom")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no example programs found in testdata/examples")
	}
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		stmts, err := Parse(string(src))
		if err != nil {
			t.Errorf("%s: %v", f, err)
			continue
		}
		got := ast.DumpStmts(stmts)
		name := strings.TrimSuffix(filepath.Base(f), filepath.Ext(f)) + ".golden"
		golden := filepath.Join("../../testdata/golden", name)
		if *update {
			if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatalf("%s: missing; run `go test ./internal/parser -run TestGolden -update`", golden)
		}
		if got != string(want) {
			t.Errorf("%s: dump mismatch\n--- got ---\n%s--- want ---\n%s", f, got, want)
		}
	}
}

func TestExprShape(t *testing.T) {
	cases := []struct{ src, want string }{
		{"1 + 2 * 3", "(+ 1 (* 2 3))"},
		{"1 * 2 + 3", "(+ (* 1 2) 3)"},
		{"a - b - c", "(- (- a b) c)"},
		{"f x + g y", "(+ (apply f x) (apply g y))"},
		{"-a * b", "(* (- a) b)"},
		{"-f x", "(- (apply f x))"},
		{"a |> b >> c", "(|> a (>> b c))"},
		{"a >> b |> c", "(|> (>> a b) c)"},
		{"a.b.c", "(field (field a b) c)"},
		{`\x y -> x * y`, "(lambda (x y) (* x y))"},
		{"saw (freq * 1.01)", "(apply saw (* freq 1.01))"},
		{"adsr 0.01 0.3 0.4 0.5", "(apply adsr 0.01 0.3 0.4 0.5)"},
		{"[a b]*2", "(pat (fast (seq a b) 2))"},
		{"[a b] * 2", "(* (pat (seq a b)) 2)"},
		{"<0 3 5 7>", "(pat (alt 0 3 5 7))"},
	}
	for _, c := range cases {
		stmts, err := Parse(c.src)
		if err != nil {
			t.Errorf("Parse(%q): %v", c.src, err)
			continue
		}
		es, ok := stmts[0].(*ast.ExprStmt)
		if !ok {
			t.Errorf("Parse(%q): statement is %T, want *ast.ExprStmt", c.src, stmts[0])
			continue
		}
		if got := exprString(es.Value); got != c.want {
			t.Errorf("Parse(%q)\n got  %s\n want %s", c.src, got, c.want)
		}
	}
}

func TestPatterns(t *testing.T) {
	cases := []struct{ src, want string }{
		{"[bd*4, [~ hh]*4]", "(pat (stack (seq (fast bd 4)) (seq (fast (seq ~ hh) 4))))"},
		{"[a2 ~ e3 ~]", "(pat (seq a2 ~ e3 ~))"},
		{"[bd(3,8)]", "(pat (seq (euclid bd 3 8 0)))"},
		{"[bd (3)]", "(pat (seq bd (splice 3)))"},
		{"[c4? e4!3 g4@2]", "(pat (seq (degrade c4 0.5) (repeat e4 3) (weight g4 2)))"},
		{"[$x]", "(pat (seq (splice x)))"},
	}
	for _, c := range cases {
		stmts, err := Parse(c.src)
		if err != nil {
			t.Errorf("Parse(%q): %v", c.src, err)
			continue
		}
		var got string
		if es, ok := stmts[0].(*ast.ExprStmt); ok {
			got = exprString(es.Value)
		}
		if got != c.want {
			t.Errorf("Parse(%q)\n got  %s\n want %s", c.src, got, c.want)
		}
	}
}

func TestStatements(t *testing.T) {
	cases := []struct{ src, want string }{
		{"x = 1 + 2", "(= x (+ 1 2))"},
		{"lead.gain = 0.8", "(= lead.gain 0.8)"},
		{"lead <- <800 1600>", "(attach (lead) (pat (alt 800 1600)))"},
		{"@4 lead <- [a2 ~]", "(attach @4 (lead) (pat (seq a2 ~)))"},
		{"show lead", "(stmt (apply show lead))"},
		{"bpm 120", "(stmt (apply bpm 120))"},
	}
	for _, c := range cases {
		stmts, err := Parse(c.src)
		if err != nil {
			t.Errorf("Parse(%q): %v", c.src, err)
			continue
		}
		got := stmtString(stmts[0])
		if got != c.want {
			t.Errorf("Parse(%q)\n got  %s\n want %s", c.src, got, c.want)
		}
	}
}

func TestSynth(t *testing.T) {
	src := "pluck = synth {\n  param cutoff = 1200\n  out = saw * 0.5\n}"
	stmts, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	got := stmtString(stmts[0])
	want := "(= pluck (synth (param cutoff=1200) (out=(* saw 0.5))))"
	if got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestContinuation(t *testing.T) {
	src := "lead <- [a2 c3]*2 |> fast 2\n  |> every 4 rev\nhush"
	stmts, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(stmts) != 2 {
		t.Fatalf("got %d statements, want 2", len(stmts))
	}
	if _, ok := stmts[1].(*ast.ExprStmt); !ok {
		t.Errorf("second statement is %T, want *ast.ExprStmt", stmts[1])
	}

	// A newline inside unbalanced brackets does not end the statement.
	stmts, err = Parse("x = (1\n  + 2)\ny = 3")
	if err != nil {
		t.Fatal(err)
	}
	if len(stmts) != 2 {
		t.Fatalf("got %d statements, want 2", len(stmts))
	}
	if got := stmtString(stmts[0]); got != "(= x (+ 1 2))" {
		t.Errorf("got %s, want (= x (+ 1 2))", got)
	}
}

func TestErrors(t *testing.T) {
	for _, src := range []string{
		"x = ",
		"[a b",
		"f = = 1",
		"lead <- ",
		"synth { out = }",
	} {
		stmts, err := Parse(src)
		if err == nil {
			t.Errorf("Parse(%q): expected an error", src)
		} else if !strings.Contains(err.Error(), ":") {
			t.Errorf("Parse(%q): error %q has no line:col", src, err)
		}
		if len(stmts) != 0 {
			t.Errorf("Parse(%q): got %d statements, want 0", src, len(stmts))
		}
	}
}

// TestRecovery checks that one bad statement does not stop the ones around it,
// and that every error is reported.
func TestRecovery(t *testing.T) {
	src := "x = \nok = 1\nbad = *\nz = 2"
	stmts, err := Parse(src)
	if err == nil {
		t.Fatal("expected errors")
	}
	list, ok := err.(ErrorList)
	if !ok {
		t.Fatalf("error is %T, want ErrorList", err)
	}
	if len(list) != 2 {
		t.Fatalf("got %d errors, want 2: %v", len(list), list)
	}
	if !strings.Contains(list[0].Error(), "1:") || !strings.Contains(list[1].Error(), "3:") {
		t.Errorf("errors point at the wrong lines: %v", list)
	}
	if len(stmts) != 2 {
		t.Fatalf("got %d statements, want 2", len(stmts))
	}
	if got := stmtString(stmts[0]); got != "(= ok 1)" {
		t.Errorf("first recovered statement: got %s", got)
	}
	if got := stmtString(stmts[1]); got != "(= z 2)" {
		t.Errorf("second recovered statement: got %s", got)
	}
}

// FuzzParse must never panic or hang on any input.
func FuzzParse(f *testing.F) {
	seeds := []string{
		"", "x = 1 + 2 * 3", `\x -> x`, "[a b]*2", "<a <b c>>",
		"kit <- [bd*4, [~ hh]*4]", "synth { param c = 1; out = saw }",
		"@4 lead <- [a2 ~ e3 ~] |> fast 2", "a |> b >> c", "bd(3,8) c4? e4!3",
		"lead <- [[c4 e4] g4, [a4 ~]*2]", "[$x ~ <f2 g2> ~]!2",
		"drums <- [bd(3,8), [~ cp]*4, [~ ~ ht?0.4]!2]", "total = (1 +\n  2 * 3)",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		_, _ = Parse(src)
	})
}

// --- AST rendering for tests ---

func stmtString(s ast.Stmt) string {
	switch n := s.(type) {
	case *ast.Assign:
		return "(= " + pathString(n.Target) + " " + exprString(n.Value) + ")"
	case *ast.Attach:
		if n.HasQuant {
			return "(attach @" + trimNum(n.Quant) + " (" + pathString(n.Target) + ") " + exprString(n.Value) + ")"
		}
		return "(attach (" + pathString(n.Target) + ") " + exprString(n.Value) + ")"
	case *ast.ExprStmt:
		return "(stmt " + exprString(n.Value) + ")"
	}
	return "?"
}

func pathString(p []string) string {
	return strings.Join(p, ".")
}

func exprString(e ast.Expr) string {
	switch n := e.(type) {
	case *ast.Number:
		return trimNum(n.Value)
	case *ast.Note:
		return n.Text
	case *ast.String:
		return `"` + n.Value + `"`
	case *ast.Ident:
		return n.Name
	case *ast.Unary:
		return "(" + n.Op + " " + exprString(n.X) + ")"
	case *ast.Binary:
		return "(" + n.Op + " " + exprString(n.X) + " " + exprString(n.Y) + ")"
	case *ast.Apply:
		parts := make([]string, 0, len(n.Args)+1)
		parts = append(parts, exprString(n.Func))
		for _, a := range n.Args {
			parts = append(parts, exprString(a))
		}
		return "(apply " + strings.Join(parts, " ") + ")"
	case *ast.Field:
		return "(field " + exprString(n.X) + " " + n.Name + ")"
	case *ast.Lambda:
		return "(lambda (" + strings.Join(n.Params, " ") + ") " + exprString(n.Body) + ")"
	case *ast.SynthDef:
		parts := make([]string, 0, len(n.Stmts))
		for _, st := range n.Stmts {
			name := st.Name + "="
			if st.Param {
				parts = append(parts, "(param "+name+exprString(st.Value)+")")
			} else {
				parts = append(parts, "("+name+exprString(st.Value)+")")
			}
		}
		return "(synth " + strings.Join(parts, " ") + ")"
	case *ast.PatternExpr:
		return "(pat " + patString(n.Pat) + ")"
	}
	return "?"
}

func patString(p ast.Pat) string {
	switch n := p.(type) {
	case *ast.Seq:
		return "(seq" + children(n.Steps) + ")"
	case *ast.Alt:
		return "(alt" + children(n.Steps) + ")"
	case *ast.Stack:
		return "(stack" + children(n.Layers) + ")"
	case *ast.Atom:
		return n.Text
	case *ast.Rest:
		return "~"
	case *ast.Splice:
		return "(splice " + exprString(n.Expr) + ")"
	case *ast.Fast:
		return "(fast " + patString(n.Pat) + " " + factorString(n.Factor) + ")"
	case *ast.Slow:
		return "(slow " + patString(n.Pat) + " " + factorString(n.Factor) + ")"
	case *ast.Repeat:
		return "(repeat " + patString(n.Pat) + " " + trimNum(float64(n.Count)) + ")"
	case *ast.Weight:
		return "(weight " + patString(n.Pat) + " " + trimNum(n.Amount) + ")"
	case *ast.Degrade:
		return "(degrade " + patString(n.Pat) + " " + trimNum(n.Amount) + ")"
	case *ast.Euclid:
		return "(euclid " + patString(n.Pat) + " " +
			trimNum(float64(n.Steps)) + " " + trimNum(float64(n.Pulses)) + " " + trimNum(float64(n.Rotate)) + ")"
	}
	return "?"
}

func factorString(f ast.Factor) string {
	if f.Expr != nil {
		return exprString(f.Expr)
	}
	return patString(f.Pat)
}

func children(ps []ast.Pat) string {
	var b strings.Builder
	for _, p := range ps {
		b.WriteByte(' ')
		b.WriteString(patString(p))
	}
	return b.String()
}

func trimNum(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}
