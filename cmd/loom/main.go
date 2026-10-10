// Command loom is the REPL entry point and the composition root: it wires the
// evaluator, scheduler, engine and audio device together. Nothing else in the
// repo imports it.
package main

import (
	"fmt"
	"os"
)

const version = "0.0.0-dev"

func main() {
	fmt.Fprintf(os.Stderr, "loom %s: nothing to run yet (the REPL arrives in milestone M2)\n", version)
}
