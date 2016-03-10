package storage

import (
	"encoding/json"
	"fmt"
	"supergiant/core/model"
)

type AppStorage struct {
	client *Client
}

// TODO
func (store *AppStorage) CreateBaseDirectory() {
	if _, err := store.client.Get("/apps"); err != nil {
		if _, err := store.client.CreateDirectory("/apps"); err != nil {
			panic(err)
		}
	}
}

func (store *AppStorage) Create(e *model.App) (*model.App, error) {
	key := fmt.Sprintf("/apps/%s", e.Name)
	value, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	if _, err = store.client.Create(key, string(value)); err != nil {
		return nil, err
	}

	// Create all the other base dirs
	_, err = store.client.CreateDirectory(fmt.Sprintf("components/%s", e.Name))
	_, err = store.client.CreateDirectory(fmt.Sprintf("releases/%s", e.Name))
	return e, err
}

func (store *AppStorage) List() ([]*model.App, error) {
	key := "/apps"
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}

	apps := make([]*model.App, 0)

	for _, node := range resp.Node.Nodes {
		value := node.Value
		e := new(model.App)
		if err := json.Unmarshal([]byte(value), e); err != nil {
			return nil, err
		}
		apps = append(apps, e)
	}
	return apps, nil
}

func (store *AppStorage) Get(id string) (*model.App, error) {
	// TODO repeated, move to method
	key := fmt.Sprintf("/apps/%s", id)
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}
	value := resp.Node.Value

	e := new(model.App)
	if err := json.Unmarshal([]byte(value), e); err != nil {
		return nil, err
	}
	return e, nil
}

// No update for App

func (store *AppStorage) Delete(id string) error {
	// TODO repeated
	_, err := store.client.Delete(fmt.Sprintf("/apps/%s", id))
	_, err = store.client.Delete(fmt.Sprintf("/components/%s", id))
	_, err = store.client.Delete(fmt.Sprintf("/releases/%s", id))
	return err
}
