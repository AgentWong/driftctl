package test

import (
	"bytes"

	"github.com/spf13/cobra"
)

// Execute runs the given cobra command with the provided arguments and returns the combined output.
func Execute(cmd *cobra.Command, args ...string) (output string, err error) {
	_, output, err = ExecuteC(cmd, args...)
	return output, err
}

// ExecuteC runs the given cobra command with the provided arguments and returns the executed command, output, and error.
func ExecuteC(cmd *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	c, err = cmd.ExecuteC()

	return c, buf.String(), err
}
