package command

import "github.com/urfave/cli"

// NewAddCMD create a add command. Used to add actions.
func NewAddCMD() cli.Command {
	return cli.Command{
		Name:   "add",
		Usage:  "add an action to db",
		Action: action,
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

func action(c *cli.Context) error {

	return nil
}
