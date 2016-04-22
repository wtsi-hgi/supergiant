package core

import (
	"encoding/json"

	"github.com/supergiant/supergiant/common"
)

type TasksInterface interface {
	List() (*TaskList, error)
	New() *TaskResource
	Start(*Action) (*TaskResource, error)
	Create(*TaskResource) error
	Get(common.ID) (*TaskResource, error)
	Update(common.ID, *TaskResource) error
	Delete(*TaskResource) error
}

type TaskCollection struct {
	core *Core
}

type TaskResource struct {
	core       *Core
	collection TasksInterface
	*common.Task
}

// NOTE this does not inherit from common like model does; all we need is a List
// object, internally, that has a slice of our composed model above.
type TaskList struct {
	Items []*TaskResource `json:"items"`
}

const (
	statusQueued  = "QUEUED"
	statusRunning = "RUNNING"
	statusFailed  = "FAILED"
)

// initializeResource implements the Collection interface.
func (c *TaskCollection) initializeResource(in Resource) {
	r := in.(*TaskResource)
	r.collection = c
	r.core = c.core
}

// List returns an TaskList.
func (c *TaskCollection) List() (*TaskList, error) {
	list := new(TaskList)
	err := c.core.db.list(c, list)
	return list, err
}

// New initializes an Task with a pointer to the Collection.
func (c *TaskCollection) New() *TaskResource {
	r := &TaskResource{
		Task: &common.Task{
			Meta: common.NewMeta(),
		},
	}
	c.initializeResource(r)
	return r
}

// Start builds a Task from an Action and creates it in etcd. The QUEUED status
// will inform the Supervisor to Claim the Task.
func (c *TaskCollection) Start(action *Action) (*TaskResource, error) {
	data, err := json.Marshal(action)
	if err != nil {
		return nil, err
	}

	task := c.New()
	task.ID = action.ID()
	task.ActionData = string(data)
	task.MaxAttempts = 10

	// NOTE we could set status to RUNNING here, and simply fire off directly
	// to Supervisor in order to prevent re-loading the Resource from the DB
	// before performing the action. However, that could introduce complications
	// if the task fails to start -- at that point it would still be flagged
	// RUNNING, and never picked up by the Supervisor.
	task.Status = statusQueued

	if err := c.Create(task); err != nil {
		return nil, err
	}
	return task, nil
}

// Create takes an Task and creates it in etcd.
func (c *TaskCollection) Create(r *TaskResource) error {
	return c.core.db.create(c, r.ID, r)
}

// Get takes a name and returns an TaskResource if it exists.
func (c *TaskCollection) Get(id common.ID) (*TaskResource, error) {
	r := c.New()
	if err := c.core.db.get(c, id, r); err != nil {
		return nil, err
	}
	return r, nil
}

// Update updates the Task in etcd.
func (c *TaskCollection) Update(id common.ID, r *TaskResource) error {
	return c.core.db.patch(c, id, r)
}

// Delete deletes the Task in etcd.
func (c *TaskCollection) Delete(r *TaskResource) error {
	return c.core.db.delete(c, r.ID)
}

//------------------------------------------------------------------------------

// Key implements the Locatable interface.
func (c *TaskCollection) locationKey() string {
	return "tasks"
}

// Parent implements the Locatable interface. It returns nil here because Core
// is the parent, and it is the root, which we exclude from paths.
func (c *TaskCollection) parent() (l Locatable) {
	return
}

// Child implements the Locatable interface.
func (c *TaskCollection) child(key string) Locatable {
	task, err := c.Get(common.IDString(key))
	if err != nil {
		Log.Panicf("No child with key %s for %T", key, c)
	}
	return task
}

// Key implements the Locatable interface.
func (r *TaskResource) locationKey() string {
	return common.StringID(r.ID)
}

// Parent implements the Locatable interface.
func (r *TaskResource) parent() Locatable {
	return r.collection.(Locatable)
}

// Child implements the Locatable interface.
func (r *TaskResource) child(key string) (l Locatable) {
	switch key {
	default:
		Log.Panicf("No child with key %s for %T", key, r)
	}
	return
}

// Action implements the Resource interface.
// NOTE Tasks are inextricably linked to Actions, but can have Actions of their
// own, since they are Resources themselves. The ToAction() method is very, very
// different from this method, and is used to convert a Task into the Action.
func (r *TaskResource) Action(name string) *Action {
	var fn ActionPerformer
	switch name {
	default:
		Log.Panicf("No action %s for Task", name)
	}
	return &Action{
		ActionName: name,
		core:       r.core,
		resource:   r,
		performer:  fn,
	}
}

//------------------------------------------------------------------------------

// decorate implements the Resource interface
func (r *TaskResource) decorate() (err error) {
	return
}

// See NOTE on Action() method
func (r *TaskResource) ToAction() *Action {
	a := &Action{
		core: r.core,
	}
	if err := json.Unmarshal([]byte(r.ActionData), a); err != nil {
		panic(err)
	}
	return a
}

// Delete deletes the Task in etcd.
func (r *TaskResource) Delete() error {
	return r.collection.Delete(r)
}

// Update saves the Task in etcd through an update.
func (r *TaskResource) Update() error {
	return r.collection.Update(r.ID, r)
}

func (r *TaskResource) IsQueued() bool {
	return r.Status == statusQueued
}

// Claim updates the Task status to "RUNNING" and returns nil. compareAndSwap is
// used to prevent a race condition and ensure only one worker performs the task.
func (r *TaskResource) Claim() error {
	// NOTE we de-ref the task because the DB will strip the ID (maybe a TODO)
	prev := *r

	// NOTE we have to do this instead of the above, because nested pointers are
	// not de-referenced.
	t := *r.Task
	t.Status = statusRunning
	next := &TaskResource{Task: &t}

	return r.core.db.compareAndSwap(r.collection.(Collection), r.ID, &prev, next)
}

func (r *TaskResource) RecordError(err error) error {
	Log.Error(err)

	r.Error = err.Error()
	if r.Attempts < r.MaxAttempts {
		r.Status = statusQueued // Add back to queue for retry
	} else {
		// TODO ideally we should save these to see failure.
		// However, with the resource ID scheme, it means actions cannot be repeated
		// r.Status = statusFailed
		Log.Error("Deleting failed Task")
		return r.Delete()
	}
	r.Attempts++

	return r.Update()
}
