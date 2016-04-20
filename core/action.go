package core

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"github.com/supergiant/supergiant/common"
)

type ActionPerformer func(r Resource) error

type Action struct {
	ResourceLocation string `json:"resource_location"`
	ActionName       string `json:"action_name"`
	core             *Core
	resource         Resource        // corresponds to ResourceLocation
	performer        ActionPerformer // corresponds to ActionName
}

// initialize takes a *Core, and loads resource and performer on the Action.
// It is used by Supervisor when retrying a Task loaded from the DB.
// It returns the *Action purely for the sake of doing initialize(c).Perform()
func (a *Action) initialize(c *Core) *Action {
	a.core = c
	a.resource = LocateResource(c, a.ResourceLocation)
	a.performer = a.resource.Action(a.ActionName).performer
	return a
}

// ID returns a common.ID which is an SHA1 checksum of action_name:resource_key.
// This creates a simple "mutex" on Resource actions since a create operation on
// an existing key will fail.
func (a *Action) ID() common.ID {
	data := fmt.Sprintf("%s:%s", a.ActionName, a.ResourceLocation)
	hash := sha1.New()
	hash.Write([]byte(data))
	id := hex.EncodeToString(hash.Sum(nil))
	return common.IDString(id)
}

// Perform performs the Action by calling the performer func.
func (a *Action) Perform() error {
	return a.performer(a.resource)
}

// Supervise sets the ResourceLocation of the resource and creates a Task from
// the Action.
func (a *Action) Supervise() error {
	a.ResourceLocation = ResourceLocation(a.resource.(Locatable))
	_, err := a.core.Tasks().Start(a)
	return err
}
