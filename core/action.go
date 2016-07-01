package core

import (
	"encoding/base64"
	"reflect"

	"github.com/supergiant/supergiant/common"
)

type Action struct {
	ResourceLocation string `json:"resource_location"`
	ActionName       string `json:"action_name"`
	core             *Core
	resource         Resource // corresponds to ResourceLocation
}

func NewAction(core *Core, r Resource, name string) *Action {
	return &Action{
		core:       core,
		resource:   r,
		ActionName: name,
	}
}

// initialize takes a *Core, and loads resource and performer on the Action.
// It is used by Supervisor when retrying a Task loaded from the db.
// It returns the *Action purely for the sake of doing initialize(c).Perform()
func (a *Action) initialize(c *Core) *Action {
	a.core = c
	a.resource = LocateResource(c, a.ResourceLocation)
	return a
}

// ID returns a common.ID which is an SHA1 checksum of action_name:resource_key.
// This creates a simple "mutex" on Resource actions since a create operation on
// an existing key will fail.
func (a *Action) ID() common.ID {
	data := a.ActionName + ":" + a.ResourceLocation
	id := base64.StdEncoding.EncodeToString([]byte(data))
	return common.IDString(id)
}

// Perform performs the Action by calling the performer func.
func (a *Action) Perform() error {
	rv := reflect.ValueOf(a.resource)
	colv := rv.Elem().FieldByName("Collection")
	fnv := colv.MethodByName(a.ActionName)

	retv := fnv.Call([]reflect.Value{
		reflect.ValueOf(rv.Interface().(Resource)),
	})

	ret := retv[0].Interface()
	if ret != nil {
		return ret.(error)
	}
	return nil
}

// Supervise sets the ResourceLocation of the resource and creates a Task from
// the Action.
func (a *Action) Supervise() error {
	a.ResourceLocation = ResourceLocation(a.resource.(Locatable))
	_, err := a.core.Tasks().Start(a)
	return err
}

func (a *Action) CancelTasks() *Action {
	if err := a.core.Tasks().DeleteByResource(a.resource.(Locatable)); err != nil {
		Log.Error(err)
	}
	return a
}
