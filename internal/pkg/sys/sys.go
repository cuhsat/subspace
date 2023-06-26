// Package sys contains operating system specific utility functions.
package sys

import (
	"fmt"
	"io"
	"os"
)

// Stats is the temporary file for statistics output.
// This will only work on POSIX compatible systems.
var Stats, _ = os.OpenFile("/tmp/subspace", os.O_RDWR|os.O_CREATE, 0666)

// Stdin reads all input from the processes standard input
// until EOF is reached or an error occurs. It returns it
// as an array of bytes.
func Stdin() (b []byte) {
	fi, err := os.Stdin.Stat()

	if err != nil {
		Fatal(err)
	}

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		b, _ = io.ReadAll(os.Stdin)
	}

	return
}

// Error prints any given error to the standard error output.
func Error(e ...any) {
	fmt.Fprint(os.Stderr, fmt.Sprintln("⇌", e))
}

// Fatal prints any given error to the standard error output
// and exits the current program with a non-zero exit code
// indicating an error.
func Fatal(e ...any) {
	fmt.Fprint(os.Stderr, fmt.Sprintln("⇌", e))
	os.Exit(1)
}
