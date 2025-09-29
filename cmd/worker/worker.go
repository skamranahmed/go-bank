package worker

import (
	"context"

	"github.com/skamranahmed/go-bank/config"
	"github.com/skamranahmed/go-bank/internal"
	userTasks "github.com/skamranahmed/go-bank/internal/user/tasks"
	"github.com/skamranahmed/go-bank/pkg/logger"
	tasksHelper "github.com/skamranahmed/go-bank/pkg/tasks"
)

func Start(queueName string, services *internal.Services) {
	ctx := context.TODO()
	redisConfig := config.GetRedisConfig()

	/*
		worker setup
	*/
	workerStopSignalChannel := make(chan struct{}) // workerStopSignalChannel signals when worker stops
	taskWorker := tasksHelper.NewAsynqTaskWorker(redisConfig, queueName)
	RegisterTaskProcessors(taskWorker, services)
	go taskWorker.Start(ctx, workerStopSignalChannel)

	/*
		scheduler setup
	*/
	schedulerStopSignalChannel := make(chan struct{}) // schedulerShutdownSignalChannel signals when scheduler stops
	taskScheduler := tasksHelper.NewAsynqTaskScheduler(redisConfig)
	RegisterSchedulableTasks(taskScheduler)
	go taskScheduler.Start(ctx, schedulerStopSignalChannel)

	<-workerStopSignalChannel // blocks until the worker stops
	logger.Info(ctx, "Worker stopped")

	<-schedulerStopSignalChannel // blocks until the scheduler stops
	logger.Info(ctx, "Scheduler stopped")
}

func RegisterSchedulableTasks(taskScheduler tasksHelper.TaskScheduler) {
	// user tasks
	userTasks.RegisterSchedulableTasks(taskScheduler)
}

func RegisterTaskProcessors(taskWorker tasksHelper.TaskWorker, services *internal.Services) {
	// user tasks
	userTasks.RegisterTaskProcessors(taskWorker.Router(), services)
}
