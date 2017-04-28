package command

import (
	"github.com/tracerun/clientgo"
	"github.com/urfave/cli"
)

// NewAddCMD create a add command. Used to add actions.
func NewAddCMD() cli.Command {
	return cli.Command{
		Name:   "add",
		Usage:  "add an action to db",
		Action: addAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "target, t",
				Usage: "Target for the action",
			},
			cli.StringFlag{
				Name:  "addr",
				Usage: "Address that need to connect",
				Value: "127.0.0.1",
			},
		},
	}
}

func addAction(c *cli.Context) error {
	target := c.String("target")
	if target == "" {
		return cli.NewExitError("missing target, -h help", 2)
	}

	addr := c.String("addr")
	p := uint16(c.GlobalUint("p"))
	client, exist, err := clientgo.NewSendClient(p, addr)
	if err != nil {
		return cli.NewExitError(err, 2)
	}
	if !exist {
		return cli.NewExitError("service unavailable", 2)
	}

	return client.SendAction(target)
}
