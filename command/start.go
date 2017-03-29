package command

import (
	"tracerun/grpcd"

	"github.com/urfave/cli"
)

// NewStartCMD create a start command.
func NewStartCMD() cli.Command {
	return cli.Command{
		Name:   "start",
		Usage:  "start grpc service",
		Action: action,
	}
}

func action(c *cli.Context) error {
	grpcd.Start()
	return nil
}
