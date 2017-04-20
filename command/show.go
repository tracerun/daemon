package command

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tracerun/tracerun/db/action"
	"github.com/tracerun/tracerun/db/segment"
	"github.com/urfave/cli"
)

// NewShowCMD to query db.
func NewShowCMD() cli.Command {
	return cli.Command{
		Name:   "show",
		Usage:  "show content in db",
		Action: showAction,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "actions",
				Usage: "List all actions.",
			},
			cli.BoolFlag{
				Name:  "targets",
				Usage: "List all actions.",
			},
			cli.BoolFlag{
				Name:  "json",
				Usage: "Show result with JSON.",
			},
		},
	}
}

// Action struct for a single action
type Action struct {
	Target string
	Start  uint32
	Last   uint32
}

func showAction(c *cli.Context) error {
	jsonFormat := c.Bool("json")

	if c.Bool("actions") {
		targets, starts, lasts, err := action.GetAll()
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		if jsonFormat {
			var actions []Action
			for i := 0; i < len(targets); i++ {
				actions = append(actions, Action{
					Target: targets[i],
					Start:  starts[i],
					Last:   lasts[i],
				})
			}
			b, err := json.Marshal(map[string][]Action{"actions": actions})
			if err != nil {
				return cli.NewExitError(err, 1)
			}
			fmt.Println(string(b))
		} else {
			fmt.Println("actions:")
			for i := 0; i < len(targets); i++ {
				fmt.Printf("  %s:\n", targets[i])
				sT := time.Unix(int64(starts[i]), 0)
				lT := time.Unix(int64(lasts[i]), 0)
				fmt.Printf("    %s\t\t%s\n", sT.Format("2006-01-02 15:04:05"), lT.Format("2006-01-02 15:04:05"))
			}
		}
	} else if c.Bool("targets") {
		targets, err := segment.GetTargets()
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		if jsonFormat {
			b, err := json.Marshal(map[string][]string{"targets": targets})
			if err != nil {
				return cli.NewExitError(err, 1)
			}
			fmt.Println(string(b))
		} else {
			fmt.Println("targets:")
			for i := 0; i < len(targets); i++ {
				fmt.Printf("  %s\n", targets[i])
			}
		}
	}

	return nil
}
