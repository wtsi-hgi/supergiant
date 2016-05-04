package core

import "time"

const (
	interval = time.Second
)

type Supervisor struct {
	core    *Core
	workers int
}

func NewSupervisor(c *Core, workers int) *Supervisor {
	return &Supervisor{c, workers}
}

func (s *Supervisor) Run() {
	// This starts all workers listening on the channel
	tasks := make(chan *TaskResource)
	for i := 0; i < s.workers; i++ {
		go s.startWorker(tasks)
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
		tasks <- task
	}
}

func (s *Supervisor) startWorker(tasks <-chan *TaskResource) {
	for task := range tasks {
		// recover from panic, capture error and report
		defer func() {
			if err := recover(); err != nil {
				recordError(task, err.(error))
			}
		}()

		action := task.ToAction().initialize(s.core)

		Log.Infof("Starting Task %s : %s", action.ActionName, action.ResourceLocation)
		if err := action.Perform(); err != nil {
			recordError(task, err)
			continue
		}

		Log.Infof("Completed Task %s : %s", action.ActionName, action.ResourceLocation)
		task.Delete() // Task is successful, delete from Queue
	}
}

// Record error, and panic if that goes wrong
func recordError(task *TaskResource, err error) {
	if uberErr := task.RecordError(err); uberErr != nil {
		panic(uberErr)
	}
}
