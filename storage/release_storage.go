package storage

import (
	"encoding/json"
	"fmt"
	"supergiant/core/model"
)

type ReleaseStorage struct {
	client *Client
}

// // TODO
// func (store *ReleaseStorage) CreateBaseDirectory() {
// 	if _, err := store.client.Get("/releases"); err != nil {
// 		if _, err := store.client.CreateDirectory("/releases"); err != nil {
// 			panic(err)
// 		}
// 	}
// }

func (store *ReleaseStorage) Create(appName string, compName string, s *model.Release) (*model.Release, error) {

	// NOTE that controller will need to autogenerate ID

	key := fmt.Sprintf("/releases/%s/%s/%d", appName, compName, s.ID)
	value, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	_, err = store.client.Create(key, string(value))

	return s, err
}

func (store *ReleaseStorage) List(appName string, compName string) ([]*model.Release, error) {
	key := fmt.Sprintf("/releases/%s/%s", appName, compName)
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}

	releases := make([]*model.Release, 0)

	for _, node := range resp.Node.Nodes {
		value := node.Value
		s := new(model.Release)
		if err := json.Unmarshal([]byte(value), s); err != nil {
			return nil, err
		}
		releases = append(releases, s)
	}
	return releases, nil
}

func (store *ReleaseStorage) Get(appName string, compName string, id string) (*model.Release, error) {
	// TODO repeated, move to method
	key := fmt.Sprintf("/releases/%s/%s/%s", appName, compName, id)
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}
	value := resp.Node.Value

	s := new(model.Release)
	if err := json.Unmarshal([]byte(value), s); err != nil {
		return nil, err
	}
	return s, nil
}

// No update for Component

func (store *ReleaseStorage) Delete(appName string, name string, id string) error {
	_, err := store.client.Delete(fmt.Sprintf("/releases/%s/%s/%s", appName, name, id))
	return err
}
