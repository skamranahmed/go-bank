package tasks

import (
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
