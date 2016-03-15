package storage

import (
	"encoding/json"
	"fmt"
	"supergiant/core/model"
)

type InstanceStorage struct {
	client *Client
}

// // TODO
// func (store *InstanceStorage) CreateBaseDirectory() {
// 	if _, err := store.client.Get("/instances"); err != nil {
// 		if _, err := store.client.CreateDirectory("/instances"); err != nil {
// 			panic(err)
// 		}
// 	}
// }

func (store *InstanceStorage) Create(deploymentID string, s *model.Instance) (*model.Instance, error) {
	key := fmt.Sprintf("/instances/%s/%s", deploymentID, s.ID)
	value, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	_, err = store.client.Create(key, string(value))

	// Create all the other base dirs
	// _, err = store.client.CreateDirectory(fmt.Sprintf("/releases/%s/%s", deploymentID, s.ID))

	return s, err
}

func (store *InstanceStorage) List(deploymentID string) ([]*model.Instance, error) {
	key := fmt.Sprintf("/instances/%s", deploymentID)
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}

	instances := make([]*model.Instance, 0)

	for _, node := range resp.Node.Nodes {
		value := node.Value
		s := new(model.Instance)
		if err := json.Unmarshal([]byte(value), s); err != nil {
			return nil, err
		}
		instances = append(instances, s)
	}
	return instances, nil
}

func (store *InstanceStorage) Get(deploymentID string, id string) (*model.Instance, error) {
	// TODO repeated, move to method
	key := fmt.Sprintf("/instances/%s/%s", deploymentID, id)
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}
	value := resp.Node.Value

	s := new(model.Instance)
	if err := json.Unmarshal([]byte(value), s); err != nil {
		return nil, err
	}
	return s, nil
}

// No update for Component

func (store *InstanceStorage) Delete(deploymentID string, id string) error {
	_, err := store.client.Delete(fmt.Sprintf("/instances/%s/%s", deploymentID, id))
	return err
}
