package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/supergiant/supergiant/common"
)

type TasksInterface interface {
	DeleteByResource(l Locatable) error

	List() (*TaskList, error)
	New() *TaskResource
	Start(*Action) (*TaskResource, error)
	Create(*TaskResource) error
	Get(common.ID) (*TaskResource, error)
	Update(common.ID, *TaskResource) error
	Patch(common.ID, *TaskResource) error
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

func (c *TaskCollection) DeleteByResource(l Locatable) error {
	list, err := c.List()
	if err != nil {
		return err
	}
	key := ResourceLocation(l)
	for _, task := range list.Items {

		// NOTE due to the way resource locations work, we can cancel all tasks
		// that have a matching starting ID. That way, app delete causes cascading
		// cancellation of all sub-tasks.
		if strings.HasPrefix(task.ResourceLocation(), key) {

			Log.Debugf("Requesting cancellation of Task with ID %s", common.StringID(task.ID))

			c.core.supervisor.cancel(task)
			if err := task.Delete(); err != nil {
				return err
			}
		}
	}
	return nil
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
	// to Supervisor in order to prevent re-loading the Resource from the db
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
	return c.core.db.update(c, id, r)
}

// Patch partially updates the App in etcd.
func (c *TaskCollection) Patch(name common.ID, r *TaskResource) error {
	return c.core.db.patch(c, name, r)
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
		panic(fmt.Errorf("No child with key %s for %T", key, c))
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
		panic(fmt.Errorf("No child with key %s for %T", key, r))
	}
}

//------------------------------------------------------------------------------

// decorate implements the Resource interface
func (r *TaskResource) decorate() (err error) {
	return
}

func (r *TaskResource) ResourceLocation() string {
	data, err := base64.StdEncoding.DecodeString(common.StringID(r.ID))
	if err != nil {
		panic(err)
	}
	return strings.Split(string(data), ":")[1]
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

// Patch is a proxy method to collection Patch.
func (r *TaskResource) Patch() error {
	return r.collection.Patch(r.ID, r)
}

func (r *TaskResource) IsQueued() bool {
	return r.Status == statusQueued
}

func (r *TaskResource) IsRunning() bool {
	return r.Status == statusRunning
}

// Claim updates the Task status to "RUNNING" and returns nil. compareAndSwap is
// used to prevent a race condition and ensure only one worker performs the task.
func (r *TaskResource) Claim() error {
	// NOTE we de-ref the task because the db will strip the ID (maybe a TODO)
	prev := *r

	// NOTE we have to do this instead of the above, because nested pointers are
	// not de-referenced.
	t := *r.Task
	t.Status = statusRunning
	next := r.collection.New()
	next.Task = &t

	if err := r.core.db.compareAndSwap(r.collection.(Collection), r.ID, &prev, next); err != nil {
		return err
	}

	// This will update to RUNNING status on r
	*r = *next
	return nil
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

	r.WorkerID = ""

	return r.Update()
}
