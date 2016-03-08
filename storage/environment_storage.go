package storage

import (
	"encoding/json"
	"fmt"
	"supergiant/core/model"
)

type EnvironmentStorage struct {
	client *Client
}

// TODO
func (store *EnvironmentStorage) CreateBaseDirectory() {
	if _, err := store.client.Get("/environments"); err != nil {
		if _, err := store.client.CreateDirectory("/environments"); err != nil {
			panic(err)
		}
	}
}

func (store *EnvironmentStorage) Create(e *model.Environment) (*model.Environment, error) {
	key := fmt.Sprintf("/environments/%s", e.Name)
	value, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	if _, err = store.client.Create(key, string(value)); err != nil {
		return nil, err
	}

	// Create all the other base dirs
	_, err = store.client.CreateDirectory(fmt.Sprintf("services/%s", e.Name))
	return e, err
}

func (store *EnvironmentStorage) List() ([]*model.Environment, error) {
	key := "/environments"
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}

	var environments []*model.Environment

	for _, node := range resp.Node.Nodes {
		id := node.Key
		value := node.Value
		e := &model.Environment{Name: id}
		if err := json.Unmarshal([]byte(value), e); err != nil {
			return nil, err
		}
		environments = append(environments, e)
	}
	return environments, nil
}

func (store *EnvironmentStorage) Get(id string) (*model.Environment, error) {
	// TODO repeated, move to method
	key := fmt.Sprintf("/environments/%s", id)
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}
	value := resp.Node.Value

	e := &model.Environment{Name: id}
	if err := json.Unmarshal([]byte(value), e); err != nil {
		return nil, err
	}
	return e, nil
}

// No update for Environment

func (store *EnvironmentStorage) Delete(id string) error {
	// TODO repeated
	key := fmt.Sprintf("/environments/%s", id)
	_, err := store.client.Delete(key)
	if err != nil {
		return err
	}
	return nil
}
