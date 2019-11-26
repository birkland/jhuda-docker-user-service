package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "JHUDA user service",
		Usage: "Provides an http endpoint for determining user info based on shibboleth headers",
		Commands: []*cli.Command{
			serve(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
