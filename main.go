package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/supergiant/supergiant/pkg/api"
	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/ui"
)

func main() {
	app := cli.NewApp()
	app.Name = "supergiant"
	app.Usage = "Supergiant"

	c := new(core.Core)

	app.Action = func(ctx *cli.Context) {

		requiredFlags := map[string]string{
			"psql-host":       c.PsqlHost,
			"psql-db":         c.PsqlDb,
			"psql-user":       c.PsqlUser,
			"psql-pass":       c.PsqlPass,
			"http-port":       c.HTTPPort,
			"http-basic-user": c.HTTPBasicUser,
			"http-basic-pass": c.HTTPBasicPass,
		}

		for flag, val := range requiredFlags {
			if val == "" {
				cli.ShowCommandHelp(ctx, fmt.Sprintf("%s required\n", flag))
				os.Exit(5)
			}
		}

		c.Initialize()

		// We do this here, and not in core, so that we can ensure the file closes on exit.
		if c.LogPath != "" {
			file, err := os.OpenFile(c.LogPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			path, err := filepath.Abs(c.LogPath)
			if err != nil {
				panic(err)
			}
			fmt.Println("Writing log to " + path)
			c.Log.Out = file
		}

		apiRouter := api.NewRouter(c)
		router := ui.NewRouter(c.NewAPIClient(), apiRouter)

		c.Log.Info(fmt.Sprintf(":%s/api/v0", c.HTTPPort))
		c.Log.Info(fmt.Sprintf(":%s/ui", c.HTTPPort))
		c.Log.Info(http.ListenAndServe(fmt.Sprintf(":%s", c.HTTPPort), router))
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "psql-host",
			Usage:       "PostgreSQL host",
			Destination: &c.PsqlHost,
		},
		cli.StringFlag{
			Name:        "psql-db",
			Usage:       "PostgreSQL database name",
			Destination: &c.PsqlDb,
		},
		cli.StringFlag{
			Name:        "psql-user",
			Usage:       "PostgreSQL database user",
			Destination: &c.PsqlUser,
		},
		cli.StringFlag{
			Name:        "psql-pass",
			Usage:       "PostgreSQL database password",
			Destination: &c.PsqlPass,
		},
		cli.StringFlag{
			Name:        "http-port",
			Usage:       "HTTP port for the web interfaces",
			Destination: &c.HTTPPort,
		},
		cli.StringFlag{
			Name:        "http-basic-user",
			Usage:       "HTTP Basic Auth username used to secure the web interfaces",
			Destination: &c.HTTPBasicUser,
		},
		cli.StringFlag{
			Name:        "http-basic-pass",
			Usage:       "HTTP Basic Auth password used to secure the web interfaces",
			Destination: &c.HTTPBasicPass,
		},
		cli.StringFlag{
			Name:        "log-file",
			Usage:       "Log output filepath",
			Destination: &c.LogPath,
		},
		cli.StringFlag{
			Name:        "log-level",
			Usage:       "Log level",
			Destination: &c.LogLevel,
			// Value:  <--- NOTE just cuz you always forget you can set defaults
		},
	}

	app.Run(os.Args)
}
