package storage

import (
	"encoding/json"
	"fmt"
	"supergiant/core/model"
)

type ComponentStorage struct {
	client *Client
}

// TODO
func (store *ComponentStorage) CreateBaseDirectory() {
	if _, err := store.client.Get("/components"); err != nil {
		if _, err := store.client.CreateDirectory("/components"); err != nil {
			panic(err)
		}
	}
}

func (store *ComponentStorage) Create(appName string, s *model.Component) (*model.Component, error) {
	key := fmt.Sprintf("/components/%s/%s", appName, s.Name)
	value, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	_, err = store.client.Create(key, string(value))

	// Create all the other base dirs
	_, err = store.client.CreateDirectory(fmt.Sprintf("/releases/%s/%s", appName, s.Name))

	return s, err
}

func (store *ComponentStorage) List(appName string) ([]*model.Component, error) {
	key := fmt.Sprintf("/components/%s", appName)
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}

	components := make([]*model.Component, 0)

	for _, node := range resp.Node.Nodes {
		value := node.Value
		s := new(model.Component)
		if err := json.Unmarshal([]byte(value), s); err != nil {
			return nil, err
		}
		components = append(components, s)
	}
	return components, nil
}

func (store *ComponentStorage) Get(appName string, name string) (*model.Component, error) {
	// TODO repeated, move to method
	key := fmt.Sprintf("/components/%s/%s", appName, name)
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}
	value := resp.Node.Value

	s := new(model.Component)
	if err := json.Unmarshal([]byte(value), s); err != nil {
		return nil, err
	}
	return s, nil
}

// No update for Component

func (store *ComponentStorage) Delete(appName string, name string) error {
	_, err := store.client.Delete(fmt.Sprintf("/components/%s/%s", appName, name))
	_, err = store.client.Delete(fmt.Sprintf("/releases/%s/%s", appName, name))
	return err
}
