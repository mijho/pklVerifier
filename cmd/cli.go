package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

// Build-time variables, set with -ldflags "-X main.Version=... -X main.Commit=...".
var (
	Version = "dev"
	Commit  = "none"
)

// CLIApp constructs the command line application.
var CLIApp = &cli.App{
	Name:    "pklVerifier",
	Usage:   "Validate DCP asset hashes and sizes against the PKL",
	Version: Version + " (" + Commit + ")",
	Commands: []*cli.Command{
		{
			Name:  "validate",
			Usage: "validate a DCP directory",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "input",
					Aliases:  []string{"i", "d"},
					Usage:    "Path to the DCP directory",
					Required: true,
				},
				&cli.StringFlag{
					Name:    "output",
					Aliases: []string{"o"},
					Usage:   "Write JSON results to this file instead of stdout",
				},
			},
			Action: ValidateHandler,
		},
	},
}

// ValidateHandler runs validation, emits JSON, and exits 2 if any asset failed.
func ValidateHandler(c *cli.Context) error {
	res, err := Validate(c.String("input"))
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling results: %w", err)
	}

	if out := c.String("output"); out != "" {
		if err := os.WriteFile(out, append(b, '\n'), 0o644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	} else {
		fmt.Println(string(b))
	}

	if res.Errors > 0 {
		return cli.Exit(fmt.Sprintf("%d asset(s) failed validation", res.Errors), 2)
	}
	return nil
}
