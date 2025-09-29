package tasks

import (
	"context"
	"encoding/json"

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

func (t *asynqTaskEnqueuer) Enqueue(ctx context.Context, task Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	queue := task.Queue()
	if queue == "" {
		queue = DefaultQueue
	}

	defaultTaskOptions := []asynq.Option{
		asynq.MaxRetry(task.MaxRetryCount()),
		asynq.Queue(queue),
	}

	// Merge default options with caller-provided options
	// If the same option appears in both, the caller's value takes precedence
	taskOptions := append(defaultTaskOptions, opts...)

	taskToBeEnqueued, err := NewAsynqTask(ctx, task.Name(), task.Payload(), taskOptions...)
	if err != nil {
		return nil, err
	}
	return t.client.EnqueueContext(ctx, taskToBeEnqueued)
}

func (t *asynqTaskEnqueuer) Close() error {
	return t.client.Close()
}

func NewAsynqTask[T any](ctx context.Context, name string, data T, opts ...asynq.Option) (*asynq.Task, error) {
	val := ctx.Value("correlation_id")
	correlationID, ok := val.(string)
	if !ok {
		correlationID = ""
	}

	payload := Payload[T]{CorrelationID: correlationID, Data: data}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(name, payloadBytes, opts...), nil
}
