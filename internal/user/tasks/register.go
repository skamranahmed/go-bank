package tasks

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/skamranahmed/go-bank/internal"
	"github.com/skamranahmed/go-bank/pkg/logger"
	tasksHelper "github.com/skamranahmed/go-bank/pkg/tasks"
)

func RegisterTaskProcessors(backgroundTaskHandler *asynq.ServeMux, services *internal.Services) {
	backgroundTaskHandler.Handle(SendWelcomeEmailTaskName, NewSendWelcomeEmailTaskProcessor(services))
	backgroundTaskHandler.Handle(SendMonthlyAccountStatementOrchestratorTaskName, NewSendMonthlyAccountStatementOrchestratorTaskProcessor(services))
}

func RegisterScheduledTasks(backgroundTaskScheduler *asynq.Scheduler) {
	ctx := context.TODO()

	// TODO: I think this should be refactored to avoid direct dependency on asynq.Scheduler.
	// Consider creating an abstraction like "TaskScheduler" that handles registration
	// of schedulable tasks, so the concrete scheduler implementation (e.g. asynq)
	// is hidden from this loop
	for _, task := range schedulableTasks {
		asynqTask, err := tasksHelper.NewAsynqTask(ctx, task.Name(), task.Payload())
		if err != nil {
			logger.Error(ctx, "Unable to create task: %+v, error: %+v", task.Name(), err)
		}
		entryID, err := backgroundTaskScheduler.Register(
			task.CronSpec(),
			asynqTask,
		)
		if err != nil {
			logger.Error(ctx, "Scheduler unable to register task: %+v, error: %+v", task.Name(), err)
		}

		logger.Info(ctx, "Registered task: %+v with schedule: %+v, entryID: %+v", task.Name(), task.CronSpec(), entryID)
	}
}

var schedulableTasks []tasksHelper.SchedulableTask = []tasksHelper.SchedulableTask{
	NewSendMonthlyAccountStatementOrchestratorTask(),
}
