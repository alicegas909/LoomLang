package repl

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/you/loom/internal/eval"
)

func runInput(t *testing.T, input string) string {
	t.Helper()
	var out bytes.Buffer
	if err := New(eval.NewTreeWalker()).Run(context.Background(), strings.NewReader(input), &out); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return out.String()
}

func TestPrintsValue(t *testing.T) {
	if got := runInput(t, "1 + 2\n"); got != "3\n" {
		t.Errorf("got %q, want %q", got, "3\n")
	}
}

func TestAssignmentPrintsNothing(t *testing.T) {
	if got := runInput(t, "x = 1\nx\n"); got != "1\n" {
		t.Errorf("got %q, want %q", got, "1\n")
	}
}

func TestMultilineContinuation(t *testing.T) {
	if got := runInput(t, "1 +\n2\n"); got != "3\n" {
		t.Errorf("got %q, want %q", got, "3\n")
	}
}

func TestErrorDisplayed(t *testing.T) {
	got := runInput(t, "foo\n")
	if !strings.Contains(got, "1:1: undefined: foo") {
		t.Errorf("got %q", got)
	}
}

func TestIncompleteAtEOF(t *testing.T) {
	got := runInput(t, "1 +")
	if !strings.Contains(got, eval.ErrIncomplete.Error()) {
		t.Errorf("got %q, want incomplete notice", got)
	}
}

func TestQuitStopsLoop(t *testing.T) {
	// quit must stop the loop before later statements are evaluated.
	if got := runInput(t, "1\nquit\n2\n"); got != "1\n" {
		t.Errorf("got %q, want %q", got, "1\n")
	}
	// surrounding whitespace is ignored.
	if got := runInput(t, "  quit  \n"); got != "" {
		t.Errorf("got %q, want empty output", got)
	}
}

func TestQuitOnlyAsWholeSubmission(t *testing.T) {
	if got := runInput(t, "quitter = 1\nquitter\n"); got != "1\n" {
		t.Errorf("got %q, want %q (quit must match the whole submission)", got, "1\n")
	}
}

func TestPrompts(t *testing.T) {
	var out bytes.Buffer
	r := New(eval.NewTreeWalker())
	r.Prompt = "> "
	r.PromptContinue = "... "
	if err := r.Run(context.Background(), strings.NewReader("1 +\n2\n"), &out); err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), "> ... 3\n> "; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := New(eval.NewTreeWalker()).Run(ctx, strings.NewReader("1\n"), io.Discard)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("got %v, want context.Canceled", err)
	}
}

func TestInterruptClearsBuffer(t *testing.T) {
	r := New(eval.NewTreeWalker())
	lines := []struct {
		line string
		err  error
	}{
		{"1 +", nil},
		{"", ErrInterrupt},
		{"2", nil},
		{"", io.EOF},
	}
	i := 0
	read := func(string) (string, error) {
		x := lines[i]
		i++
		return x.line, x.err
	}

	var out bytes.Buffer
	if err := r.run(context.Background(), &out, read); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "2\n" {
		t.Errorf("got %q, want %q (interrupt should discard the pending line)", got, "2\n")
	}
}
