package tasks

import (
	"context"

	"github.com/hibiken/asynq"
)

type Task interface {
	Name() string
	Queue() string
	MaxRetryCount() int
	Payload() any
}

type TaskEnqueuer interface {
	Enqueue(ctx context.Context, task Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
	Close() error
}
