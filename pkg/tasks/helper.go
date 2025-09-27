package tasks

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	DefaultQueue  string = "default"
	PriorityQueue string = "priority"
)

type ScheduledTask struct {
	Task       *asynq.Task
	CronSpec   string
	Queue      string
	MaxRetries int
}

type Payload[T any] struct {
	CorrelationID string `json:"correlation_id"`
	Data          T      `json:"data"`
}

func New[T any](ctx context.Context, name string, data T, opts ...asynq.Option) (*asynq.Task, error) {
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
