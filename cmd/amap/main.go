package main

import (
	"io"
	"os"

	"github.com/minggeorgelei/AMAP-client/cli"
)

var exit = os.Exit

func main() {
	exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	return cli.Run(args, stdout, stderr)
}
