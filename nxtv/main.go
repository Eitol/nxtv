package main

import (
	"encoding/json"
	"fmt"
	"github.com/Eitol/nxtv/pkg/nxtv"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:        "nxtv",
		Usage:       "Conventional commit tool",
		Description: "it allows to get the next version of the release",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "path",
				Aliases:   []string{"p"},
				Usage:     "path to git repository. i.e: /home/user/repo",
				TakesFile: true,
				Required:  true,
			},
			&cli.StringFlag{
				Name:     "sourceBranch",
				Aliases:  []string{"s"},
				Usage:    "source branch. i.e: 'feature/foo",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "targetBranch",
				Aliases: []string{"t"},
				Usage:   "target branch. i.e: 'main",
			},
		},
		Action: func(c *cli.Context) error {
			path := c.String("path")
			sourceBranch := c.String("sourceBranch")
			targetBranch := c.String("targetBranch")
			output, err := nxtv.GetNextVersionBasedOnMR(path, sourceBranch, targetBranch)
			var jsonOut []byte
			if err != nil {
				jsonOut, _ = json.MarshalIndent(nxtv.Output{Error: err.Error()}, "", "  ")
			} else {
				jsonOut, _ = json.MarshalIndent(output, "", "  ")
			}
			fmt.Printf("%s\n", jsonOut)
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
