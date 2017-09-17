package main

import (
	"fmt"
  "os"

  "github.com/urfave/cli"
)

func main() {
  app := cli.NewApp()
  app.Name = "cbt"
	app.Usage = "Cranberries unit test Build Tool"
	app.Version = "0.0.1"

		app.Commands = []cli.Command{
			// command config
			{
					Name:    "build",
					Aliases: []string{"b"},
					Usage:   "hello world を表示します",
					Action:  Build,
			},
	}

	app.Before = func(c *cli.Context) error {
		fmt.Println("Build Start. plz wait...")
		return nil
	}

	app.After = func(c *cli.Context) error {
		fmt.Println("Successfuly")
		return nil
	}

	app.Action = func(c *cli.Context) error {
    fmt.Println("Hello CLI!")
    return nil
  }

  app.Run(os.Args)
}

func Build(c *cli.Context) {
	
		// グローバルオプション
		var isDry = c.GlobalBool("dryrun")
		if isDry {
			fmt.Println("this is dry-run")
		}
	
		// パラメータ
		var paramFirst = ""
		if len(c.Args()) > 0 {
			paramFirst = c.Args().First() // c.Args()[0] と同じ意味
		}
	
		fmt.Printf("Hello world! %s\n", paramFirst)
	}