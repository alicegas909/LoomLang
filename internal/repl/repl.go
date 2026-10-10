// Package repl is the interactive loop (milestone M2): it reads submissions,
// evaluates each one completely, and prints the resulting value or error.
// Enter submits when the statement is complete (balanced brackets, no trailing
// operator); otherwise the line is buffered and a continuation is read.
//
// A submission that is just the word "quit" shuts the loop down (like EOF). It
// is handled here rather than in the grammar: quitting is a property of the
// REPL, not of the language.
package repl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/you/loom/internal/eval"
)

// ErrInterrupt is reported by the line reader when the user presses Ctrl-C.
// It clears the line being typed rather than ending the session.
var ErrInterrupt = errors.New("interrupt")

// QuitWord is the submission that shuts the loop down, like EOF.
const QuitWord = "quit"

// readLine reads one line, writing prompt first where the reader does not.
type readLine func(prompt string) (string, error)

// REPL reads submissions and sends them to an evaluator.
type REPL struct {
	ev eval.Evaluator

	// Prompt is written before the first line of a submission and
	// PromptContinue before later lines. Empty strings print nothing.
	Prompt         string
	PromptContinue string
}

// New returns a REPL backed by ev.
func New(ev eval.Evaluator) *REPL {
	return &REPL{ev: ev}
}

// Run reads from in until EOF or ctx is cancelled, writing prompts, results
// and errors to out. It is the plain, non-interactive reader used for pipes,
// files and tests.
func (r *REPL) Run(ctx context.Context, in io.Reader, out io.Writer) error {
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	return r.run(ctx, out, func(prompt string) (string, error) {
		if prompt != "" {
			io.WriteString(out, prompt)
		}
		if !sc.Scan() {
			if err := sc.Err(); err != nil {
				return "", err
			}
			return "", io.EOF
		}
		return sc.Text(), nil
	})
}

func (r *REPL) run(ctx context.Context, out io.Writer, read readLine) error {
	var buf strings.Builder
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		prompt := r.Prompt
		if buf.Len() > 0 {
			prompt = r.PromptContinue
		}

		line, err := read(prompt)
		switch {
		case errors.Is(err, io.EOF):
			return r.flush(out, &buf)
		case errors.Is(err, ErrInterrupt):
			buf.Reset()
			continue
		case err != nil:
			return err
		}

		buf.WriteString(line)
		buf.WriteByte('\n')

		if strings.TrimSpace(buf.String()) == QuitWord {
			return nil
		}

		if errors.Is(r.submit(buf.String(), out), eval.ErrIncomplete) {
			continue
		}
		buf.Reset()
	}
}

// flush evaluates whatever is left in buf when the input ends.
func (r *REPL) flush(out io.Writer, buf *strings.Builder) error {
	if buf.Len() == 0 {
		return nil
	}
	if errors.Is(r.submit(buf.String(), out), eval.ErrIncomplete) {
		fmt.Fprintln(out, eval.ErrIncomplete)
	}
	return nil
}

// submit evaluates src and prints its value, or the error if evaluation fails
// for a reason other than an unfinished submission. It returns the error so
// the caller can tell incomplete input apart from a real failure.
func (r *REPL) submit(src string, out io.Writer) error {
	res, err := r.ev.Eval(src)
	switch {
	case errors.Is(err, eval.ErrIncomplete):
		return err
	case err != nil:
		fmt.Fprintln(out, err)
		return err
	}
	if res.Value != nil {
		fmt.Fprintln(out, eval.Display(res.Value))
	}
	return nil
}
