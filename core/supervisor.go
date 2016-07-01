package core

import (
	"time"

	"github.com/satori/go.uuid"
	"github.com/supergiant/supergiant/common"
)

const (
	interval = time.Second
)

type Supervisor struct {
	core       *Core
	numWorkers int
	tasks      chan *TaskResource
	workers    map[string]*worker
}

func NewSupervisor(c *Core, numWorkers int) *Supervisor {
	return &Supervisor{core: c, numWorkers: numWorkers}
}

func (s *Supervisor) Run() {
	// This starts all workers listening on the channel
	s.workers = make(map[string]*worker)
	s.tasks = make(chan *TaskResource)
	for i := 0; i < s.numWorkers; i++ {
		worker := newWorker(s)
		s.workers[worker.id] = worker
		go worker.start()
	}

	// Clear any hanging tasks
	list, err := s.core.Tasks().List()
	if err != nil {
		panic(err)
	}
	for _, task := range list.Items {
		if task.IsRunning() {
			Log.Warnf("Deleting hanging task with ID %s", common.StringID(task.ID))
			if err := task.Delete(); err != nil {
				panic(err)
			}
		}
	}

	// loop every <interval> seconds
	for _ = range time.NewTicker(interval).C {

		list, err := s.core.Tasks().List()
		if err != nil {
			panic(err)
		}

		// Find first queued task, or return.
		// Claim task and return if claim fails.
		var task *TaskResource

		for _, j := range list.Items {
			if j.IsQueued() {
				task = j
				break
			}
		}
		if task == nil {
			continue
		}

		if err := task.Claim(); err != nil {
			// NOTE should be a CompareAndSwap error if anything
			Log.Error(err)
			continue
		}

		// NOTE since this is not using a buffered channel, this will block if all
		// workers are busy, which I think should be the expected behavior.
		Log.Debugf("Submitting Task with ID %s", common.StringID(task.ID))
		s.tasks <- task
	}
}

func (s *Supervisor) cancel(task *TaskResource) {
	if !task.IsRunning() {
		Log.Warnf("Can't cancel non-running Task with ID %s", common.StringID(task.ID))
		return
	}

	worker := s.workers[task.WorkerID]
	if worker == nil {
		Log.Errorf("Could not find worker for task with ID %s", common.StringID(task.ID))
	}

	Log.Debugf("Supervisor is attempting cancellation of Task with ID %s", common.StringID(task.ID))
	worker.cancel <- struct{}{}
	Log.Debugf("Supervisor submitted cancellation of Task with ID %s", common.StringID(task.ID))
}

type worker struct {
	s      *Supervisor
	id     string
	cancel chan struct{}
}

func newWorker(s *Supervisor) *worker {
	return &worker{s: s, id: uuid.NewV4().String(), cancel: make(chan struct{})}
}

func (w *worker) start() {
	for task := range w.s.tasks {
		// recover from panic, capture error and report
		defer func() {
			if ret := recover(); ret != nil {
				if err, isErr := ret.(error); isErr && task != nil {
					recordError(task, err)
				} else {
					panic(ret)
				}
			}
		}()

		task.WorkerID = w.id
		if err := task.Update(); err != nil {
			panic(err)
		}

		action := task.ToAction().initialize(w.s.core)

		done := make(chan error)
		go func() {
			Log.Infof("Starting Task %s : %s", action.ActionName, action.ResourceLocation)
			done <- action.Perform()
		}()

		select {
		case err := <-done:
			if err != nil {
				recordError(task, err)
			} else {
				Log.Infof("Completed Task %s : %s", action.ActionName, action.ResourceLocation)
				task.Delete() // Task is successful or cancelled, delete from Queue
			}
		case <-w.cancel:
			Log.Infof("Cancelling Task %s : %s", action.ActionName, action.ResourceLocation)
			close(done) // this should stop the action.Perform go routine
		}
	}
}

// Record error, and panic if that goes wrong
func recordError(task *TaskResource, err error) {
	if uberErr := task.RecordError(err); uberErr != nil {
		panic(uberErr)
	}
}
