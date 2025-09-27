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
}

func RegisterScheduledTasks(backgroundTaskScheduler *asynq.Scheduler) {
	ctx := context.TODO()

	scheduledTasks, err := getScheduledTasks(ctx)
	if err != nil {
		logger.Fatal(ctx, "error: %+v", err)
	}

	for _, task := range scheduledTasks {
		entryID, err := backgroundTaskScheduler.Register(
			task.CronSpec,
			task.Task,
			asynq.Queue(task.Queue),
			asynq.MaxRetry(task.MaxRetries),
		)
		if err != nil {
			logger.Fatal(ctx, "error: %+v", err)
		}
		logger.Info(ctx, "registered an entry: %+v", entryID)
	}
}

func getScheduledTasks(ctx context.Context) ([]*tasksHelper.ScheduledTask, error) {
	var scheduledTasks []*tasksHelper.ScheduledTask

	/*
		Add scheduled tasks here
	*/

	return scheduledTasks, nil
}
