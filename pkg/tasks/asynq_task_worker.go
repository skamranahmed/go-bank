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
	taskRouter *asynq.ServeMux
}

func NewAsynqTaskWorker(redisConfig config.RedisConfig, queueName string) TaskWorker {
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
		taskRouter: asynq.NewServeMux(),
	}
}

func (w *asynqTaskWorker) Start(ctx context.Context, workerStopSignalChannel chan struct{}) {
	logger.Info(ctx, "Worker is starting")
	err := w.server.Run(w.taskRouter)
	if err != nil {
		logger.Fatal(ctx, "Could not run worker server: %+v", err)
	}
	close(workerStopSignalChannel) // signals worker has been stopped
}

func (w *asynqTaskWorker) Router() *asynq.ServeMux {
	return w.taskRouter
}
