package lexer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func kinds(src string) []Kind {
	var out []Kind
	for _, t := range Tokenize(src) {
		if t.Kind != EOF {
			out = append(out, t.Kind)
		}
	}
	return out
}

func texts(src string) []string {
	var out []string
	for _, t := range Tokenize(src) {
		if t.Kind != EOF {
			out = append(out, t.Text)
		}
	}
	return out
}

// flags renders spacing: a leading _ means SpaceBefore, a trailing _ means SpaceAfter.
func flags(src string) string {
	var parts []string
	for _, tok := range Tokenize(src) {
		if tok.Kind == EOF {
			continue
		}
		s := tok.Text
		if tok.Kind == Newline {
			s = "NL"
		}
		if tok.SpaceBefore {
			s = "_" + s
		}
		if tok.SpaceAfter {
			s += "_"
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, " ")
}

func TestKinds(t *testing.T) {
	cases := []struct {
		src  string
		want []Kind
	}{
		{``, nil},
		{"\n", []Kind{Newline}},
		{`x = 1`, []Kind{Ident, Equals, Number}},
		{`[a2 c3 e3 g3]*2`, []Kind{LBracket, Note, Note, Note, Note, RBracket, Star, Number}},
		{`[bd*4, [~ hh]*4]`, []Kind{LBracket, Ident, Star, Number, Comma, LBracket, Tilde, Ident, RBracket, Star, Number, RBracket}},
		{`<a <b c>>`, []Kind{LAngle, Ident, LAngle, Ident, Ident, RAngle, RAngle}},
		{`<-1 2>`, []Kind{LAngle, Minus, Number, Number, RAngle}},
		{`lead <- x`, []Kind{Ident, LAngle, Minus, Ident}},
		{`x |> f >> g`, []Kind{Ident, Pipe, RAngle, Ident, RAngle, RAngle, Ident}},
		{`\x -> x`, []Kind{Backslash, Ident, Minus, RAngle, Ident}},
		{`bd(3,8)`, []Kind{Ident, LParen, Number, Comma, Number, RParen}},
		{`c4? e4!3 g4@2 ~`, []Kind{Note, Question, Note, Bang, Number, Note, At, Number, Tilde}},
		{`@4 lead`, []Kind{At, Number, Ident}},
		{`lead.gain = 0.8`, []Kind{Ident, Dot, Ident, Equals, Number}},
		{`$x`, []Kind{Dollar, Ident}},
		{`-1.5 - 2 + 3 / 4`, []Kind{Minus, Number, Minus, Number, Plus, Number, Slash, Number}},
		{`synth { param cutoff = 1200; out = saw }`, []Kind{KwSynth, LBrace, KwParam, Ident, Equals, Number, Semicolon, Ident, Equals, Ident, RBrace}},
		// note literals vs identifiers
		{`ef2 fs3 bd4 a2b bf as4 b2`, []Kind{Note, Note, Ident, Ident, Ident, Note, Note}},
		// comments and line endings
		{"x # comment\ny", []Kind{Ident, Newline, Ident}},
		{"a\r\nb", []Kind{Ident, Newline, Ident}},
		// strings
		{`sample "kick.wav"`, []Kind{Ident, String}},
		{"\"abc\nx", []Kind{Illegal, Newline, Ident}},
		{`"abc`, []Kind{Illegal}},
		// illegal input
		{`x & y`, []Kind{Ident, Illegal, Ident}},
		{`x é y`, []Kind{Ident, Illegal, Ident}},
	}
	for _, c := range cases {
		got, want := fmt.Sprint(kinds(c.src)), fmt.Sprint(c.want)
		if got != want {
			t.Errorf("kinds(%q)\n got  %s\n want %s", c.src, got, want)
		}
	}
}

func TestTokenText(t *testing.T) {
	cases := []struct {
		src  string
		want []string
	}{
		{`0.25 1.2.3 3.`, []string{"0.25", "1.2", ".", "3", "3", "."}},
		{`lead.gain`, []string{"lead", ".", "gain"}},
		{`ef2 "hi"`, []string{"ef2", `"hi"`}},
		{`<-1`, []string{"<", "-", "1"}},
	}
	for _, c := range cases {
		got, want := fmt.Sprintf("%q", texts(c.src)), fmt.Sprintf("%q", c.want)
		if got != want {
			t.Errorf("texts(%q)\n got  %s\n want %s", c.src, got, want)
		}
	}
}

func TestSpacingFlags(t *testing.T) {
	cases := []struct {
		src, want string
	}{
		// tight modifier vs spaced arithmetic
		{`[a b]*2`, `_[ a_ _b ] * 2_`},
		{`[a b] * 2`, `_[ a_ _b ]_ _*_ _2_`},
		// euclid vs splice
		{`bd(3,8) bd (3)`, `_bd ( 3 , 8 )_ _bd_ _( 3 )_`},
		// start of a line counts as space before
		{"x\ny", `_x_ NL _y_`},
		{"x\n  |> y", `_x_ NL_ _| >_ _y_`},
		// a comment counts as space
		{"x#c\ny", `_x_ _NL _y_`},
	}
	for _, c := range cases {
		if got := flags(c.src); got != c.want {
			t.Errorf("flags(%q)\n got  %s\n want %s", c.src, got, c.want)
		}
	}
}

func TestPositions(t *testing.T) {
	toks := Tokenize("a\n  bc")
	want := []Pos{{1, 1}, {1, 2}, {2, 3}, {2, 5}} // a, newline, bc, EOF
	if len(toks) != len(want) {
		t.Fatalf("got %d tokens, want %d: %v", len(toks), len(want), toks)
	}
	for i, p := range want {
		if toks[i].Pos != p {
			t.Errorf("token %d (%v): pos %v, want %v", i, toks[i], toks[i].Pos, p)
		}
	}

	// columns count runes, not bytes
	toks = Tokenize("é x")
	if got := toks[1].Pos; got != (Pos{1, 3}) {
		t.Errorf("x after a 2-byte rune is at %v, want 1:3", got)
	}
}

func TestEOFRepeats(t *testing.T) {
	l := New("x")
	l.Next()
	for i := 0; i < 3; i++ {
		if tok := l.Next(); tok.Kind != EOF {
			t.Fatalf("call %d after the end returned %v, want EOF", i, tok)
		}
	}
}

func TestKindString(t *testing.T) {
	if kindCount != 32 {
		t.Errorf("kindCount = %d, want 32; update the spec and these tests together", kindCount)
	}
	for k := Kind(0); k < kindCount; k++ {
		if name := k.String(); name == "" || strings.HasPrefix(name, "Kind(") {
			t.Errorf("kind %d has no name", int(k))
		}
	}
	if got := Kind(99).String(); got != "Kind(99)" {
		t.Errorf("out-of-range kind prints %q", got)
	}
}

// Every example program must lex without Illegal tokens and end in EOF.
func TestExamplesLexClean(t *testing.T) {
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
		toks := Tokenize(string(src))
		for _, tok := range toks {
			if tok.Kind == Illegal {
				t.Errorf("%s: illegal token %v", f, tok)
			}
		}
		if toks[len(toks)-1].Kind != EOF {
			t.Errorf("%s: last token is not EOF", f)
		}
	}
}

// FuzzLex checks the lexer's core promises on arbitrary input: it terminates,
// every token consumes text, positions are valid, and EOF comes last.
func FuzzLex(f *testing.F) {
	seeds := []string{"", "[a2 c3]*2", "<a <b c>>", "x <- [bd sn]", "\"unterminated", "é\x00\xff", "0.5.", "# only a comment"}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		l := New(src)
		for i := 0; i <= len(src)+1; i++ {
			tok := l.Next()
			if tok.Pos.Line < 1 || tok.Pos.Col < 1 {
				t.Fatalf("bad position %v", tok)
			}
			if tok.Kind == EOF {
				return
			}
			if tok.Text == "" {
				t.Fatalf("empty text for %v", tok)
			}
		}
		t.Fatalf("no EOF within %d tokens for %q", len(src)+2, src)
	})
}
