package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

type asynqTaskEnqueuer struct {
	client *asynq.Client
}

type AsynqRedisConfig struct {
	Host     string
	Port     int
	Password string
	DbIndex  int
}

func NewAsynqTaskEnqueuer(redisConfig AsynqRedisConfig) TaskEnqueuer {
	asyncqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password: redisConfig.Password,
		DB:       redisConfig.DbIndex,
	})

	return &asynqTaskEnqueuer{
		client: asyncqClient,
	}
}

func (t *asynqTaskEnqueuer) Enqueue(ctx context.Context, task Task, maxRetryCount *int, queueName *string) error {
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
	taskOptions := defaultTaskOptions
	if maxRetryCount != nil {
		taskOptions = append(taskOptions, asynq.MaxRetry(*maxRetryCount))
	}
	if queueName != nil {
		taskOptions = append(taskOptions, asynq.Queue(*queueName))
	}

	taskToBeEnqueued, err := NewAsynqTask(ctx, task.Name(), task.Payload(), taskOptions...)
	if err != nil {
		return err
	}

	taskInfo, err := t.client.EnqueueContext(ctx, taskToBeEnqueued)
	if err != nil {
		return err
	}
	logger.Info(ctx, "Enqueued task: %s in queue: %s", task.Name(), taskInfo.Queue)
	return nil
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
