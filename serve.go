package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/urfave/cli/v2"
)

func serve() *cli.Command {

	var us UserService
	var port int

	return &cli.Command{
		Name:  "serve",
		Usage: "Start the user service web service",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "port",
				Usage:       "Port for serving http user service",
				Required:    false,
				Destination: &port,
				EnvVars:     []string{"USER_SERVICE_PORT"},
				Value:       8091,
			},
			&cli.StringFlag{
				Name:        "emailHeader",
				Required:    false,
				Destination: &us.HeaderDefs.Email,
				EnvVars:     []string{"SHIB_HEADER_EMAIL"},
				Value:       DefaultShibHeaders.Email,
			},
			&cli.StringFlag{
				Name:     "locatorHeaders",
				Usage:    "comma-separated list of headers to use as locators",
				Required: false,
				EnvVars:  []string{"SHIB_HEADERS_LOCATOR"},
				Value:    strings.Join(DefaultShibHeaders.LocatorIDs, ","),
			},
		},
		Action: func(c *cli.Context) error {
			us.HeaderDefs.LocatorIDs = strings.Split(c.String("locatorHeaders"), ",")

			return serveAction(us, port)
		},
	}
}

func serveAction(us UserService, port int) error {
	http.Handle("/whoami", httpUserService(us))
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	log.Printf("Listening on port %d", listener.Addr().(*net.TCPAddr).Port)

	return http.Serve(listener, nil)
}
