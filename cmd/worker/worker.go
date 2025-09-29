package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
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
	workerDone := make(chan struct{}) // workerDone channel signals when worker.start() exits
	worker := newWorker(queueName, redisConfig, services)
	go worker.start(ctx, workerDone)

	/*
		scheduler setup
	*/
	schedulerStopSignalChannel := make(chan struct{}) // schedulerShutdownSignalChannel signals when scheduler stops
	taskScheduler := tasksHelper.NewAsynqTaskScheduler(redisConfig)
	RegisterSchedulableTasks(taskScheduler)
	go taskScheduler.Start(ctx, schedulerStopSignalChannel)

	<-workerDone // blocks until the worker stops
	logger.Info(ctx, "Worker stopped")

	<-schedulerStopSignalChannel // blocks until the scheduler stops
	logger.Info(ctx, "Scheduler stopped")
}

type worker struct {
	*asynq.Server
	TaskHandler *asynq.ServeMux
	Services    *internal.Services
}

func newWorker(queueName string, redisConfig config.RedisConfig, services *internal.Services) *worker {
	return &worker{
		Server: asynq.NewServer(
			asynq.RedisClientOpt{
				Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
				Password: redisConfig.Password,
				DB:       redisConfig.DbIndex,
			},
			asynq.Config{
				Concurrency: 1,
				Queues: map[string]int{
					queueName: 1,
				},
			},
		),
		TaskHandler: asynq.NewServeMux(),
		Services:    services,
	}
}

/*
start sets up task processors and runs the worker.
It blocks until an os signal to exit the program is received.
Once it receives a signal, it gracefully shuts down all active workers and other goroutines to process the tasks.
*/
func (w *worker) start(ctx context.Context, workerDone chan struct{}) {
	w.registerTaskProcessors()
	logger.Info(ctx, "Worker is starting")
	err := w.Run(w.TaskHandler)
	if err != nil {
		logger.Fatal(ctx, "Could not run worker server: %+v", err)
	}
	close(workerDone) // signal completion
}

func (w *worker) registerTaskProcessors() {
	// user tasks
	userTasks.RegisterTaskProcessors(w.TaskHandler, w.Services)
}

func RegisterSchedulableTasks(taskScheduler tasksHelper.TaskScheduler) {
	// user tasks
	userTasks.RegisterSchedulableTasks(taskScheduler)
}
