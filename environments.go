package core

import "encoding/json"

type Environments struct {
	Client *Client
	Items  []*Environment
}

func (e *Environments) resourceName() string {
	return "environments"
}

// Implements the EntityList interface
func (e *Environments) NewEntity() Entity {
	entity := &Environment{Client: e.Client}
	e.Items = append(e.Items, entity)
	return entity
}

func (e *Environments) List() *Environments {
	e.Client.List(e.resourceName()).Into(e)
	return e
}

func (e *Environments) Get(name string) *Environment {
	env := &Environment{Client: e.Client}
	e.Client.Get(e.resourceName(), name).Into(env)
	return env
}

func (e *Environments) Create(name string, env *Environment) *Environment {
	value, err := json.Marshal(env)
	if err != nil {
		panic(err)
	}
	e.Client.Create(e.resourceName(), name, string(value)).Into(env)
	return env
}
