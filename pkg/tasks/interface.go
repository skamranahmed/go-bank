package tasks

import (
	"context"

	"github.com/hibiken/asynq"
)

// Task represents a unit of work that can be enqueued and executed
type Task interface {
	// Name returns the unique name of the task
	Name() string

	// Queue returns the name of the queue the task should be enqueued in
	Queue() string

	// MaxRetryCount returns the maximum number of times the task can be retried
	MaxRetryCount() int

	// Payload returns the data associated with the task
	Payload() any
}

// SchedulableTask represents a task that can be scheduled to run periodically
type SchedulableTask interface {
	Task

	// CronSpec returns a cron expression specifying the schedule for the task
	CronSpec() string
}

// TaskEnqueuer defines the behavior of a component that can enqueue tasks for execution
type TaskEnqueuer interface {
	// Enqueue adds a task to the queue.
	// TODO: Consider defining our own task options abstraction instead of depending on asynq.Option
	// The concrete asynq enqueuer can then convert them to asynq options internally
	Enqueue(ctx context.Context, task Task, opts ...asynq.Option) (*asynq.TaskInfo, error)

	// Close releases any resources held by the enqueuer.
	// After Close is called, the enqueuer should not be used
	Close() error
}

// TaskScheduler defines the behavior of a component that can schedule tasks
// to be enqueued at specific times or intervals
type TaskScheduler interface {
	/*
		Start begins running the scheduler loop

		The call should block until either:
			- the scheduler becomes active, or
			- an os signal to exit the program is received on schedulerStopSignalChannel

		Implementations are responsible for ensuring that only one active
		scheduler instance runs at a time. For example, the asynq task
		scheduler uses a distributed lock (via Redlock) to enforce this.
	*/
	Start(ctx context.Context, schedulerStopSignalChannel chan struct{})

	/*
		RegisterTask registers a task with the scheduler
		so it can be enqueued according to its schedule
	*/
	RegisterTask(ctx context.Context, schedulableTask SchedulableTask) (taskEntryID string, err error)
}
