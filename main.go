package main

import (
	"net/http"
	"os"

	"github.com/codegangsta/cli"
	"github.com/supergiant/supergiant/api"
	"github.com/supergiant/supergiant/core"
)

func main() {

	app := cli.NewApp()
	app.Name = "supergiant-api"
	app.Usage = "The Supergiant api server."

	c := new(core.Core)

	app.Action = func(ctx *cli.Context) {

		core.SetLogLevel(ctx.String("log-level"))

		// Check the args. The ones we don't have default values for...
		if c.K8sUser == "" {
			core.Log.Error("Kubernetes HTTP basic username required")
			cli.ShowCommandHelp(ctx, "")
			os.Exit(5)
		}
		if c.K8sPass == "" {
			core.Log.Error("Kubernetes HTTP basic password required")
			cli.ShowCommandHelp(ctx, "")
			os.Exit(5)
		}

		c.EtcdEndpoints = ctx.StringSlice("etcd-hosts")
		if len(c.EtcdEndpoints) < 0 {
			c.EtcdEndpoints = []string{"http://etcd:2379"}
		}

		// Log args that have default values.
		core.Log.Info("ETCD hosts,", c.EtcdEndpoints)
		core.Log.Info("Kubernetes Host,", c.K8sHost)

		c.Initialize()

		router := api.NewRouter(c)

		core.Log.Info("Serving API on port :8080")
		core.Log.Info(http.ListenAndServe(":8080", router))
	}

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:   "etcd-hosts",
			Usage:  "Array of etcd hosts.",
			EnvVar: "ETCD_ENDPOINT",
		},
		cli.StringFlag{
			Name:        "k8s-host, kh",
			Value:       "kubernetes", // TODO is this working?
			Usage:       "IP of a Kuberntes api.",
			EnvVar:      "K8S_HOST",
			Destination: &c.K8sHost,
		},
		cli.StringFlag{
			Name:        "k8s-user, ku",
			Usage:       "Username used to connect to your Kubernetes api.",
			EnvVar:      "K8S_USER",
			Destination: &c.K8sUser,
		},
		cli.StringFlag{
			Name:        "k8s-pass, kp",
			Usage:       "Password used to connect to your Kubernetes api.",
			EnvVar:      "K8S_PASS",
			Destination: &c.K8sPass,
		},
		cli.StringFlag{
			Name:        "aws-region, ar",
			Usage:       "AWS Region in which your kubernetes cluster resides.",
			EnvVar:      "AWS_REGION",
			Destination: &c.AwsRegion,
		},
		cli.StringFlag{
			Name:        "aws-az, az",
			Usage:       "AWS Availability Zone in which your kubernetes cluster resides.",
			EnvVar:      "AWS_AZ",
			Destination: &c.AwsAZ,
		},
		cli.StringFlag{
			Name:        "aws-sg-id, sg",
			Usage:       "AWS Security Group in which your kubernetes cluster resides.",
			EnvVar:      "AWS_SG_ID",
			Destination: &c.AwsSgID,
		},
		cli.StringFlag{
			Name:        "aws-subnet-id, sid",
			Usage:       "AWS Subnet ID in which your kubernetes cluster resides.",
			EnvVar:      "AWS_SUBNET_ID",
			Destination: &c.AwsSubnetID,
		},
		cli.StringFlag{
			Name:        "aws-access-key",
			Usage:       "AWS Access key.",
			Destination: &c.AwsAccessKey,
		},
		cli.StringFlag{
			Name:        "aws-secret-key",
			Usage:       "AWS Secret key.",
			Destination: &c.AwsSecretKey,
		},
		cli.BoolFlag{
			Name:        "k8s-insecure-https",
			Usage:       "Skip verification if HTTPS mode when connecting to Kubernetes.",
			Destination: &c.K8sInsecureHTTPS,
		},
		cli.BoolFlag{
			Name:        "enable-capacity-service",
			Usage:       "Enable the automatic creation/deletion of servers to meet requested capacity.",
			Destination: &c.CapacityServiceEnabled,
		},
		cli.StringFlag{
			Name:  "log-level",
			Value: "info",
			Usage: "Set the API log level",
		},
	}

	app.Run(os.Args)
}
