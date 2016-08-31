package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/imdario/mergo"
	"github.com/supergiant/supergiant/pkg/api"
	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/ui"
)

// TODO move
type SecureInfoHandler struct {
	core *core.Core
}

func (s *SecureInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	msg := "<p>Supergiant is running securely at <a href='" + s.core.BaseURL() + "'>" + s.core.BaseURL() + "</a>.</p>"
	msg += "<p>Unless you have provided your own SSL certificate, this will be a self-signed certificate.</p>"
	msg += "<p>If using self-signed, your browser will most likely warn of an insecure connection. <strong>You must manually trust the certificate to utilize SSL.</strong></p>"
	w.Write([]byte(msg))
}

func main() {
	var configFilePath string
	var configFileSettings core.Settings

	c := new(core.Core)

	app := cli.NewApp()
	app.Name = "supergiant"
	app.Usage = "Supergiant"

	app.Action = func(ctx *cli.Context) {
		// Load and parse config file if provided
		if configFilePath != "" {
			configFile, err := os.Open(configFilePath)
			if err != nil {
				panic(err)
			}
			if err := json.NewDecoder(configFile).Decode(&configFileSettings); err != nil {
				panic(err)
			}
		}

		// Merge in command line settings (which overwrite respective config file settings)
		if err := mergo.Merge(&c.Settings, configFileSettings); err != nil {
			panic(err)
		}

		requiredFlags := map[string]string{
			"psql-host":       c.PsqlHost,
			"psql-db":         c.PsqlDb,
			"psql-user":       c.PsqlUser,
			"psql-pass":       c.PsqlPass,
			"publish-host":    c.PublishHost,
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

		if c.SSLEnabled() {
			c.Log.Info("SSL enabled")

			// Secure info page (concurrently)
			go func() {
				c.Log.Info(http.ListenAndServe(fmt.Sprintf(":%s", c.HTTPPort), &SecureInfoHandler{c}))
			}()

			c.Log.Info(fmt.Sprintf(":%s/api/v0", c.HTTPSPort))
			c.Log.Info(fmt.Sprintf(":%s/ui", c.HTTPSPort))
			c.Log.Info(http.ListenAndServeTLS(fmt.Sprintf(":%s", c.HTTPSPort), c.SSLCertFile, c.SSLKeyFile, router))
		} else {
			c.Log.Info(fmt.Sprintf(":%s/api/v0", c.HTTPPort))
			c.Log.Info(fmt.Sprintf(":%s/ui", c.HTTPPort))
			c.Log.Info(http.ListenAndServe(fmt.Sprintf(":%s", c.HTTPPort), router))
		}

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
			Name:        "publish-host",
			Usage:       "Host that can be used to connect to this Supergiant server remotely",
			Destination: &c.PublishHost,
		},
		cli.StringFlag{
			Name:        "http-port",
			Usage:       "HTTP port for the web interfaces",
			Destination: &c.HTTPPort,
		},
		cli.StringFlag{
			Name:        "https-port",
			Usage:       "HTTPS (SSL) port for the web interfaces",
			Destination: &c.HTTPSPort,
		},
		cli.StringFlag{
			Name:        "ssl-cert-file",
			Usage:       "SSL certificate file",
			Destination: &c.SSLCertFile,
		},
		cli.StringFlag{
			Name:        "ssl-key-file",
			Usage:       "SSL key file",
			Destination: &c.SSLKeyFile,
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
		cli.StringFlag{
			Name:        "config-file",
			Usage:       "JSON config filepath (command line arguments will override the values set here)",
			Destination: &configFilePath,
		},
	}

	app.Run(os.Args)
}
