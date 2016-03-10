package storage

import (
	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type Client struct {
	endpoints         []string
	kapi              etcd.KeysAPI
	AppStorage        *AppStorage
	ComponentStorage  *ComponentStorage
	DeploymentStorage *DeploymentStorage
	InstanceStorage   *InstanceStorage
	ReleaseStorage    *ReleaseStorage
}

func NewClient(endpoints []string) *Client {
	c := Client{endpoints: endpoints}
	c.AppStorage = c.newAppStorage()
	c.ComponentStorage = c.newComponentStorage()
	c.DeploymentStorage = c.newDeploymentStorage()
	c.InstanceStorage = c.newInstanceStorage()
	c.ReleaseStorage = c.newReleaseStorage()
	return &c
}

func (c *Client) keysAPI() etcd.KeysAPI {
	if c.kapi == nil {
		cfg := etcd.Config{Endpoints: c.endpoints}
		etcdClient, err := etcd.New(cfg)
		if err != nil {
			panic(err)
		}
		c.kapi = etcd.NewKeysAPI(etcdClient)
	}
	return c.kapi
}

func (c *Client) CreateDirectory(key string) (*etcd.Response, error) {
	return c.keysAPI().Set(context.Background(), key, "", &etcd.SetOptions{Dir: true})
}

func (c *Client) Create(key string, value string) (*etcd.Response, error) {
	return c.keysAPI().Create(context.Background(), key, value)
}

func (c *Client) Get(key string) (*etcd.Response, error) {
	return c.keysAPI().Get(context.Background(), key, nil)
}

func (c *Client) Update(key string, value string) (*etcd.Response, error) {
	return c.keysAPI().Update(context.Background(), key, value)
}

func (c *Client) Delete(key string) (*etcd.Response, error) {
	return c.keysAPI().Delete(context.Background(), key, nil)
}

func (c *Client) newAppStorage() *AppStorage {
	appStorage := AppStorage{client: c}
	appStorage.CreateBaseDirectory()
	return &appStorage
}

func (c *Client) newComponentStorage() *ComponentStorage {
	compStorage := ComponentStorage{client: c}
	compStorage.CreateBaseDirectory()
	return &compStorage
}

func (c *Client) newDeploymentStorage() *DeploymentStorage {
	deploymentStorage := DeploymentStorage{client: c}
	deploymentStorage.CreateBaseDirectory()
	return &deploymentStorage
}

func (c *Client) newInstanceStorage() *InstanceStorage {
	instanceStorage := InstanceStorage{client: c}
	instanceStorage.CreateBaseDirectory()
	return &instanceStorage
}

func (c *Client) newReleaseStorage() *ReleaseStorage {
	releaseStorage := ReleaseStorage{client: c}
	releaseStorage.CreateBaseDirectory()
	return &releaseStorage
}
