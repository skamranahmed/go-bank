package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	account "github.com/skamranahmed/go-bank/internal/account/service"
	user "github.com/skamranahmed/go-bank/internal/user/service"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

const SendWelcomeEmailTaskName string = "send-welcome-email"

type SendWelcomeEmailTaskPayload struct {
	CorrelationID string
	UserID        string
}

func NewSendWelcomeEmailTask(ctx context.Context, userID string) (*asynq.Task, error) {
	val := ctx.Value("correlation_id")
	correlationID, ok := val.(string)
	if !ok {
		correlationID = ""
	}

	payload, err := json.Marshal(SendWelcomeEmailTaskPayload{
		CorrelationID: correlationID,
		UserID:        userID,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(SendWelcomeEmailTaskName, payload), nil
}

type SendWelcomeEmailTaskProcessor struct {
	userService    user.UserService
	accountService account.AccountService
}

func NewSendWelcomeEmailTaskProcessor(userService *user.UserService, accountService *account.AccountService) *SendWelcomeEmailTaskProcessor {
	return &SendWelcomeEmailTaskProcessor{
		userService:    *userService,
		accountService: *accountService,
	}
}

func (processor *SendWelcomeEmailTaskProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload SendWelcomeEmailTaskPayload
	err := json.Unmarshal(t.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	ctx = context.WithValue(ctx, "correlation_id", payload.CorrelationID)

	// TODO: maybe add a real email provider here in the future
	logger.Info(ctx, "[Dummy] send welcome email to userID: %+v", payload.UserID)
	return nil
}
