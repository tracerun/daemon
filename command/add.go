package command

import "github.com/urfave/cli"
import "tracerun/db/action"

// NewAddCMD create a add command. Used to add actions.
func NewAddCMD() cli.Command {
	return cli.Command{
		Name:   "add",
		Usage:  "add an action to db",
		Action: addAction,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "c",
				Usage: "A close action. Default for active action.",
			},
			cli.StringFlag{
				Name:  "target, t",
				Usage: "Target for the action",
			},
		},
	}
}

func addAction(c *cli.Context) error {
	target := c.String("target")
	if target == "" {
		return cli.NewExitError("missing target, -h help", 2)
	}
	if c.Bool("c") {
		action.AddToDB(target, false)
	} else {
		action.AddToDB(target, true)
	}
	return nil
}
