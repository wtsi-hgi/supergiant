package storage

import (
	"encoding/json"
	"fmt"
	"supergiant/core/model"
)

type DeploymentStorage struct {
	client *Client
}

// // TODO
// func (store *DeploymentStorage) CreateBaseDirectory() {
// 	if _, err := store.client.Get("/deployments"); err != nil {
// 		if _, err := store.client.CreateDirectory("/deployments"); err != nil {
// 			panic(err)
// 		}
// 	}
// }

func (store *DeploymentStorage) Create(e *model.Deployment) (*model.Deployment, error) {
	key := fmt.Sprintf("/deployments/%s", e.ID)
	value, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	if _, err = store.client.Create(key, string(value)); err != nil {
		return nil, err
	}
	return e, nil
}

// func (store *DeploymentStorage) List() ([]*model.Deployment, error) {
// 	key := "/deployments"
// 	resp, err := store.client.Get(key)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var apps []*model.App
//
// 	for _, node := range resp.Node.Nodes {
// 		value := node.Value
// 		e := new(model.App)
// 		if err := json.Unmarshal([]byte(value), e); err != nil {
// 			return nil, err
// 		}
// 		apps = append(apps, e)
// 	}
// 	return apps, nil
// }

func (store *DeploymentStorage) Get(id string) (*model.Deployment, error) {
	// TODO repeated, move to method
	key := fmt.Sprintf("/deployments/%s", id)
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}
	value := resp.Node.Value

	e := new(model.Deployment)
	if err := json.Unmarshal([]byte(value), e); err != nil {
		return nil, err
	}
	return e, nil
}

// No update for App

func (store *DeploymentStorage) Delete(id string) error {
	// TODO repeated
	_, err := store.client.Delete(fmt.Sprintf("/deployments/%s", id))
	// _, err = store.client.Delete(fmt.Sprintf("/instances/%s", id))
	return err
}
