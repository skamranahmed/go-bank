package tasks

import (
	"github.com/hibiken/asynq"
)

type asynqTaskEnqueuer struct {
	client *asynq.Client
}

func NewAsynqTaskEnqueuer(opt asynq.RedisConnOpt) TaskEnqueuer {
	asyncqClient := asynq.NewClient(opt)

	return &asynqTaskEnqueuer{
		client: asyncqClient,
	}
}

func (t *asynqTaskEnqueuer) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return t.client.Enqueue(task, opts...)
}

func (t *asynqTaskEnqueuer) Close() error {
	return t.client.Close()
}
