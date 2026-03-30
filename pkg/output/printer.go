// Package output provides console output, progress indicators, and formatting for scan results.
package output

import (
	"fmt"
	"os"
)

var globalPrinter Printer = &VoidPrinter{}

// ChangePrinter sets the active global printer.
func ChangePrinter(printer Printer) {
	globalPrinter = printer
}

// Printf writes formatted output using the current global printer.
func Printf(format string, args ...interface{}) {
	globalPrinter.Printf(format, args...)
}

// Printer is the interface for formatted output.
type Printer interface {
	Printf(format string, args ...interface{})
}

// ConsolePrinter writes to stderr.
type ConsolePrinter struct{}

// NewConsolePrinter creates a ConsolePrinter.
func NewConsolePrinter() *ConsolePrinter {
	return &ConsolePrinter{}
}

// Printf writes formatted output to stderr.
func (c *ConsolePrinter) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}

// VoidPrinter discards all output.
type VoidPrinter struct{}

// Printf discards the output.
func (v *VoidPrinter) Printf(_ string, _ ...interface{}) {}
