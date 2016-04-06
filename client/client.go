package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type Client struct {
	baseURL string
	// Host string
	Username string
	Password string
	http     *http.Client
}

func New(url string, user string, pass string, verify bool) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: verify},
	}
	return &Client{url, user, pass, &http.Client{Transport: tr}}
}

// Non-Client misc
//==============================================================================
func serialize(in interface{}) (*bytes.Buffer, error) {
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(data), nil
}

func deserialize(in io.ReadCloser, out interface{}) error {
	defer in.Close()
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

// Util
//==============================================================================
func (c *Client) url(path string) string {
	return c.baseURL + "/" + path
}

func (c *Client) request(method string, path string, in interface{}, out interface{}) error {
	body := new(bytes.Buffer)
	if in != nil {
		buff, err := serialize(in)
		if err != nil {
			return err
		}
		body = buff
	}

	req, err := http.NewRequest(method, c.url(path), body)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	if status := resp.Status; status[:2] != "20" {
		return fmt.Errorf("Request failed with status %s", status)
	}

	if out != nil {
		if err = deserialize(resp.Body, out); err != nil {
			return err
		}
	}

	return nil
}

// Request methods
//==============================================================================
func (c *Client) Get(path string, out interface{}) error {
	return c.request("GET", path, nil, out)
}

func (c *Client) Post(path string, in interface{}, out interface{}) error {
	return c.request("POST", path, in, out)
}

func (c *Client) Delete(path string) error {
	return c.request("DELETE", path, nil, nil)
}

// Resources
//==============================================================================
func (c *Client) Apps() *AppCollection {
	return &AppCollection{c}
}

func (c *Client) Entrypoints() *EntrypointCollection {
	return &EntrypointCollection{c}
}
