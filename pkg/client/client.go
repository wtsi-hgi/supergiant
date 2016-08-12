package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Client struct {
	BaseURL  string
	Username string
	Password string

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
}

func New(url string, user string, pass string) *Client {
	client := &Client{
		BaseURL:  url,
		Username: user,
		Password: pass,
	}

	client.CloudAccounts = &CloudAccounts{Collection{client, "cloud_accounts"}}
	client.Kubes = &Kubes{Collection{client, "kubes"}}
	client.Apps = &Apps{Collection{client, "apps"}}
	client.Components = &Components{Collection{client, "components"}}
	client.Releases = &Releases{Collection{client, "releases"}}
	client.Instances = &Instances{Collection{client, "instances"}}
	client.Volumes = &Volumes{Collection{client, "volumes"}}
	client.PrivateImageKeys = &PrivateImageKeys{Collection{client, "private_image_keys"}}
	client.Entrypoints = &Entrypoints{Collection{client, "entrypoints"}}
	client.Nodes = &Nodes{Collection{client, "nodes"}}

	return client
}

func (c *Client) request(method string, path string, in interface{}, out interface{}, queryValues map[string]string) error {
	body := new(bytes.Buffer)
	if in != nil {
		if err := json.NewEncoder(body).Encode(in); err != nil {
			return err
		}
	}

	requestURL, err := url.Parse(c.BaseURL + "/" + path)
	if err != nil {
		return err
	}

	q := requestURL.Query()
	for key, value := range queryValues {
		q.Set(key, value)
	}
	requestURL.RawQuery = q.Encode()

	req, err := http.NewRequest(method, requestURL.String(), body)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.Username, c.Password)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return err
	}

	if resp.Status[:2] != "20" {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(body))
	}

	if out != nil {
		defer resp.Body.Close()
		if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
			return err
		}
	}

	return nil
}
