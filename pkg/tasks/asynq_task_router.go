package tasks

import (
	"context"

	"github.com/hibiken/asynq"
)

type asynqTaskRouter struct {
	mux *asynq.ServeMux
}

func NewAsynqTaskRouter() TaskRouter {
	return &asynqTaskRouter{
		mux: asynq.NewServeMux(),
	}
}

func (r *asynqTaskRouter) RegisterTaskProcessor(taskName string, taskProcessor TaskProcessor) {
	r.mux.HandleFunc(taskName, func(ctx context.Context, asynqTask *asynq.Task) error {
		// convert asynqTask to Task interface
		task := &asynqTaskToTaskAdapter{
			typename: asynqTask.Type(),
			payload:  asynqTask.Payload(),
		}
		return taskProcessor.ProcessTask(ctx, task)
	})
}

func (r *asynqTaskRouter) Handler() any {
	return r.mux
}

type asynqTaskToTaskAdapter struct {
	typename string
	payload  []byte
	opts     []asynq.Option
}

func (t *asynqTaskToTaskAdapter) Name() string {
	return t.typename
}

func (t *asynqTaskToTaskAdapter) Queue() string {
	for _, opt := range t.opts {
		if opt.Type() == asynq.QueueOpt {
			return opt.String()
		}
	}
	return DefaultQueue
}

func (t *asynqTaskToTaskAdapter) MaxRetryCount() int {
	for _, opt := range t.opts {
		if opt.Type() == asynq.MaxRetryOpt {
			return opt.Value().(int)
		}
	}
	return 0
}

func (t *asynqTaskToTaskAdapter) Payload() any {
	return t.payload
}
