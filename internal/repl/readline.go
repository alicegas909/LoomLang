package repl

import (
	"context"
	"errors"
	"io"

	"github.com/ergochat/readline"
)

// RunInteractive reads lines with line editing and history. cfg configures the
// terminal; its Prompt defaults to r.Prompt. Use it only when stdin is a
// terminal — for pipes and tests call Run instead.
func (r *REPL) RunInteractive(ctx context.Context, cfg *readline.Config) error {
	if cfg == nil {
		cfg = &readline.Config{}
	}
	if cfg.Prompt == "" {
		cfg.Prompt = r.Prompt
	}
	rl, err := readline.NewFromConfig(cfg)
	if err != nil {
		return err
	}
	defer rl.Close()

	return r.run(ctx, rl.Stdout(), func(prompt string) (string, error) {
		rl.SetPrompt(prompt)
		line, err := rl.ReadLine()
		switch {
		case errors.Is(err, readline.ErrInterrupt):
			return "", ErrInterrupt
		case errors.Is(err, io.EOF):
			return "", io.EOF
		case err != nil:
			return "", err
		}
		return line, nil
	})
}
