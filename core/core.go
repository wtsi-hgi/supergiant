package core

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/supergiant/guber"
	"github.com/supergiant/supergiant/common"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
)

type Core struct {
	EtcdEndpoints          []string
	K8sHost                string
	K8sUser                string
	K8sPass                string
	K8sInsecureHTTPS       bool
	AwsRegion              string
	AwsAZ                  string
	AwsSgID                string
	AwsSubnetID            string
	AwsAccessKey           string
	AwsSecretKey           string
	CapacityServiceEnabled bool

	db          *database
	k8s         guber.Client
	ec2         *ec2.EC2
	elb         elbiface.ELBAPI
	autoscaling autoscalingiface.AutoScalingAPI
}

var (
	Log = logrus.New()
)

// TODO inconsistent with method in Guber and client/
func SetLogLevel(level string) {
	levelInt, err := logrus.ParseLevel(level)
	if err != nil {
		panic(err)
	}
	Log.Level = levelInt

	// guber.Log.SetLevel(level)
}

// NOTE this used to be core.New(), but due to how we load in values from the
// cli package, I needed to first actually initialize a Core struct and then
// configure.
func (c *Core) Initialize() {
	c.db = newDB(c.EtcdEndpoints)
	c.k8s = guber.NewClient(c.K8sHost, c.K8sUser, c.K8sPass, c.K8sInsecureHTTPS)

	checkForAWSMeta(c)
	// If you're working with temporary security credentials,
	// you can also keep the session token in AWS_SESSION_TOKEN.
	// TODO: We need to set this up when we have more timez
	token := ""

	creds := credentials.NewStaticCredentials(c.AwsAccessKey, c.AwsSecretKey, token)
	_, err := creds.Get()
	if err != nil {
		Log.Error("AWS Credentials Install Failed...", err)
	}
	Log.Info("AWS Credentials Installed.")

	awsSession := session.New()
	awsConf := aws.NewConfig().WithRegion(c.AwsRegion).WithCredentials(creds)

	c.ec2 = ec2.New(awsSession, awsConf)
	c.elb = elb.New(awsSession, awsConf)
	c.autoscaling = autoscaling.New(awsSession, awsConf)

	// TODO expose as worker num option in main
	go NewSupervisor(c, 4).Run()

	if c.CapacityServiceEnabled {
		go newCapacityService(c).Run()
	}

	// TODO
	if err := c.Nodes().populate(); err != nil {
		panic(err)
	}
}

// Core implements the Locatable interface, but is only utilized for the child()
// method, since it is the root object and does not have a locationKey() or parent().
func (c *Core) locationKey() (k string) {
	return
}

func (c *Core) parent() (l Locatable) {
	return
}

func (c *Core) child(key string) (l Locatable) {
	switch key {
	case "apps":
		l = c.Apps().(Locatable)
	case "entrypoints":
		l = c.Entrypoints().(Locatable)
	case "registries":
		l = c.ImageRegistries().(Locatable)
	case "nodes":
		l = c.Nodes().(Locatable)
	case "tasks":
		l = c.Tasks().(Locatable)
	default:
		panic(fmt.Errorf("No child with key %s for %T", key, c))
	}
	return
}

// Top-level resources

func (c *Core) Apps() AppsInterface {
	return &AppCollection{c}
}

func (c *Core) Entrypoints() EntrypointsInterface {
	return &EntrypointCollection{c}
}

func (c *Core) ImageRegistries() ImageRegistriesInterface {
	return &ImageRegistryCollection{c}
}

// TODO this goes away when Dockerhub is not the only ImageRegistry
func (c *Core) ImageRepos() ImageReposInterface {
	registry := c.ImageRegistries().New()
	registry.Name = common.IDString("dockerhub")
	return registry.ImageRepos()
}

func (c *Core) Nodes() NodesInterface {
	return &NodeCollection{c}
}

func (c *Core) Tasks() TasksInterface {
	return &TaskCollection{c}
}
