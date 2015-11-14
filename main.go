package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/rosenhouse/tubes/application/commands"
)

func main() {
	commands := commands.New()
	parser := flags.NewParser(commands, flags.HelpFlag|flags.PassDoubleDash)

	_, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		if ferr, ok := err.(*flags.Error); ok && ferr.Type != flags.ErrHelp {
			parser.WriteHelp(os.Stderr)
		}
		os.Exit(1)
	}
}
