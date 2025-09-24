package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/skamranahmed/go-bank/config"
	"github.com/skamranahmed/go-bank/internal"
	userTasks "github.com/skamranahmed/go-bank/internal/user/tasks"
	"github.com/skamranahmed/go-bank/pkg/logger"
)

const (
	DefaultQueue  string = "default"
	PriorityQueue string = "priority"
)

func Start(queueName string, services *internal.Services) {
	redisConfig := config.GetRedisConfig()
	worker := newWorker(queueName, redisConfig)

	backgroundTaskRouter := asynq.NewServeMux()
	registerTaskHandlers(backgroundTaskRouter, services)

	err := worker.Run(backgroundTaskRouter)
	if err != nil {
		logger.Fatal(context.TODO(), "Could not run worker server: %+v", err)
	}
}

type worker struct {
	*asynq.Server
}

func newWorker(queueName string, redisConfig config.RedisConfig) *worker {
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
	}
}

func registerTaskHandlers(backgroundTaskRouter *asynq.ServeMux, services *internal.Services) {
	// keep this in alphabetical order for easy scanning by eyes

	// user tasks
	registerUserTasks(backgroundTaskRouter, services)
}

func registerUserTasks(backgroundTaskRouter *asynq.ServeMux, services *internal.Services) {
	// send-welcome-email task
	backgroundTaskRouter.Handle(
		userTasks.SendWelcomeEmailTaskName,
		userTasks.NewSendWelcomeEmailTaskProcessor(&services.UserService, &services.AccountService),
	)
}
