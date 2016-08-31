package core

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/supergiant/supergiant/pkg/client"
	"github.com/supergiant/supergiant/pkg/models"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type NodeSize struct {
	Name     string  `json:"name"`
	RAMGIB   float64 `json:"ram_gib"`
	CPUCores float64 `json:"cpu_cores"`
}

type Settings struct {
	PsqlHost      string `json:"psql_host"`
	PsqlDb        string `json:"psql_db"`
	PsqlUser      string `json:"psql_user"`
	PsqlPass      string `json:"psql_pass"`
	HTTPBasicUser string `json:"http_basic_user"`
	HTTPBasicPass string `json:"http_basic_pass"`
	PublishHost   string `json:"publish_host"`
	HTTPPort      string `json:"http_port"`
	HTTPSPort     string `json:"https_port"`
	SSLCertFile   string `json:"ssl_cert_file"`
	SSLKeyFile    string `json:"ssl_key_file"`
	LogPath       string `json:"log_file"`
	LogLevel      string `json:"log_level"`

	// NOTE these MUST be provided in ascending order by cost in order to
	// correctly provision the smallest size on Kube creation
	//
	// NodeSizes is a map of provider name (ex. "aws") and node sizes
	NodeSizes map[string][]*NodeSize `json:"node_sizes"`
}

type Core struct {
	Settings

	Log *logrus.Logger

	DB *DB

	CloudAccounts    *CloudAccounts
	Kubes            *Kubes
	Apps             *Apps
	Components       *Components
	Releases         *Releases
	Instances        *Instances
	Volumes          *Volumes
	PrivateImageKeys *PrivateImageKeys
	Entrypoints      *Entrypoints
	Nodes            *Nodes

	Actions              map[string]*Action
	actionRequestChannel chan *actionRequest
}

type actionRequestType int

const (
	requestActionFetch actionRequestType = iota
	requestActionStart
	requestActionStop
)

type actionRequest struct {
	returnChannel chan *Action
	requestType   actionRequestType
	action        *Action
}

// NOTE this used to be core.New(), but due to how we load in values from the
// cli package, I needed to first actually initialize a Core struct and then
// configure.
func (c *Core) Initialize() {
	// Log
	c.Log = logrus.New()

	if c.LogLevel != "" {
		levelInt, err := logrus.ParseLevel(c.LogLevel)
		if err != nil {
			panic(err)
		}
		c.Log.Level = levelInt
	}

	// db.LogMode(true)
	// guber.Log.SetLevel("debug")

	// Database
	options := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable", c.PsqlHost, c.PsqlDb, c.PsqlUser, c.PsqlPass)
	db, err := gorm.Open("postgres", options)
	if err != nil {
		panic(err)
	}
	c.DB = &DB{c, db}
	err = c.DB.AutoMigrate(
		&models.Kube{},
		&models.CloudAccount{},
		&models.PrivateImageKey{},
		&models.App{},
		&models.Component{},
		&models.ComponentPrivateImageKey{},
		&models.Release{},
		&models.Instance{},
		&models.Volume{},
		&models.Entrypoint{},
		&models.Node{},
	).Error
	if err != nil {
		panic(err)
	}

	c.Kubes = &Kubes{Collection{c}}
	c.CloudAccounts = &CloudAccounts{Collection{c}}
	c.Apps = &Apps{Collection{c}}
	c.Components = &Components{Collection{c}}
	c.Releases = &Releases{Collection{c}}
	c.Instances = &Instances{Collection{c}}
	c.Volumes = &Volumes{Collection{c}}
	c.PrivateImageKeys = &PrivateImageKeys{Collection{c}}
	c.Entrypoints = &Entrypoints{Collection{c}}
	c.Nodes = &Nodes{Collection{c}}

	// Initialize Actions map
	c.Actions = make(map[string]*Action)

	// Can't have concurrent map access...
	c.actionRequestChannel = make(chan *actionRequest)
	go func() {
		for {
			r := <-c.actionRequestChannel

			switch r.requestType {
			case requestActionFetch:
				// Handled below

			case requestActionStart:
				c.Log.Infof("Begin  : %s", r.action.description())
				c.Actions[r.action.resourceID] = r.action

			case requestActionStop:
				c.Log.Infof("End    : %s", r.action.description())
				delete(c.Actions, r.action.resourceID)
			}

			r.returnChannel <- c.Actions[r.action.resourceID]
		}
	}()

	// Recurring services
	capacityService := &RecurringService{
		core:     c,
		service:  &CapacityService{c},
		interval: 30 * time.Second,
	}
	nodeObserver := &RecurringService{
		core:     c,
		service:  &NodeObserver{c},
		interval: 30 * time.Second,
	}
	instanceObserver := &RecurringService{
		core:     c,
		service:  &InstanceObserver{c},
		interval: 30 * time.Second,
	}

	go capacityService.Run()
	go nodeObserver.Run()
	go instanceObserver.Run()
}

//------------------------------------------------------------------------------

func (c *Core) SSLEnabled() bool {
	return c.HTTPSPort != "" && c.SSLCertFile != "" && c.SSLKeyFile != ""
}

func (c *Core) BaseURL() string {
	var protocol string
	var port string
	if c.SSLEnabled() {
		protocol = "https"
		port = c.HTTPSPort
	} else {
		protocol = "http"
		port = c.HTTPPort
	}
	return fmt.Sprintf("%s://%s:%s", protocol, c.PublishHost, port)
}

func (c *Core) NewAPIClient() *client.Client {
	return client.New(fmt.Sprintf("%s/api/v0", c.BaseURL()), c.HTTPBasicUser, c.HTTPBasicPass, c.SSLCertFile)
}
