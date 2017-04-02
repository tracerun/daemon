package command

import "github.com/urfave/cli"

// NewShowCMD to query db.
func NewShowCMD() cli.Command {
	return cli.Command{
		Name:   "show",
		Usage:  "show content in db",
		Action: showAction,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "a",
				Usage: "List all",
			},
			cli.BoolFlag{
				Name:  "json",
				Usage: "Show result with JSON.",
			},
		},
	}
}

func showAction(c *cli.Context) error {
	// jsonFormat := c.Bool("json")

	return nil
}
