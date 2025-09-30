package tasks

import (
	"context"
	"fmt"

	"github.com/skamranahmed/go-bank/internal"
	"github.com/skamranahmed/go-bank/pkg/logger"
	tasksHelper "github.com/skamranahmed/go-bank/pkg/tasks"
)

const SendMonthlyAccountStatementOrchestratorTaskName string = "periodic_task:send_monthly_account_statement_orchestrator"

type SendMonthlyAccountStatementOrchestratorTaskPayload struct {
}

type SendMonthlyAccountStatementOrchestratorTask struct {
	name          string
	queue         string
	cronSpec      string
	maxRetryCount int
	payload       SendMonthlyAccountStatementOrchestratorTaskPayload
}

func NewSendMonthlyAccountStatementOrchestratorTask() tasksHelper.SchedulableTask {
	return &SendMonthlyAccountStatementOrchestratorTask{
		name:          SendMonthlyAccountStatementOrchestratorTaskName,
		queue:         tasksHelper.DefaultQueue,
		cronSpec:      "* * * * *", // run every minute for testing
		maxRetryCount: 0,
		payload:       SendMonthlyAccountStatementOrchestratorTaskPayload{},
	}
}

func (t *SendMonthlyAccountStatementOrchestratorTask) Name() string {
	return t.name
}

func (t *SendMonthlyAccountStatementOrchestratorTask) Queue() string {
	return t.queue
}

func (t *SendMonthlyAccountStatementOrchestratorTask) CronSpec() string {
	return t.cronSpec
}

func (t *SendMonthlyAccountStatementOrchestratorTask) MaxRetryCount() int {
	return t.maxRetryCount
}

func (t *SendMonthlyAccountStatementOrchestratorTask) Payload() any {
	return t.payload
}

type SendMonthlyAccountStatementOrchestratorTaskProcessor struct {
	services *internal.Services
}

func NewSendMonthlyAccountStatementOrchestratorTaskProcessor(services *internal.Services) tasksHelper.TaskProcessor {
	return &SendMonthlyAccountStatementOrchestratorTaskProcessor{
		services: services,
	}
}

func (processor *SendMonthlyAccountStatementOrchestratorTaskProcessor) ProcessTask(ctx context.Context, t tasksHelper.Task) error {
	taskPayloadInBytes := t.Payload().([]byte)
	payload, err := tasksHelper.ExtractPayload[SendMonthlyAccountStatementOrchestratorTaskPayload](taskPayloadInBytes)
	if err != nil {
		return fmt.Errorf("Unable to extract payload for task: %s, error: %v", t.Name(), err)
	}

	ctx = context.WithValue(ctx, "correlation_id", payload.CorrelationID)

	// TODO: add actual logic later. This is just for testing and validating my poc.
	logger.Info(ctx, "[Dummy] SendMonthlyAccountStatementOrchestratorTask ran")
	return nil
}
