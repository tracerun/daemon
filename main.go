package main

import (
	"os"
	"tracerun/command"
	"tracerun/db"
	"tracerun/db/action"
	"tracerun/lg"

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
			Value: "tracerun.db",
			Usage: "Path for db file.",
		},
		cli.UintFlag{
			Name:  "e",
			Value: 15,
			Usage: "Action expiration seconds.",
		},
	}

	app.Before = func(c *cli.Context) error {
		logPath := c.GlobalString("o")
		lg.InitLogger(c.GlobalBool("debug"), c.GlobalBool("nostd"), logPath)
		lg.L.Debug("logger initialized")
		// set db path
		db.SetDBPath(c.GlobalString("db"))
		// set action expiration
		action.Expired = uint32(c.GlobalUint("e"))
		return nil
	}

	app.Commands = []cli.Command{
		command.NewStartCMD(),
		command.NewAddCMD(),
		command.NewShowCMD(),
	}

	app.Run(os.Args)
}
