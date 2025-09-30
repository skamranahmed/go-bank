package tasks

import (
	"context"

	"github.com/skamranahmed/go-bank/internal"
	"github.com/skamranahmed/go-bank/pkg/logger"
	tasksHelper "github.com/skamranahmed/go-bank/pkg/tasks"
)

func RegisterTaskProcessors(taskRouter tasksHelper.TaskRouter, services *internal.Services) {
	taskRouter.RegisterTaskProcessor(SendWelcomeEmailTaskName, NewSendWelcomeEmailTaskProcessor(services))
	taskRouter.RegisterTaskProcessor(SendMonthlyAccountStatementOrchestratorTaskName, NewSendMonthlyAccountStatementOrchestratorTaskProcessor(services))
}

func RegisterSchedulableTasks(taskScheduler tasksHelper.TaskScheduler) {
	ctx := context.TODO()
	for _, schedulableTask := range schedulableTasks {
		entryID, err := taskScheduler.RegisterTask(ctx, schedulableTask)
		if err != nil {
			logger.Error(ctx, "Scheduler was unable to register task: %+v, error: %+v", schedulableTask.Name(), err)
			continue
		}
		logger.Info(ctx, "Registered scheduled task: %+v with schedule: %+v, entryID: %+v", schedulableTask.Name(), schedulableTask.CronSpec(), entryID)
	}
}

var schedulableTasks []tasksHelper.SchedulableTask = []tasksHelper.SchedulableTask{
	NewSendMonthlyAccountStatementOrchestratorTask(),
}
