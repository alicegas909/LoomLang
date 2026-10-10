package eval

import "testing"

func TestEnvChainAndShadowing(t *testing.T) {
	global := NewEnv(nil)
	global.Set("x", Number(1))

	child := NewEnv(global)
	if v, ok := child.Get("x"); !ok || v != Number(1) {
		t.Fatalf("child lookup: got %v, %v", v, ok)
	}

	child.Set("x", Number(2))
	if v, _ := child.Get("x"); v != Number(2) {
		t.Errorf("shadowed value: got %v, want 2", v)
	}
	if v, _ := global.Get("x"); v != Number(1) {
		t.Errorf("global mutated: got %v, want 1", v)
	}
	if _, ok := child.Get("y"); ok {
		t.Error("undefined y reported as found")
	}
}

func TestNoteToMIDI(t *testing.T) {
	cases := []struct {
		name string
		want float64
		ok   bool
	}{
		{"c4", 60, true},
		{"a4", 69, true},
		{"fs3", 54, true},
		{"bf2", 46, true},
		{"c0", 12, true},
		{"bd", 0, false},
		{"h4", 0, false},
		{"c", 0, false},
	}
	for _, c := range cases {
		got, ok := NoteToMIDI(c.name)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("NoteToMIDI(%q) = %v, %v; want %v, %v", c.name, got, ok, c.want, c.ok)
		}
	}
}
