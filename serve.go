package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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
				Name:        "context",
				Usage:       "JSON-LD context URI",
				Required:    false,
				Destination: &us.JsonldContext,
				EnvVars:     []string{"USER_SERVICE_JSONLD_CONTEXT"},
			},
			&cli.StringFlag{
				Name:        "eppnHeader",
				Required:    false,
				Destination: &us.HeaderDefs.Eppn,
				EnvVars:     []string{"SHIB_HEADER_EPPN"},
				Value:       DefaultShibHeaders.Eppn,
			},
			&cli.StringFlag{
				Name:        "displayNameHeader",
				Required:    false,
				Destination: &us.HeaderDefs.Displayname,
				EnvVars:     []string{"SHIB_HEADER_DISPLAYNAME"},
				Value:       DefaultShibHeaders.Displayname,
			},
			&cli.StringFlag{
				Name:        "emailHeader",
				Required:    false,
				Destination: &us.HeaderDefs.Email,
				EnvVars:     []string{"SHIB_HEADER_EMAIL"},
				Value:       DefaultShibHeaders.Email,
			},
			&cli.StringFlag{
				Name:        "givenNameHeader",
				Required:    false,
				Destination: &us.HeaderDefs.GivenName,
				EnvVars:     []string{"SHIB_HEADER_GIVEN_NAME"},
				Value:       DefaultShibHeaders.GivenName,
			},
			&cli.StringFlag{
				Name:        "lastNameHeader",
				Required:    false,
				Destination: &us.HeaderDefs.LastName,
				EnvVars:     []string{"SHIB_HEADER_LAST_NAME"},
				Value:       DefaultShibHeaders.LastName,
			},
			&cli.StringFlag{
				Name:     "locatorHeaders",
				Usage:    "comma-separated list of headers to use as locators",
				Required: false,
				EnvVars:  []string{"SHIB_HEADERS_LOCATOR"},
				Value:    strings.Join(DefaultShibHeaders.LocatorIDs, ","),
			},
			&cli.StringFlag{
				Name:        "userBaseUrl",
				Usage:       "BaseURL for User resources",
				Required:    false,
				Destination: &us.UserBase,
				EnvVars:     []string{"USER_SERVICE_USER_BASEURL"},
			},
		},
		Action: func(c *cli.Context) error {
			us.HeaderDefs.LocatorIDs = strings.Split(c.String("locatorHeaders"), ",")

			return serveAction(us, port)
		},
	}
}

func serveAction(us UserService, port int) error {
	stop := make(chan os.Signal, 1)
	done := make(chan error, 1)
	signal.Notify(stop, os.Interrupt)

	mux := http.NewServeMux()
	mux.Handle("/whoami", httpUserService(us))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		log.Printf("Listening on port %d", port)
		done <- server.ListenAndServe()
	}()

	select {
	case <-stop:
		_ = server.Shutdown(context.Background())
		log.Printf("Goodbye!")
		return nil
	case err := <-done:
		return err
	}
}
