package tasks

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/skamranahmed/go-bank/config"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

type asynqTaskWorker struct {
	server     *asynq.Server
	taskRouter TaskRouter
}

func NewAsynqTaskWorker(redisConfig config.RedisConfig, queueName string) TaskWorker {
	taskRouter := NewAsynqTaskRouter()
	_, ok := taskRouter.Handler().(*asynq.ServeMux)
	if !ok {
		logger.Fatal(context.TODO(), "taskRouter for asynq worker must provide an asynq.ServeMux handler")
	}

	return &asynqTaskWorker{
		server: asynq.NewServer(
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
		taskRouter: taskRouter,
	}
}

func (w *asynqTaskWorker) Start(ctx context.Context, workerStopSignalChannel chan struct{}) {
	handler, ok := w.taskRouter.Handler().(*asynq.ServeMux)
	if !ok {
		logger.Fatal(ctx, "taskRouter for asynq worker must provide an asynq.ServeMux handler")
	}

	logger.Info(ctx, "Worker is starting")
	err := w.server.Run(handler)
	if err != nil {
		logger.Fatal(ctx, "Could not run worker server: %+v", err)
	}
	close(workerStopSignalChannel) // signals worker has been stopped
}

func (w *asynqTaskWorker) Router() TaskRouter {
	return w.taskRouter
}
