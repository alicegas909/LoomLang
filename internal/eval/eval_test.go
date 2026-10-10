package eval

import (
	"errors"
	"testing"
)

func evalOne(t *testing.T, ev *TreeWalker, src string) Result {
	t.Helper()
	res, err := ev.Eval(src)
	if err != nil {
		t.Fatalf("Eval(%q): %v", src, err)
	}
	return res
}

func wantNumber(t *testing.T, v Value, want float64) {
	t.Helper()
	n, ok := v.(Number)
	if !ok {
		t.Fatalf("value is %T, want Number", v)
	}
	if float64(n) != want {
		t.Errorf("got %v, want %v", n, want)
	}
}

func TestArithmeticAndPrecedence(t *testing.T) {
	cases := []struct {
		src  string
		want float64
	}{
		{"1 + 2 * 3", 7},
		{"1 * 2 + 3", 5},
		{"10 - 2 - 3", 5},
		{"10 / 4", 2.5},
		{"-2 + 3", 1},
		{"-(1 + 2)", -3},
		{"2 * -3", -6},
	}
	for _, c := range cases {
		res := evalOne(t, NewTreeWalker(), c.src)
		wantNumber(t, res.Value, c.want)
	}
}

func TestLiterals(t *testing.T) {
	ev := NewTreeWalker()
	wantNumber(t, evalOne(t, ev, "c4").Value, 60)
	wantNumber(t, evalOne(t, ev, "a4").Value, 69)
	if s, ok := evalOne(t, ev, `"hi"`).Value.(String); !ok || s != "hi" {
		t.Errorf(`"hi" evaluated to %v`, evalOne(t, ev, `"hi"`).Value)
	}
}

func TestBindingsPersist(t *testing.T) {
	ev := NewTreeWalker()
	res := evalOne(t, ev, "x = 5")
	if res.Value != nil {
		t.Errorf("assignment produced a value: %v", res.Value)
	}
	wantNumber(t, evalOne(t, ev, "x + 1").Value, 6)
}

func TestLambdaAndApplication(t *testing.T) {
	cases := []struct {
		src  string
		want float64
	}{
		{`(\x -> x * 2) 21`, 42},
		{"add = \\x y -> x + y\nadd 1 2", 3},
		{"add = \\x y -> x + y\ninc = add 1\ninc 41", 42},
		{"inc = (\\x -> x + 1) >> (\\x -> x * 2)\ninc 3", 8},
		{"5 |> (\\x -> x * 3)", 15},
	}
	for _, c := range cases {
		res := evalOne(t, NewTreeWalker(), c.src)
		wantNumber(t, res.Value, c.want)
	}
}

func TestRuntimeErrors(t *testing.T) {
	cases := []string{
		"foo",
		"1 / 0",
		`1 + "a"`,
		"1 2",
		"x = 1\nx.y = 2",
		"kit <- [a b]",
		"synth { out = saw }",
		"[a b]",
	}
	for _, src := range cases {
		_, err := NewTreeWalker().Eval(src)
		if err == nil {
			t.Errorf("Eval(%q): expected an error", src)
			continue
		}
		if !errors.Is(err, ErrIncomplete) && !containsColon(err.Error()) {
			t.Errorf("Eval(%q): error %q lacks line:col", src, err)
		}
	}
}

func TestIncomplete(t *testing.T) {
	for _, src := range []string{"1 +", "[a b", "x = (1", "synth { out = saw", "a |>", "\\x ->"} {
		_, err := NewTreeWalker().Eval(src)
		if !errors.Is(err, ErrIncomplete) {
			t.Errorf("Eval(%q): got %v, want ErrIncomplete", src, err)
		}
	}
}

func TestRollbackOnError(t *testing.T) {
	ev := NewTreeWalker()
	evalOne(t, ev, "y = 1")

	if _, err := ev.Eval("y = 2\nz = nope"); err == nil {
		t.Fatal("expected an error")
	}
	wantNumber(t, evalOne(t, ev, "y").Value, 1)
	if _, err := ev.Eval("z"); err == nil {
		t.Error("z should not exist after a rolled-back submission")
	}
}

func containsColon(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return true
		}
	}
	return false
}
