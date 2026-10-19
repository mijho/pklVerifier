package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	err := CLIApp.Run(os.Args)
	if err == nil {
		return
	}
	var ec cli.ExitCoder
	if errors.As(err, &ec) {
		os.Exit(ec.ExitCode())
	}
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
