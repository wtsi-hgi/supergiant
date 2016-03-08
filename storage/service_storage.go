package storage

import (
	"encoding/json"
	"fmt"
	"supergiant/core/model"
)

type ServiceStorage struct {
	client *Client
}

// TODO
func (store *ServiceStorage) CreateBaseDirectory() {
	if _, err := store.client.Get("/services"); err != nil {
		if _, err := store.client.CreateDirectory("/services"); err != nil {
			panic(err)
		}
	}
}

func (store *ServiceStorage) Create(environmentID string, s *model.Service) (*model.Service, error) {
	key := fmt.Sprintf("/services/%s/%s", environmentID, s.Name)
	value, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	_, err = store.client.Create(key, string(value))
	return s, err
}

func (store *ServiceStorage) List(environmentID string) ([]*model.Service, error) {
	key := fmt.Sprintf("/services/%s", environmentID)
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}

	var services []*model.Service

	for _, node := range resp.Node.Nodes {
		id := node.Key
		value := node.Value
		s := &model.Service{Name: id}
		if err := json.Unmarshal([]byte(value), s); err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, nil
}

func (store *ServiceStorage) Get(environmentID string, id string) (*model.Service, error) {
	// TODO repeated, move to method
	key := fmt.Sprintf("/services/%s/%s", environmentID, id)
	resp, err := store.client.Get(key)
	if err != nil {
		return nil, err
	}
	value := resp.Node.Value

	s := &model.Service{Name: id}
	if err := json.Unmarshal([]byte(value), s); err != nil {
		return nil, err
	}
	return s, nil
}

// No update for Service

func (store *ServiceStorage) Delete(environmentID string, id string) error {
	// TODO repeated
	key := fmt.Sprintf("/services/%s/%s", environmentID, id)
	_, err := store.client.Delete(key)
	if err != nil {
		return err
	}
	return nil
}
