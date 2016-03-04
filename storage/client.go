package storage

import (
	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type Client struct {
	Endpoints []string
	kapi      etcd.KeysAPI
}

func (c *Client) keysAPI() etcd.KeysAPI {
	if c.kapi == nil {
		cfg := etcd.Config{Endpoints: c.Endpoints}
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
