package command

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tracerun/clientgo"
	"github.com/urfave/cli"
)

// NewListCMD to query.
func NewListCMD() cli.Command {
	return cli.Command{
		Name:   "list",
		Usage:  "list content in db",
		Action: listAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "addr",
				Usage: "Address that need to connect",
				Value: "127.0.0.1",
			},
			cli.BoolFlag{
				Name:  "json, j",
				Usage: "Show result with JSON.",
			},
			cli.BoolFlag{
				Name:  "actions, a",
				Usage: "List all actions.",
			},
			cli.BoolFlag{
				Name:  "targets, t",
				Usage: "List all targets available.",
			},
			cli.StringFlag{
				Name:  "slots",
				Usage: "List all slots of a target.",
			},
			cli.UintFlag{
				Name:  "start, s",
				Usage: "The start unixtime to query for slots, 0 for 1970-1-1 00:00:00.",
				Value: uint(0),
			},
			cli.UintFlag{
				Name:  "end, e",
				Usage: "The end unixtime to query for slots, 0 for 2106-2-7 06:28:15.",
				Value: uint(0),
			},
		},
	}
}

// Action struct for a single action
type Action struct {
	Target string `json:"target"`
	Start  uint32 `json:"start"`
	Last   uint32 `json:"last"`
}

// Slot information
type Slot struct {
	Start uint32 `json:"start"`
	Slot  uint32 `json:"slot"`
}

func listAction(c *cli.Context) error {
	jsonFormat := c.Bool("json")

	addr := c.String("addr")
	p := uint16(c.GlobalUint("p"))
	client, exist, err := clientgo.NewExchClient(p, addr)
	if err != nil {
		return cli.NewExitError(err, 2)
	}
	if !exist {
		return cli.NewExitError("service unavailable", 2)
	}

	if c.Bool("actions") {
		// get all action information
		actions, err := getActions(client)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		if jsonFormat {
			printActionsJSON(actions)
		} else {
			printActions(actions)
		}
	} else if c.Bool("targets") {
		// get all target information
		targets, err := client.GetTargets()
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		if jsonFormat {
			printTargetsJSON(targets)
		} else {
			printTargets(targets)
		}
	} else if target := c.String("slots"); len(target) != 0 {
		// get target slots
		start := uint32(c.Uint("start"))
		end := uint32(c.Uint("end"))

		starts, slots, err := client.GetSlots(target, start, end)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		if jsonFormat {
			printSlotsJSON(starts, slots)
		} else {
			printSlots(target, starts, slots)
		}
	}

	return nil
}

func getActions(client *clientgo.ExchClient) ([]Action, error) {
	targets, starts, lasts, err := client.GetActions()
	if err != nil {
		return nil, err
	}

	var actions []Action
	for i := 0; i < len(targets); i++ {
		actions = append(actions, Action{
			Target: targets[i],
			Start:  starts[i],
			Last:   lasts[i],
		})
	}
	return actions, nil
}

func printActions(actions []Action) {
	if actions == nil || len(actions) == 0 {
		fmt.Println("no actions")
		return
	}

	fmt.Println("actions:")
	for i := 0; i < len(actions); i++ {
		fmt.Printf("  %s:\n", actions[i].Target)
		sT := time.Unix(int64(actions[i].Start), 0)
		lT := time.Unix(int64(actions[i].Last), 0)
		fmt.Printf("    %s\t\t%s\n", sT.Format("2006-01-02 15:04:05"), lT.Format("2006-01-02 15:04:05"))
	}
}

func printActionsJSON(actions []Action) {
	b, _ := json.Marshal(map[string][]Action{"actions": actions})
	fmt.Println(string(b))
}

func printTargets(targets []string) {
	if targets == nil || len(targets) == 0 {
		fmt.Println("no targets")
		return
	}

	fmt.Println("targets:")
	for i := 0; i < len(targets); i++ {
		fmt.Printf("  %s\n", targets[i])
	}
}

func printTargetsJSON(targets []string) {
	b, _ := json.Marshal(map[string][]string{"targets": targets})
	fmt.Println(string(b))
}

func printSlots(target string, starts, slots []uint32) {
	if starts == nil || len(starts) == 0 {
		fmt.Println("no slots")
		return
	}

	fmt.Println("slots:")
	for i := 0; i < len(starts); i++ {
		sT := time.Unix(int64(starts[i]), 0)
		fmt.Printf("  %s\t\t%d\n", sT.Format("2006-01-02 15:04:05"), slots[i])
	}
}

func printSlotsJSON(starts, slots []uint32) {
	var slotsInfo []Slot
	for i := 0; i < len(starts); i++ {
		slotsInfo = append(slotsInfo, Slot{
			Start: starts[i],
			Slot:  slots[i],
		})
	}
	b, _ := json.Marshal(map[string][]Slot{"slots": slotsInfo})
	fmt.Println(string(b))
}
