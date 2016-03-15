package storage

import (
	"encoding/json"
	"fmt"
	"supergiant/core/model"
)

type ImageRepoStorage struct {
	client *Client
}

func (store *ImageRepoStorage) Create(e *model.ImageRepo) (*model.ImageRepo, error) {
	key := fmt.Sprintf("/repos/dockerhub/%s", e.Name)
	value, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	if _, err = store.client.Create(key, string(value)); err != nil {
		return nil, err
	}
	return e, nil
}

func (store *ImageRepoStorage) List() ([]*model.ImageRepo, error) {
	key := "/repos/dockerhub"
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}

	repos := make([]*model.ImageRepo, 0)

	for _, node := range resp.Node.Nodes {
		value := node.Value
		e := new(model.ImageRepo)
		if err := json.Unmarshal([]byte(value), e); err != nil {
			return nil, err
		}
		repos = append(repos, e)
	}
	return repos, nil
}

func (store *ImageRepoStorage) Get(id string) (*model.ImageRepo, error) {
	// TODO repeated, move to method
	key := fmt.Sprintf("/repos/dockerhub/%s", id)
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}
	value := resp.Node.Value

	e := new(model.ImageRepo)
	if err := json.Unmarshal([]byte(value), e); err != nil {
		return nil, err
	}
	return e, nil
}

// No update for ImageRepo

func (store *ImageRepoStorage) Delete(id string) error {
	// TODO repeated
	_, err := store.client.Delete(fmt.Sprintf("/repos/dockerhub/%s", id))
	return err
}
