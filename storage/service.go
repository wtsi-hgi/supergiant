package storage

import (
	"encoding/json"
	"fmt"
	"supergiant/core/model"
)

type Service struct {
	Client *Client
}

// TODO
func (store *Service) CreateBaseDirectory() {
	if _, err := store.Client.Get("/services"); err != nil {
		if _, err := store.Client.CreateDirectory("/services"); err != nil {
			panic(err)
		}
	}
}

func (store *Service) Create(environmentID string, s *model.Service) (*model.Service, error) {
	key := fmt.Sprintf("/services/%s/%s", environmentID, s.Name)
	value, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	_, err = store.Client.Create(key, string(value))
	return s, err
}

func (store *Service) List(environmentID string) ([]*model.Service, error) {
	key := fmt.Sprintf("/services/%s", environmentID)
	resp, err := store.Client.Get(key)
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

func (store *Service) Get(environmentID string, id string) (*model.Service, error) {
	// TODO repeated, move to method
	key := fmt.Sprintf("/services/%s/%s", environmentID, id)
	resp, err := store.Client.Get(key)
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

func (store *Service) Delete(environmentID string, id string) error {
	// TODO repeated
	key := fmt.Sprintf("/services/%s/%s", environmentID, id)
	_, err := store.Client.Delete(key)
	if err != nil {
		return err
	}
	return nil
}
