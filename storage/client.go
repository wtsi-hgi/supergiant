package storage

import (
	"fmt"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

const (
	baseDir = "/supergiant"
)

type Client struct {
	endpoints         []string
	kapi              etcd.KeysAPI
	AppStorage        *AppStorage
	ImageRepoStorage  *ImageRepoStorage
	ComponentStorage  *ComponentStorage
	DeploymentStorage *DeploymentStorage
	InstanceStorage   *InstanceStorage
	ReleaseStorage    *ReleaseStorage
	JobStorage        *JobStorage
}

func NewClient(endpoints []string) *Client {
	c := Client{endpoints: endpoints}

	// c.createBaseDirectory()

	c.AppStorage = c.newAppStorage()
	c.ImageRepoStorage = c.newImageRepoStorage()
	c.ComponentStorage = c.newComponentStorage()
	c.DeploymentStorage = c.newDeploymentStorage()
	c.InstanceStorage = c.newInstanceStorage()
	c.ReleaseStorage = c.newReleaseStorage()
	c.JobStorage = c.newJobStorage()
	return &c
}

// func (c *Client) createBaseDirectory() {
// 	if _, err := c.Get(""); err != nil {
// 		if _, err := c.CreateDirectory(""); err != nil {
// 			panic(err)
// 		}
// 	}
// }

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

func fullKey(key string) string {
	return fmt.Sprintf("%s%s", baseDir, key)
}

// func (c *Client) CreateDirectory(key string) (*etcd.Response, error) {
// 	return c.keysAPI().Set(context.Background(), fullKey(key), "", &etcd.SetOptions{Dir: true})
// }

func (c *Client) CompareAndSwap(key string, prevValue string, value string) (*etcd.Response, error) {
	return c.keysAPI().Set(context.Background(), fullKey(key), value, &etcd.SetOptions{PrevValue: prevValue})
}

func (c *Client) CreateInOrder(dir string, value string) (*etcd.Response, error) {
	return c.keysAPI().CreateInOrder(context.Background(), fullKey(dir), value, nil)
}

func (c *Client) GetInOrder(dir string) (*etcd.Response, error) {
	return c.keysAPI().Get(context.Background(), fullKey(dir), &etcd.GetOptions{Sort: true})
}

func (c *Client) Create(key string, value string) (*etcd.Response, error) {
	return c.keysAPI().Create(context.Background(), fullKey(key), value)
}

func (c *Client) Get(key string) (*etcd.Response, error) {
	return c.keysAPI().Get(context.Background(), fullKey(key), nil)
}

func (c *Client) Update(key string, value string) (*etcd.Response, error) {
	return c.keysAPI().Update(context.Background(), fullKey(key), value)
}

func (c *Client) Delete(key string) (*etcd.Response, error) {
	return c.keysAPI().Delete(context.Background(), fullKey(key), nil)
}

func (c *Client) newAppStorage() *AppStorage {
	return &AppStorage{client: c}
}

func (c *Client) newImageRepoStorage() *ImageRepoStorage {
	return &ImageRepoStorage{client: c}
}

func (c *Client) newComponentStorage() *ComponentStorage {
	return &ComponentStorage{client: c}
}

func (c *Client) newDeploymentStorage() *DeploymentStorage {
	return &DeploymentStorage{client: c}
}

func (c *Client) newInstanceStorage() *InstanceStorage {
	return &InstanceStorage{client: c}
}

func (c *Client) newReleaseStorage() *ReleaseStorage {
	return &ReleaseStorage{client: c}
}

func (c *Client) newJobStorage() *JobStorage {
	return &JobStorage{client: c}
}
