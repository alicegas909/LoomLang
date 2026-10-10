// Command loom is the REPL entry point and the composition root: it wires the
// evaluator, scheduler, engine and audio device together. Nothing else in the
// repo imports it.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ergochat/readline"
	"github.com/you/loom/internal/eval"
	"github.com/you/loom/internal/repl"
)

const version = "0.0.0-dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "loom:", err)
		os.Exit(1)
	}
}

func run() error {
	r := repl.New(eval.NewTreeWalker())

	ctx := context.Background()
	if !isTerminal(os.Stdin) {
		return r.Run(ctx, os.Stdin, os.Stdout)
	}

	r.Prompt = "loom> "
	r.PromptContinue = "  ... "
	return r.RunInteractive(ctx, &readline.Config{
		Prompt:          r.Prompt,
		Stdin:           os.Stdin,
		Stdout:          os.Stdout,
		Stderr:          os.Stderr,
		HistoryFile:     historyFile(),
		HistoryLimit:    1000,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func historyFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".loom_history")
}
