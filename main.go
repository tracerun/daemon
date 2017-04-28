package main

import (
	"os"

	"github.com/tracerun/tracerun/command"
	"github.com/tracerun/tracerun/lg"
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
			Usage: "Run in debug level.",
		},
		cli.BoolFlag{
			Name:  "nostd",
			Usage: "No need output to stderr.",
		},
		cli.StringFlag{
			Name:  "o",
			Usage: "Path for output log file.",
		},
		cli.StringFlag{
			Name:  "db",
			Value: "tracerun",
			Usage: "Path for db folder.",
		},
		cli.UintFlag{
			Name:  "p",
			Value: 19869,
			Usage: "TCP port.",
		},
	}

	app.Before = func(c *cli.Context) error {
		logPath := c.GlobalString("o")
		lg.InitLogger(c.GlobalBool("debug"), c.GlobalBool("nostd"), logPath)
		lg.L.Debug("logger initialized")
		return nil
	}

	app.Commands = []cli.Command{
		command.NewStartCMD(),
		command.NewAddCMD(),
		command.NewListCMD(),
	}

	app.Run(os.Args)
}
