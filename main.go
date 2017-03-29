package main

import (
	"os"
	"tracerun/command"

	"github.com/drkaka/lg"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "tracerun"
	app.Usage = "command line application for TraceRun"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Run in debug level",
		},
	}
	app.Before = func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			lg.InitLogger(true)
		} else {
			lg.InitLogger(false)
		}
		lg.L(nil).Debug("lg initialized")
		return nil
	}

	app.Commands = []cli.Command{
		command.NewStartCMD(),
	}

	app.Run(os.Args)
}
