package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "tracerun"
	app.Usage = "command line application for TraceRun"
	app.Run(os.Args)
}
