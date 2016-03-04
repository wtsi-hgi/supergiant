package storage

import (
	"encoding/json"
	"fmt"
	"supergiant/core/model"
)

type Environment struct {
	Client *Client
}

// TODO
func (store *Environment) CreateBaseDirectory() {
	if _, err := store.Client.Get("/environments"); err != nil {
		if _, err := store.Client.CreateDirectory("/environments"); err != nil {
			panic(err)
		}
	}
}

func (store *Environment) Create(e *model.Environment) (*model.Environment, error) {
	key := fmt.Sprintf("/environments/%s", e.Name)
	value, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	if _, err = store.Client.Create(key, string(value)); err != nil {
		return nil, err
	}

	// Create all the other base dirs
	_, err = store.Client.CreateDirectory(fmt.Sprintf("services/%s", e.Name))
	return e, err
}

func (store *Environment) List() ([]*model.Environment, error) {
	key := "/environments"
	resp, err := store.Client.Get(key)
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

func (store *Environment) Get(id string) (*model.Environment, error) {
	// TODO repeated, move to method
	key := fmt.Sprintf("/environments/%s", id)
	resp, err := store.Client.Get(key)
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

func (store *Environment) Delete(id string) error {
	// TODO repeated
	key := fmt.Sprintf("/environments/%s", id)
	_, err := store.Client.Delete(key)
	if err != nil {
		return err
	}
	return nil
}
