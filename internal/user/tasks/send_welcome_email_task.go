package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/skamranahmed/go-bank/internal"
	"github.com/skamranahmed/go-bank/pkg/logger"
	tasksHelper "github.com/skamranahmed/go-bank/pkg/tasks"
)

const SendWelcomeEmailTaskName string = "task:send_welcome_email"

type SendWelcomeEmailTaskPayload struct {
	UserID string
}

type SendWelcomeEmailTask struct {
	name          string
	queue         string
	maxRetryCount int
	payload       SendWelcomeEmailTaskPayload
}

func NewSendWelcomeEmailTask(userID string) tasksHelper.Task {
	return &SendWelcomeEmailTask{
		name:          SendWelcomeEmailTaskName,
		maxRetryCount: 1,
		queue:         tasksHelper.DefaultQueue,
		payload: SendWelcomeEmailTaskPayload{
			UserID: userID,
		},
	}
}

func (t *SendWelcomeEmailTask) Name() string {
	return t.name
}

func (t *SendWelcomeEmailTask) Queue() string {
	return t.queue
}

func (t *SendWelcomeEmailTask) MaxRetryCount() int {
	return t.maxRetryCount
}

func (t *SendWelcomeEmailTask) Payload() any {
	return t.payload
}

type SendWelcomeEmailTaskProcessor struct {
	services *internal.Services
}

func NewSendWelcomeEmailTaskProcessor(services *internal.Services) *SendWelcomeEmailTaskProcessor {
	return &SendWelcomeEmailTaskProcessor{
		services: services,
	}
}

func (processor *SendWelcomeEmailTaskProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload tasksHelper.Payload[SendWelcomeEmailTaskPayload]
	err := json.Unmarshal(t.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	ctx = context.WithValue(ctx, "correlation_id", payload.CorrelationID)

	// TODO: maybe add a real email provider here in the future
	logger.Info(ctx, "[Dummy] send welcome email to userID: %+v", payload.Data.UserID)
	return nil
}
