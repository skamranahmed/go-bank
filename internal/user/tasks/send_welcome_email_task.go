package tasks

import (
	"context"
	"fmt"

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
		queue:         tasksHelper.DefaultQueue,
		maxRetryCount: 0,
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

func NewSendWelcomeEmailTaskProcessor(services *internal.Services) tasksHelper.TaskProcessor {
	return &SendWelcomeEmailTaskProcessor{
		services: services,
	}
}

func (processor *SendWelcomeEmailTaskProcessor) ProcessTask(ctx context.Context, t tasksHelper.Task) error {
	taskPayloadInBytes := t.Payload().([]byte)
	payload, err := tasksHelper.ExtractPayload[SendWelcomeEmailTaskPayload](taskPayloadInBytes)
	if err != nil {
		return fmt.Errorf("Unable to extract payload for task: %s, error: %v", t.Name(), err)
	}

	ctx = context.WithValue(ctx, "correlation_id", payload.CorrelationID)

	// TODO: maybe add a real email provider here in the future
	logger.Info(ctx, "[Dummy] send welcome email to userID: %+v", payload.Data.UserID)
	return nil
}
