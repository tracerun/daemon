package command

import (
	"errors"
	"os"
	"os/exec"

	"github.com/tracerun/tracerun/service"
	"github.com/urfave/cli"
)

// NewStartCMD create a start command.
func NewStartCMD() cli.Command {
	return cli.Command{
		Name:   "start",
		Usage:  "start TCP service",
		Action: startAction,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "d",
				Usage: "Run in background mode.",
			},
		},
	}
}

func startAction(c *cli.Context) error {
	p := c.GlobalUint("p")
	if c.Bool("d") {
		idx := 0
		for i := 0; i < len(os.Args); i++ {
			if os.Args[i] == "-d" {
				idx = i
				break
			}
		}

		// "-d" can't be the first two arg.
		if idx < 2 {
			return errors.New("-d flag is wrong")
		}
		args := append(os.Args[:idx], os.Args[idx+1:]...)
		cmd := exec.Command(args[0], args[1:]...)

		// start the command
		if err := cmd.Start(); err != nil {
			return err
		}

		// release the process from this thread, make it a daemon
		if err := cmd.Process.Release(); err != nil {
			return err
		}
	} else {
		service.Start(uint16(p), c.GlobalString("db"))
	}

	return nil
}
