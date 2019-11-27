package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	run(os.Args)
}

func run(args []string) {
	app := &cli.App{
		Name:  "JHUDA user service",
		Usage: "Provides an http endpoint for determining user info based on shibboleth headers",
		Commands: []*cli.Command{
			serve(),
		},
	}

	err := app.Run(args)
	if err != nil {
		log.Fatal(err)
	}
}
