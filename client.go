package core

import (
	"encoding/json"
	"path"

	etcd "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

// Entity interface is for CRUD resources
type Entity interface {
	Unmarshal(n *Node) error
}

// EntityList is for lists of CRUD resources
type EntityList interface {
	NewEntity() Entity
}

// Node is a wrapper around etcd client.Node
type Node struct {
	*etcd.Node
}

func (n *Node) ValueInto(e Entity) error {
	return json.Unmarshal([]byte(n.Value), e)
}

// Response is a wrapper around etcd client.Response.
// It also holds errors, to allow chaining methods to build up a request, and
// returning only once at the end.
type Response struct {
	raw *etcd.Response
	err error
}

// ListResponse is a wrapper around etcd client.Response for directories
type ListResponse struct {
	raw *etcd.Response
	err error
}

func (r *Response) Into(e Entity) error {
	return e.Unmarshal(&Node{r.raw.Node})
}

func (r *ListResponse) Into(l EntityList) error {
	nodes := r.raw.Node.Nodes
	for _, node := range nodes {
		if err := l.NewEntity().Unmarshal(&Node{node}); err != nil {
			return err
		}
	}
	return nil
}

type ClientInterface interface {
	Create(resource string, name string, value string) *Response
	List(resource string) *ListResponse
	Get(resource string, name string) *Response
	Update(resource string, name string, value string) *Response
	Destroy(resource string, name string) *Response
}

type Client struct {
	KAPI etcd.KeysAPI
}

func (c *Client) keysAPI() etcd.KeysAPI {
	if c.KAPI == nil {
		// c.KAPI = c.etcd.NewKeysAPI(c.etcd)
	}
	return c.KAPI
}

func (c *Client) createKey(key string, value string) (*etcd.Response, error) {
	return c.keysAPI().Create(context.Background(), key, value)
}

func (c *Client) getKey(key string) (*etcd.Response, error) {
	return c.keysAPI().Get(context.Background(), key, nil)
}

func (c *Client) setKey(key string, value string) (*etcd.Response, error) {
	return c.keysAPI().Set(context.Background(), key, value, nil)
}

func (c *Client) updateKey(key string, value string) (*etcd.Response, error) {
	return c.keysAPI().Update(context.Background(), key, value)
}

func (c *Client) deleteKey(key string) (*etcd.Response, error) {
	return c.keysAPI().Delete(context.Background(), key, nil)
}

func resourcePath(resource string, name string) string {
	return path.Join(resource, name)
}

func (c *Client) Create(resource string, name string, value string) *Response {
	resp, err := c.createKey(resourcePath(resource, name), value)
	return &Response{resp, err}
}

func (c *Client) List(resource string) *ListResponse {
	resp, err := c.getKey(resourcePath(resource, ""))
	return &ListResponse{resp, err}
}

func (c *Client) Get(resource string, name string) *Response {
	resp, err := c.getKey(resourcePath(resource, name))
	return &Response{resp, err}
}

// NOTE / TODO this is not responsible for deep-merging with existing values,
// since this client only deals with values as strings. There needs to be some
// middle piece that is responsible for composition / serialization.
func (c *Client) Update(resource string, name string, value string) *Response {
	resp, err := c.updateKey(resourcePath(resource, name), value)
	return &Response{resp, err}
}

func (c *Client) Destroy(resource string, name string) *Response {
	resp, err := c.deleteKey(resourcePath(resource, name))
	return &Response{resp, err}
}
