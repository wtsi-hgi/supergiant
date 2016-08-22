package core

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/supergiant/supergiant/pkg/models"
	"github.com/supergiant/supergiant/pkg/util"
)

type Action struct {
	Status *models.ActionStatus
	core   *Core

	scope *DB
	model models.Model
	id    *int64

	resourceID string

	fn             func(*Action) error
	cancelExisting bool
}

type RepeatedActionError struct {
	ResourceID string
}

func (err *RepeatedActionError) Error() string {
	return "Already perform action for " + err.ResourceID
}

//------------------------------------------------------------------------------

func (a *Action) description() string {
	modelType := strings.Split(reflect.TypeOf(a.model).String(), ".")[1]
	return fmt.Sprintf("%s %s %s", a.Status.Description, modelType, a.resourceID)
}

func (a *Action) prepare() error {
	if a.resourceID != "" {
		return nil
	}
	if err := a.scope.First(a.model, *a.id); err != nil {
		return err
	}
	a.resourceID = a.model.GetUUID()
	return nil
}

func (a *Action) request(reqType actionRequestType) *Action {
	ch := make(chan *Action)
	a.core.actionRequestChannel <- &actionRequest{ch, reqType, a}
	return <-ch
}

func (a *Action) stopUnlessCancelled() {
	if !a.Status.Cancelled {
		a.request(requestActionStop)
	}
}

func (a *Action) cancellableWaitFor(desc string, d time.Duration, i time.Duration, fn func() (bool, error)) error {
	return util.WaitFor(desc, d, i, func() (bool, error) {
		if a.Status.Cancelled {
			return false, fmt.Errorf("Action cancelled while waiting for %s", desc)
		}
		return fn()
	})
}

func (a *Action) Now() error {
	if err := a.prepare(); err != nil {
		return err
	}

	if existing := a.request(requestActionFetch); existing != nil {
		if a.cancelExisting {
			existing.Status.Cancelled = true
			existing.request(requestActionStop)
		} else {
			return &RepeatedActionError{a.resourceID}
		}
	}

	// TODO we may want some means of communicating with the existing action, to
	// know that it has stopped its goroutines before continuing.

	a.request(requestActionStart)

	// Remove Action from map regardless of success or failure
	defer a.stopUnlessCancelled()

	return a.fn(a)
}

func (a *Action) Async() error {
	if err := a.prepare(); err != nil {
		return err
	}

	if existing := a.request(requestActionFetch); existing != nil {
		if a.cancelExisting {
			existing.Status.Cancelled = true
			existing.request(requestActionStop)
		} else if existing.Status.Retries < existing.Status.MaxRetries {
			return &RepeatedActionError{a.resourceID}
		}
	}

	a.request(requestActionStart)

	go func() {
		retries := 0
		for {
			if retries >= a.Status.MaxRetries {
				return // Don't remove from Actions
			}

			if a.Status.Cancelled {
				break // Remove from Actions
			}

			err := a.fn(a)
			if err == nil {
				break // Remove from Actions
			}

			retries++
			a.Status.Retries = retries
			a.Status.Error = err.Error()

			a.core.Log.Error(err)
		}

		a.stopUnlessCancelled()
	}()

	return nil
}

////////////////////////////////////////////////////////////////////////////////
//\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\
////////////////////////////////////////////////////////////////////////////////
//\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\

func (c *Core) SetResourceActionStatus(m models.Model) {
	if action, _ := c.Actions[m.GetUUID()]; action != nil {
		m.SetActionStatus(action.Status)
	}
}
